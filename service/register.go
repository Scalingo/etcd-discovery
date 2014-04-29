package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/coreos/go-etcd/etcd"
)

const (
	HEARTBEAT_DURATION = 5
)

func Register(service string, host *Host, stop chan bool) error {
	if len(host.Name) == 0 {
		host.Name = hostname
	}

	key := fmt.Sprintf("/services/%s/%s", service, host.Name)
	hostJson, _ := json.Marshal(&host)
	value := string(hostJson)

	go func() {
		ticker := time.NewTicker((HEARTBEAT_DURATION - 1) * time.Second)
		Client().Set(key, value, HEARTBEAT_DURATION)
		for {
			select {
			case <-stop:
				ticker.Stop()
				return
			case <-ticker.C:
				_, err := Client().Set(key, value, HEARTBEAT_DURATION)
				// If for any random reason, there is an error,
				// we retry every second until it's ok.
				for err != nil {
					errEtcd := err.(*etcd.EtcdError)
					logger.Println("Lost etcd registrationg for", service, ":", errEtcd.ErrorCode)
					time.Sleep(1 * time.Second)
					_, err = Client().Set(key, value, HEARTBEAT_DURATION)
					if err == nil {
						logger.Println("Recover etcd registrationg for", service)
					}
				}
			}
		}
	}()

	return nil
}
