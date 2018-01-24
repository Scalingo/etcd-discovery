package service

import (
	"encoding/json"
	"fmt"
	"time"

	errgo "gopkg.in/errgo.v1"

	etcd "github.com/coreos/etcd/client"
	uuid "github.com/nu7hatch/gouuid"
	"golang.org/x/net/context"
)

const (
	// HEARTBEAT_DURATION time in second between two registration. The host will
	// be deleted if etcd didn't received any new registration in those 5 seocnds
	HEARTBEAT_DURATION = 5
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

	uuid, _ := uuid.NewV4()

	hostUuid := fmt.Sprintf("%s-%s", uuid.String(), host.PrivateHostname)
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

		// id is the current modification index of the service key.
		// this is used for the watcher.
		id, err := serviceRegistration(serviceKey, serviceValue)
		for err != nil {
			id, err = serviceRegistration(serviceKey, serviceValue)
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
			go watch(ctx, serviceKey, id, privateCredentialsChan)
		}

		for {
			select {
			case <-ctx.Done():
				_, err := KAPI().Delete(context.Background(), hostKey, &etcd.DeleteOptions{Recursive: false})
				if err != nil {
					logger.Println("fail to remove key", hostKey)
				}
				ticker.Stop()
				return
			case credentials := <-privateCredentialsChan: // If the credentials has benn changed
				// We update our cache
				host.User = credentials.User
				host.Password = credentials.Password
				serviceInfos.User = credentials.User
				serviceInfos.Password = credentials.Password

				// Re-marshal the host
				hostJson, _ = json.Marshal(&host)
				hostValue = string(hostJson)

				// synchro the host informations
				hostRegistration(hostKey, hostValue)
				// and transmit them to the client
				publicCredentialsChan <- credentials
			case <-ticker.C:
				err := hostRegistration(hostKey, hostValue)
				// If for any random reason, there is an error,
				// we retry every second until it's ok.
				for err != nil {
					logger.Printf("lost registration of '%v': %v (%v)", service, err, Client().Endpoints())
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

func watch(ctx context.Context, serviceKey string, id uint64, credentialsChan chan Credentials) {
	// id is the index of the last modification made to the key. The watcher will
	// start watching for modifications done after this index. This will prevent
	// packet or modification lost.

	for {
		watcher := KAPI().Watcher(serviceKey, &etcd.WatcherOptions{
			AfterIndex: id,
		})
		resp, err := watcher.Next(ctx)
		if err == context.Canceled {
			return
		}

		if err != nil {
			// We've lost the connexion to etcd. Speel 1s and retry
			logger.Printf("lost watcher of '%v': '%v' (%v)", serviceKey, err, Client().Endpoints())
			id = 0
			time.Sleep(1 * time.Second)
		}
		var serviceInfos Service
		err = json.Unmarshal([]byte(resp.Node.Value), &serviceInfos)
		if err != nil {
			logger.Printf("error while getting service key '%v': '%v' (%v)", serviceKey, err, Client().Endpoints())
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

func hostRegistration(hostKey, hostJson string) error {
	_, err := KAPI().Set(context.Background(), hostKey, hostJson, &etcd.SetOptions{TTL: HEARTBEAT_DURATION * time.Second})
	if err != nil {
		return errgo.Notef(err, "Unable to register host")
	}
	return nil

}

func serviceRegistration(serviceKey, serviceJson string) (uint64, error) {
	key, err := KAPI().Set(context.Background(), serviceKey, serviceJson, nil)
	if err != nil {
		return 0, errgo.Notef(err, "Unable to register service")
	}

	return key.Node.ModifiedIndex, nil
}
