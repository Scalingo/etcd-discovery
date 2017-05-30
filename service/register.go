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
	HEARTBEAT_DURATION = 5
)

func Register(service string, host *Host, stop chan struct{}) (string, chan Credentials) {
	uuid, _ := uuid.NewV4()

	hostUuid := uuid.String()
	host.Uuid = hostUuid

	if len(host.PrivateHostname) == 0 {
		host.PrivateHostname = hostname
	}

	host.Name = service

	serviceInfos := &Service{
		Name:     service,
		Critical: host.Critical,
		User:     host.User,
		Password: host.Password,
		Public:   host.Public,
	}

	if host.Public {
		serviceInfos.Hostname = host.Hostname
		serviceInfos.Ports = host.Ports
	}

	publicCredentialsChan := make(chan Credentials, 1)  // Communication between register and the client
	privateCredentialsChan := make(chan Credentials, 1) // Communication between watcher and register

	watcherStopper := make(chan struct{})

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
			go watch(serviceKey, id, privateCredentialsChan, watcherStopper)
		}

		for {
			select {
			case <-stop:
				_, err := KAPI().Delete(context.Background(), hostKey, &etcd.DeleteOptions{Recursive: false})
				if err != nil {
					logger.Println("fail to remove key", hostKey)
				}
				ticker.Stop()
				watcherStopper <- struct{}{}
				return
			case credentials := <-privateCredentialsChan: // If the credentials has benn changed
				// We update our cache
				host.User = credentials.User
				host.Password = credentials.Password
				serviceInfos.User = credentials.User
				serviceInfos.Password = credentials.Password
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

	return hostUuid, publicCredentialsChan
}

func watch(serviceKey string, id uint64, credentialsChan chan Credentials, stop chan struct{}) {
	done := make(chan struct{}, 1) // Signal received when the watcher found a modification or has lost the connexion to etcd
	ctx, cancelFunc := context.WithCancel(context.Background())
	done <- struct{}{} // Bootstrap to launch the first watcher instance

	// id is the index of the last modification made to the key. The watcher will start watching for modifications done after
	// this index. This will prevent packet or modification lost.

	for {
		select {
		case <-stop:
			cancelFunc()
			return
		case <-done:
			go func() {
				watcher := KAPI().Watcher(serviceKey, &etcd.WatcherOptions{
					AfterIndex: id,
				})
				resp, err := watcher.Next(ctx)
				if err != nil {
					// We've lost the connexion to etcd. Speel 1s and retry
					logger.Printf("lost watcher of '%v': '%v' (%v)", serviceKey, err, Client().Endpoints())
					time.Sleep(1 * time.Second)
					done <- struct{}{}
					return
				}
				var serviceInfos Service
				err = json.Unmarshal([]byte(resp.Node.Value), &serviceInfos)
				if err != nil {
					logger.Printf("error while getting service key '%v': '%v' (%v)", serviceKey, err, Client().Endpoints())
					time.Sleep(1 * time.Second)
					done <- struct{}{}
					return
				}
				// We've got the modification, send it to the register agent
				id = resp.Node.ModifiedIndex
				credentialsChan <- Credentials{
					User:     serviceInfos.User,
					Password: serviceInfos.Password,
				}
				done <- struct{}{} // Restart the process
			}()
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
