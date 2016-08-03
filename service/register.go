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

func Register(service string, host *Host, stop chan struct{}) (chan struct{}, error) {
	if len(host.Name) == 0 {
		host.Name = hostname
	}

	registered := make(chan struct{}, 1)
	key := fmt.Sprintf("/services/%s/%s", service, host.Name)
	hostJson, _ := json.Marshal(&host)
	value := string(hostJson)

	go func() {
		ticker := time.NewTicker((HEARTBEAT_DURATION - 1) * time.Second)
		KAPI().Set(context.Background(), key, value, &etcd.SetOptions{TTL: HEARTBEAT_DURATION * time.Second})
		registered <- struct{}{}
		for {
			select {
			case <-stop:
				_, err := KAPI().Delete(context.Background(), key, &etcd.DeleteOptions{Recursive: false})
				if err != nil {
					logger.Println("fail to remove key", key)
				}
				ticker.Stop()
				return
			case <-ticker.C:
				_, err := KAPI().Set(context.Background(), key, value, &etcd.SetOptions{TTL: HEARTBEAT_DURATION * time.Second})
				// If for any random reason, there is an error,
				// we retry every second until it's ok.
				for err != nil {
					logger.Printf("lost registration of '%v': %v (%v)", service, err, Client().Endpoints())
					time.Sleep(1 * time.Second)
					_, err = KAPI().Set(context.Background(), key, value, &etcd.SetOptions{TTL: HEARTBEAT_DURATION * time.Second})
					if err == nil {
						logger.Printf("recover registration of '%v'", service)
					}
				}
			}
		}
	}()

	return registered, nil
}
