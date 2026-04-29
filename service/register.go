package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"

	"github.com/Scalingo/etcd-discovery/v9/service/etcdwrapper"
	"github.com/Scalingo/go-utils/errors/v3"
	"github.com/Scalingo/go-utils/logger"
)

const (
	// defaultRegistrationTimeout is used when the caller does not provide a deadline on context.
	defaultRegistrationTimeout = 5 * time.Minute
)

// Register a host with a service name and a host description. The last chan is
// a stop method. If something is written on this channel, any goroutines
// launched by this method will stop.
//
// This service will launch two go routines. The first one will maintain the
// registration every 5 seconds, and the second one will check if the service
// credentials don't change and notify otherwise.
func Register(ctx context.Context, service string, host Host) *Registration {
	log := logger.Get(ctx)
	if !host.Public && host.PrivateHostname == "" {
		host.PrivateHostname = host.Hostname
	}

	if host.PrivateHostname == "" {
		host.PrivateHostname = hostname
	}
	host.Name = service

	if host.PrivateHostname != "" && len(host.PrivatePorts) == 0 {
		host.PrivatePorts = host.Ports
	}

	uuidV4, _ := uuid.NewV4()
	hostUUID := fmt.Sprintf("%s-%s", uuidV4.String(), host.PrivateHostname)
	host.UUID = hostUUID

	serviceInfos := &Service{
		Name:     service,
		Critical: host.Critical,
		Public:   host.Public,
	}

	if host.Public {
		serviceInfos.Hostname = host.Hostname
		serviceInfos.Ports = host.Ports
		serviceInfos.Password = host.Password
		serviceInfos.User = host.User
	}

	publicCredentialsChan := make(chan Credentials, 1)  // Communication between register and the client
	privateCredentialsChan := make(chan Credentials, 1) // Communication between watcher and register

	hostKey := fmt.Sprintf("/services/%s/%s", service, hostUUID)
	hostJSON, _ := json.Marshal(&host)
	hostValue := string(hostJSON)

	serviceKey := fmt.Sprintf("/services_infos/%s", service)
	serviceJSON, _ := json.Marshal(serviceInfos)
	serviceValue := string(serviceJSON)

	registration := NewRegistration(ctx, hostUUID, publicCredentialsChan)

	go func() {
		ticker := time.NewTicker(etcdwrapper.HeartbeatDuration - time.Second)
		defer ticker.Stop()

		// id is the current modification index of the service key.
		// this is used for the watcher.
		idv2, idv3, err := ensureServiceRegistration(ctx, serviceKey, serviceValue)
		if err != nil {
			registration.signalFailure(err)
			return
		}

		err = ensureInitialHostRegistration(ctx, service, hostKey, hostValue, false)
		if err != nil {
			registration.signalFailure(err)
			return
		}

		publicCredentialsChan <- Credentials{
			User:     serviceInfos.User,
			Password: serviceInfos.Password,
		}

		if host.Public {
			go watch(ctx, serviceKey, idv2, idv3, privateCredentialsChan)
		}

		for {
			select {
			case <-ctx.Done():
				err := etcdwrapper.Delete(ctx, hostKey)
				if err != nil {
					log.WithError(err).Errorf("remove host key %s", hostKey)
				}
				return
			case credentials := <-privateCredentialsChan: // If the credentials have been changed,
				// We update our cache
				host.User = credentials.User
				host.Password = credentials.Password
				serviceInfos.User = credentials.User
				serviceInfos.Password = credentials.Password

				// Re-marshal the host
				hostJSON, _ = json.Marshal(&host)
				hostValue = string(hostJSON)

				// Sync the host information
				err := ensureHostRegistration(ctx, service, hostKey, hostValue, true)
				if err != nil {
					return
				}
				// and transmit them to the client
				publicCredentialsChan <- credentials
			case <-ticker.C:
				err := ensureHostRegistration(ctx, service, hostKey, hostValue, true)
				if err != nil {
					return
				}
			}
		}
	}()

	return registration
}

func watch(ctx context.Context, serviceKey string, idV2 uint64, idV3 int64, credentialsChan chan Credentials) {
	for {
		watchChan := etcdwrapper.Watch(ctx, serviceKey, idV2, idV3)
		watchRes := <-watchChan
		credentialsChan <- Credentials{
			User:     watchRes.User,
			Password: watchRes.Password,
		}
	}
}

func ensureServiceRegistration(ctx context.Context, serviceKey, serviceJSON string) (uint64, int64, error) {
	ctx, cancel := withDefaultRegistrationTimeout(ctx)
	defer cancel()

	idxV2, idxV3, err := serviceRegistration(ctx, serviceKey, serviceJSON)
	for err != nil {
		if ctx.Err() != nil {
			return 0, 0, ctx.Err()
		}

		select {
		case <-ctx.Done():
			return 0, 0, ctx.Err()
		case <-time.After(1 * time.Second):
		}

		idxV2, idxV3, err = serviceRegistration(ctx, serviceKey, serviceJSON)
	}

	return idxV2, idxV3, nil
}

func hostRegistration(ctx context.Context, hostKey, hostJSON string) error {
	_, _, err := etcdwrapper.Set(ctx, hostKey, hostJSON, true)
	if err != nil {
		return errors.Wrap(ctx, err, "register host")
	}
	return nil
}

func ensureInitialHostRegistration(ctx context.Context, service, hostKey, hostJSON string, logFailures bool) error {
	registrationCtx, cancel := withDefaultRegistrationTimeout(ctx)
	defer cancel()

	return ensureHostRegistration(registrationCtx, service, hostKey, hostJSON, logFailures)
}

// ensureHostRegistration keeps retrying the host registration until it succeeds or the context is canceled.
func ensureHostRegistration(ctx context.Context, service, hostKey, hostJSON string, logFailures bool) error {
	log := logger.Get(ctx)

	err := hostRegistration(ctx, hostKey, hostJSON)
	for err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if logFailures {
			endpointsV2, endpointsV3 := etcdwrapper.Endpoints()
			log.WithError(err).Errorf("lost registration of '%v': %v (v2: %v) (v3: %v)", service, err, endpointsV2, endpointsV3)
		}

		// Wait for either context cancellation or the next retry attempt.
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(1 * time.Second):
		}

		err = hostRegistration(ctx, hostKey, hostJSON)
		if err == nil && logFailures {
			log.Infof("Recover registration of '%s'", service)
		}
	}

	return nil
}

func withDefaultRegistrationTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	_, hasDeadline := ctx.Deadline()
	if hasDeadline {
		return ctx, func() {}
	}

	return context.WithTimeout(ctx, defaultRegistrationTimeout)
}

func serviceRegistration(ctx context.Context, serviceKey, serviceJSON string) (uint64, int64, error) {
	idxV2, idxV3, err := etcdwrapper.Set(ctx, serviceKey, serviceJSON, false)
	if err != nil {
		return 0, 0, errors.Wrap(ctx, err, "register service")
	}

	return idxV2, idxV3, nil
}
