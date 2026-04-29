package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Scalingo/etcd-discovery/v9/service/etcdwrapper"

	"github.com/Scalingo/go-utils/errors/v3"
	"github.com/Scalingo/go-utils/logger"

	"github.com/gofrs/uuid/v5"
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

	hostKey := fmt.Sprintf("/services/%s/%s", service, hostUUID) //nolint: perfsprint
	hostJSON, _ := json.Marshal(&host)
	hostValue := string(hostJSON)

	serviceKey := fmt.Sprintf("/services_infos/%s", service) //nolint: perfsprint
	serviceJSON, _ := json.Marshal(serviceInfos)
	serviceValue := string(serviceJSON)

	go func() {
		ticker := time.NewTicker((etcdwrapper.HeartbeatDuration - 1) * time.Second)
		defer ticker.Stop()

		// id is the current modification index of the service key.
		// this is used for the watcher.
		idV2, idV3, err := serviceRegistration(ctx, serviceKey, serviceValue)
		for err != nil {
			if ctx.Err() != nil {
				return
			}

			idV2, idV3, err = serviceRegistration(ctx, serviceKey, serviceValue)
		}

		err = ensureHostRegistration(ctx, service, hostKey, hostValue, false)
		if err != nil {
			return
		}

		publicCredentialsChan <- Credentials{
			User:     serviceInfos.User,
			Password: serviceInfos.Password,
		}

		if host.Public {
			go watch(ctx, serviceKey, idV2, idV3, privateCredentialsChan)
		}

		for {
			select {
			case <-ctx.Done():
				err := etcdwrapper.Delete(ctx, hostKey)
				if err != nil {
					log.WithError(err).Errorf("remove host key %s", hostKey)
				}
				ticker.Stop()
				return
			case credentials := <-privateCredentialsChan: // If the credentials have been changed
				// We update our cache
				host.User = credentials.User
				host.Password = credentials.Password
				serviceInfos.User = credentials.User
				serviceInfos.Password = credentials.Password

				// Re-marshal the host
				hostJSON, _ = json.Marshal(&host)
				hostValue = string(hostJSON)

				// synchro the host information
				err := ensureHostRegistration(ctx, service, hostKey, hostValue, true)
				if err != nil {
					return
				}
				// and transmit them to the client
				publicCredentialsChan <- credentials
			case <-ticker.C:
			case <-ticker.C:
				err := ensureHostRegistration(ctx, service, hostKey, hostValue, true)
				if err != nil {
					return
				}
			}
		}
	}()

	return NewRegistration(ctx, hostUUID, publicCredentialsChan)
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

func hostRegistration(ctx context.Context, hostKey, hostJSON string) error {
	_, _, err := etcdwrapper.Set(ctx, hostKey, hostJSON, true)
	if err != nil {
		return errors.Wrapf(ctx, err, "unable to register host")
	}
	return nil
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
			log.Errorf("Recover registration of '%s'", service)
		}
	}

	return nil
}

func serviceRegistration(ctx context.Context, serviceKey, serviceJSON string) (uint64, int64, error) {
	idxV2, idxV3, err := etcdwrapper.Set(ctx, serviceKey, serviceJSON, false)
	if err != nil {
		return 0, 0, errors.Wrap(ctx, err, "register service")
	}

	return idxV2, idxV3, nil
}
