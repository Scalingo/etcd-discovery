package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	"gopkg.in/errgo.v1"

	"github.com/Scalingo/etcd-discovery/v8/service/etcdwrapper"
)

// Register a host with a service name and a host description. The last chan is
// a stop method. If something is written on this channel, any goroutines
// launch by this method will stop.
//
// This service will launch two go routines. The first one will maintain the
// registration every 5 seconds and the second one will check if the service
// credentials don't change and notify otherwise
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
		ticker := time.NewTicker((etcdwrapper.HEARTBEAT_DURATION - 1) * time.Second)

		// id is the current modification index of the service key.
		// this is used for the watcher.
		idV2, idV3, err := serviceRegistration(serviceKey, serviceValue)
		for err != nil {
			idV2, idV3, err = serviceRegistration(serviceKey, serviceValue)
		}

		err = hostRegistration(hostKey, hostValue)
		for err != nil {
			err = hostRegistration(hostKey, hostValue)
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
				err := etcdwrapper.Delete(context.Background(), hostKey)
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
				hostRegistration(hostKey, hostValue)
				// and transmit them to the client
				publicCredentialsChan <- credentials
			case <-ticker.C:
				err := hostRegistration(hostKey, hostValue)
				// If for any random reason, there is an error,
				// we retry every second until it's ok.
				for err != nil {
					endpointsV2, endpointsV3 := etcdwrapper.Endpoints()
					logger.Printf("lost registration of '%v': %v (v2: %v) (v3: %v)", service, err, endpointsV2, endpointsV3)
					time.Sleep(1 * time.Second)

					err = hostRegistration(hostKey, hostValue)
					if err == nil {
						logger.Printf("recover registration of '%v'", service)
					}
				}
			}
		}
	}()

	return NewRegistration(ctx, hostUuid, publicCredentialsChan)
}

func watch(ctx context.Context, serviceKey string, idV2 uint64, idV3 int64, credentialsChan chan Credentials) {
	for {
		watchChan := etcdwrapper.Watch(logger, ctx, serviceKey, idV2, idV3)
		select {
		case watchRes := <-watchChan:
			credentialsChan <- Credentials{
				User:     watchRes.User,
				Password: watchRes.Password,
			}
		}
	}
}

func hostRegistration(hostKey, hostJson string) error {
	_, _, err := etcdwrapper.Set(context.Background(), hostKey, hostJson, true)
	if err != nil {
		return errgo.Notef(err, "Unable to register host")
	}
	return nil
}

func serviceRegistration(serviceKey, serviceJson string) (uint64, int64, error) {
	idxV2, idxV3, err := etcdwrapper.Set(context.Background(), serviceKey, serviceJson, false)
	if err != nil {
		return 0, 0, errgo.Notef(err, "Unable to register service")
	}

	return idxV2, idxV3, nil
}
