package service

import (
	"encoding/json"
	"fmt"
	"time"

	etcd "github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

const (
	HEARTBEAT_DURATION = 5
)

type Infos struct {
	Critical bool `json:"critical"` // Is the service critical to the infrastructure health?
}

func Register(service string, host *Host, infos *Infos, stop chan struct{}) (chan struct{}, error) {
	if len(host.Name) == 0 {
		host.Name = hostname
	}

	registered := make(chan struct{}, 1)
	hostKey := fmt.Sprintf("/services/%s/%s", service, host.Name)
	hostJson, _ := json.Marshal(&host)
	hostValue := string(hostJson)

	serviceKey := fmt.Sprintf("/services_infos/%s", service)
	serviceJson, _ := json.Marshal(infos)
	serviceValue := string(serviceJson)

	go func() {
		ticker := time.NewTicker((HEARTBEAT_DURATION - 1) * time.Second)
		KAPI().Set(context.Background(), hostKey, hostValue, &etcd.SetOptions{TTL: HEARTBEAT_DURATION * time.Second})

		if infos != nil {
			KAPI().Set(context.Background(), serviceKey, serviceValue, nil)
		}

		registered <- struct{}{}
		for {
			select {
			case <-stop:
				_, err := KAPI().Delete(context.Background(), hostKey, &etcd.DeleteOptions{Recursive: false})
				if err != nil {
					logger.Println("fail to remove key", hostKey)
				}
				ticker.Stop()
				return
			case <-ticker.C:
				_, err := KAPI().Set(context.Background(), hostKey, hostValue, &etcd.SetOptions{TTL: HEARTBEAT_DURATION * time.Second})
				// If for any random reason, there is an error,
				// we retry every second until it's ok.
				for err != nil {
					logger.Printf("lost registration of '%v': %v (%v)", service, err, Client().Endpoints())
					time.Sleep(1 * time.Second)

					if infos != nil {
						KAPI().Set(context.Background(), serviceKey, serviceValue, nil)
					}

					_, err = KAPI().Set(context.Background(), hostKey, hostValue, &etcd.SetOptions{TTL: HEARTBEAT_DURATION * time.Second})
					if err == nil {
						logger.Printf("recover registration of '%v'", service)
					}
				}
			}
		}
	}()

	return registered, nil
}
