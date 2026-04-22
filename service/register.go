package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	etcdv2 "go.etcd.io/etcd/client/v2"
	"gopkg.in/errgo.v1"
)

const (
	// HEARTBEAT_DURATION time in second between two registrations. The host will
	// be deleted if etcd didn't receive any new registration in those 5 seconds.
	HEARTBEAT_DURATION = 5
)

// Register a host with a service name and a host description. The last chan is
// a stop method. If something is written on this channel, any goroutines
// launched by this method will stop.
//
// This service will launch two go routines. The first one will maintain the
// registration every 5 seconds, and the second one will check if the service
// credentials don't change and notify otherwise.
func Register(ctx context.Context, service string, host Host) *Registration {
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

	go func() {
		ticker := time.NewTicker((HEARTBEAT_DURATION - 1) * time.Second)
		defer ticker.Stop()

		// id is the current modification index of the service key.
		// this is used for the watcher.
		id, err := serviceRegistration(ctx, serviceKey, serviceValue)
		for err != nil {
			if ctx.Err() != nil {
				return
			}

			id, err = serviceRegistration(ctx, serviceKey, serviceValue)
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
			go watch(ctx, serviceKey, id, privateCredentialsChan)
		}

		for {
			select {
			case <-ctx.Done():
				_, err := KAPI().Delete(ctx, hostKey, &etcdv2.DeleteOptions{Recursive: false})
				if err != nil {
					logger.Println("remove host key", hostKey)
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

	return NewRegistration(ctx, hostUUID, publicCredentialsChan)
}

func watch(ctx context.Context, serviceKey string, id uint64, credentialsChan chan Credentials) {
	// id is the index of the last modification made to the key. The watcher will
	// start watching for modifications done after this index. This will prevent
	// packet or modification lost.

	for {
		watcher := KAPI().Watcher(serviceKey, &etcdv2.WatcherOptions{
			AfterIndex: id,
		})
		resp, err := watcher.Next(ctx)
		if err == context.Canceled {
			return
		}

		if err != nil {
			// We've lost the connexion to etcd. Sleep 1s and retry
			logger.Printf("Lost watcher of '%v': '%v' (%v)", serviceKey, err, Client().Endpoints())
			id = 0
			time.Sleep(1 * time.Second)
			continue
		}
		var serviceInfos Service
		err = json.Unmarshal([]byte(resp.Node.Value), &serviceInfos)
		if err != nil {
			logger.Printf(
				"Error while getting service key '%v': '%v' (%v)",
				serviceKey, err, Client().Endpoints(),
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

func hostRegistration(ctx context.Context, hostKey, hostJSON string) error {
	_, err := KAPI().Set(ctx, hostKey, hostJSON, &etcdv2.SetOptions{TTL: HEARTBEAT_DURATION * time.Second})
	if err != nil {
		return errgo.Notef(err, "Unable to register host")
	}
	return nil
}

// ensureHostRegistration keeps retrying the host registration until it succeeds or the context is canceled.
func ensureHostRegistration(ctx context.Context, service, hostKey, hostJSON string, logFailures bool) error {
	err := hostRegistration(ctx, hostKey, hostJSON)
	for err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if logFailures {
			logger.Printf("Lost registration of '%v': %v (%v)", service, err, Client().Endpoints())
		}

		// Wait for either context cancellation or the next retry attempt.
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(1 * time.Second):
		}

		err = hostRegistration(ctx, hostKey, hostJSON)
		if err == nil && logFailures {
			logger.Printf("Recover registration of '%v'", service)
		}
	}

	return nil
}

func serviceRegistration(ctx context.Context, serviceKey, serviceJSON string) (uint64, error) {
	key, err := KAPI().Set(ctx, serviceKey, serviceJSON, nil)
	if err != nil {
		return 0, errgo.Notef(err, "Unable to register service")
	}

	return key.Node.ModifiedIndex, nil
}
