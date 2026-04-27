package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	etcdv2 "go.etcd.io/etcd/client/v2"

	"github.com/Scalingo/go-utils/errors/v3"
	"github.com/Scalingo/go-utils/logger"
)

const (
	// heartbeatTTL is the etcd TTL for a host registration. The registration is
	// refreshed before this expires, so the host stays available while healthy.
	heartbeatTTL = 5 * time.Second

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

	if !host.Public && len(host.PrivateHostname) == 0 {
		host.PrivateHostname = host.Hostname
	}

	if len(host.PrivateHostname) == 0 {
		host.PrivateHostname = hostname
	}
	host.Name = service

	if len(host.PrivateHostname) != 0 && len(host.PrivatePorts) == 0 {
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
		ticker := time.NewTicker(heartbeatTTL - time.Second)
		defer ticker.Stop()

		// id is the current modification index of the service key.
		// this is used for the watcher.
		id, err := ensureServiceRegistration(ctx, serviceKey, serviceValue)
		if err != nil {
			registration.signalFailure(err)
			return
		}

		err = ensureHostRegistration(ctx, service, hostKey, hostValue, false)
		if err != nil {
			registration.signalFailure(err)
			return
		}

		publicCredentialsChan <- Credentials{
			User:     serviceInfos.User,
			Password: serviceInfos.Password,
		}

		if host.Public {
			go watch(ctx, serviceKey, id, privateCredentialsChan)
		}

		for {
			select {
			case <-ctx.Done():
				_, err := KAPI().Delete(ctx, hostKey, &etcdv2.DeleteOptions{Recursive: false})
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

func watch(ctx context.Context, serviceKey string, id uint64, credentialsChan chan Credentials) {
	log := logger.Get(ctx)

	// id is the index of the last modification made to the key. The watcher will
	// start watching for modifications done after this index. This will prevent
	// packet or modification lost.
	for {
		watcher := KAPI().Watcher(serviceKey, &etcdv2.WatcherOptions{
			AfterIndex: id,
		})
		resp, err := watcher.Next(ctx)
		if errors.Is(err, context.Canceled) {
			return
		}

		if err != nil {
			// We've lost the connexion to etcd. Sleep 1s and retry
			log.WithError(err).Errorf("Lost watcher of '%s' (%v)", serviceKey, Client().Endpoints())
			id = 0
			time.Sleep(1 * time.Second)
			continue
		}
		var serviceInfos Service
		err = json.Unmarshal([]byte(resp.Node.Value), &serviceInfos)
		if err != nil {
			log.WithError(err).Errorf(
				"Error while getting service key '%s' (%v)",
				serviceKey, Client().Endpoints(),
			)
			time.Sleep(1 * time.Second)
		}
		// We've got the modification, send it to the register agent
		id = resp.Node.ModifiedIndex
		credentialsChan <- Credentials{
			User:     serviceInfos.User,
			Password: serviceInfos.Password,
		}
	}
}

func ensureServiceRegistration(ctx context.Context, serviceKey, serviceJSON string) (uint64, error) {
	ctx, cancel := withDefaultRegistrationTimeout(ctx)
	defer cancel()

	id, err := serviceRegistration(ctx, serviceKey, serviceJSON)
	for err != nil {
		if ctx.Err() != nil {
			return 0, ctx.Err()
		}

		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-time.After(1 * time.Second):
		}

		id, err = serviceRegistration(ctx, serviceKey, serviceJSON)
	}

	return id, nil
}

func hostRegistration(ctx context.Context, hostKey, hostJSON string) error {
	_, err := KAPI().Set(ctx, hostKey, hostJSON, &etcdv2.SetOptions{TTL: heartbeatTTL})
	if err != nil {
		return errors.Wrap(ctx, err, "register host")
	}
	return nil
}

// ensureHostRegistration keeps retrying the host registration until it succeeds or the context is canceled.
func ensureHostRegistration(ctx context.Context, service, hostKey, hostJSON string, logFailures bool) error {
	ctx, cancel := withDefaultRegistrationTimeout(ctx)
	defer cancel()

	log := logger.Get(ctx)

	err := hostRegistration(ctx, hostKey, hostJSON)
	for err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if logFailures {
			log.WithError(err).Errorf("Lost registration of '%s' (%v)", service, Client().Endpoints())
		}

		// Wait for either context cancellation or the next retry attempt.
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(1 * time.Second):
		}

		registrationErr := hostRegistration(ctx, hostKey, hostJSON)
		if registrationErr == nil {
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

func serviceRegistration(ctx context.Context, serviceKey, serviceJSON string) (uint64, error) {
	key, err := KAPI().Set(ctx, serviceKey, serviceJSON, nil)
	if err != nil {
		return 0, errors.Wrap(ctx, err, "register service")
	}

	return key.Node.ModifiedIndex, nil
}
