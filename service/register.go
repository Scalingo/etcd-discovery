package service

import (
	"encoding/json"
	"fmt"
	"github.com/coreos/go-etcd/etcd"
	"time"
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
		for {
			select {
			case <-stop:
				ticker.Stop()
				return
			case <-ticker.C:
				err := updateOrCreate(key, value)
				// If for any random reason, there is an error,
				// we retry every second until it's ok.
				for err != nil {
					errEtcd := err.(*etcd.EtcdError)
					logger.Println("Lost etcd registeration for", service, ":", errEtcd.ErrorCode)
					time.Sleep(1 * time.Second)
					err = createOrUpdate(key, value)
					if err == nil {
						logger.Println("Recover etcd registeration for", service)
					}
				}
			}
		}
	}()

	return nil
}

func createOrUpdate(k, v string) error {
	_, errCreate := client.Create(k, v, HEARTBEAT_DURATION)
	if errCreate != nil {
		if IsKeyAlreadyExistError(errCreate) {
			_, errUpdate := client.Update(k, v, HEARTBEAT_DURATION)
			if errUpdate != nil {
				return errUpdate
			}
		} else {
			return errCreate
		}
	}
	return nil
}

func updateOrCreate(k, v string) error {
	_, errUpdate := client.Update(k, v, HEARTBEAT_DURATION)
	if errUpdate != nil {
		if IsKeyNotFoundError(errUpdate) {
			_, errCreate := client.Create(k, v, HEARTBEAT_DURATION)
			if errCreate != nil {
				return errCreate
			}
		} else {
			return errUpdate
		}
	}
	return nil
}
