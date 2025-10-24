package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"gopkg.in/errgo.v1"
)

// RegisterV3 a host with a service name and a host description. The last chan is
// a stop method. If something is written on this channel, any goroutines
// launch by this method will stop.
//
// This service will launch two go routines. The first one will maintain the
// registration every 5 seconds and the second one will check if the service
// credentials don't change and notify otherwise
func RegisterV3(ctx context.Context, service string, host Host) *Registration {
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
	hostUuid := fmt.Sprintf("%s-%s", uuidV4.String(), host.PrivateHostname)
	host.UUID = hostUuid

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

	hostKey := fmt.Sprintf("/services/%s/%s", service, hostUuid)
	hostJson, _ := json.Marshal(&host)
	hostValue := string(hostJson)

	serviceKey := fmt.Sprintf("/services_infos/%s", service)
	serviceJson, _ := json.Marshal(serviceInfos)
	serviceValue := string(serviceJson)

	go func() {
		ticker := time.NewTicker((HEARTBEAT_DURATION - 1) * time.Second)

		// rev is the current modification revision of the service key.
		// this is used for the watcher.
		rev, err := serviceRegistrationV3(serviceKey, serviceValue)
		for err != nil {
			rev, err = serviceRegistrationV3(serviceKey, serviceValue)
		}

		err = hostRegistrationV3(hostKey, hostValue)
		for err != nil {
			err = hostRegistrationV3(hostKey, hostValue)
		}

		publicCredentialsChan <- Credentials{
			User:     serviceInfos.User,
			Password: serviceInfos.Password,
		}

		if host.Public {
			go watchv3(ctx, serviceKey, rev, privateCredentialsChan)
		}

		for {
			select {
			case <-ctx.Done():
				_, err := KAPIV3().Delete(context.Background(), hostKey)
				if err != nil {
					logger.Println("fail to remove key", hostKey)
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
				hostJson, _ = json.Marshal(&host)
				hostValue = string(hostJson)

				// synchro the host information
				_ = hostRegistrationV3(hostKey, hostValue)
				// and transmit them to the client
				publicCredentialsChan <- credentials
			case <-ticker.C:
				err := hostRegistrationV3(hostKey, hostValue)
				// If for any random reason, there is an error,
				// we retry every second until it's ok.
				for err != nil {
					logger.Printf("lost registration of '%v': %v (%v)", service, err, Client().Endpoints())
					time.Sleep(1 * time.Second)

					err = hostRegistrationV3(hostKey, hostValue)
					if err == nil {
						logger.Printf("recover registration of '%v'", service)
					}
				}
			}
		}
	}()

	return NewRegistration(ctx, hostUuid, publicCredentialsChan)
}

func watchv3(ctx context.Context, serviceKey string, rev int64, credentialsChan chan Credentials) {
	// rev is the revision of the last modification made to the key. The watcher will
	// start watching for modifications done after this revision. This will prevent
	// packet or modification lost.

	// We always watch from rev+1 (i.e., strictly after the known write)
	startRev := rev + 1

	for {
		rch := WatchV3().Watch(ctx, serviceKey, etcdv3.WithRev(startRev))
		for resp := range rch {
			if resp.Canceled {
				// We've lost the connection to etcd or the context was canceled.
				if errors.Is(resp.Err(), context.Canceled) || errors.Is(ctx.Err(), context.Canceled) {
					return
				}

				logger.Printf("lost watcher of '%v': '%v' (%v)", serviceKey, resp.Err(), Client().Endpoints())
				time.Sleep(1 * time.Second)
				// Restart from last known startRev
				break
			}

			for _, ev := range resp.Events {
				// We only care about PUT updates on the single key
				if ev.Kv == nil || ev.Type != etcdv3.EventTypePut {
					continue
				}
				var serviceInfos Service
				if err := json.Unmarshal(ev.Kv.Value, &serviceInfos); err != nil {
					logger.Printf("error while getting service key '%v': '%v' (%v)", serviceKey, err, Client().Endpoints())
					time.Sleep(1 * time.Second)
					continue
				}
				// We've got the modification, send it to the register agent
				startRev = ev.Kv.ModRevision + 1
				credentialsChan <- Credentials{
					User:     serviceInfos.User,
					Password: serviceInfos.Password,
				}
			}
		}

		// Channel closed not due to context cancel: short backoff and retry.
		if errors.Is(ctx.Err(), context.Canceled) {
			return
		}
		time.Sleep(1 * time.Second)
	}
}

// hostRegistrationV3 writes the host key with a TTL using a fresh short-lived lease.
// This mirrors the v2 "Set with TTL" behavior and keeps the rest of the code unchanged.
func hostRegistrationV3(hostKey, hostJson string) error {
	// Create a lease with the desired TTL
	lease, err := LeaseV3().Grant(context.Background(), int64(HEARTBEAT_DURATION))
	if err != nil {
		return errgo.Notef(err, "Unable to create lease for host registration")
	}

	// Put the value attached to the lease
	_, err = KAPIV3().Put(context.Background(), hostKey, hostJson, etcdv3.WithLease(lease.ID))
	if err != nil {
		return errgo.Notef(err, "Unable to register host")
	}
	return nil
}

// serviceRegistrationV3 writes/updates the service info and returns the store revision of that write.
// Callers can start a watch from (rev+1) to get subsequent updates.
func serviceRegistrationV3(serviceKey, serviceJson string) (int64, error) {
	resp, err := KAPIV3().Put(context.Background(), serviceKey, serviceJson)
	if err != nil {
		return 0, errgo.Notef(err, "Unable to register service")
	}
	return resp.Header.GetRevision(), nil
}
