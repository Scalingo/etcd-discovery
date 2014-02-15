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
	_, err := client.Create(key, value, HEARTBEAT_DURATION)

	// If the service has been restarted, we update the TTL
	if err != nil {
		if etcdErr, ok := err.(*etcd.EtcdError); ok && etcdErr.ErrorCode == 105 {
			client.Update(key, value, HEARTBEAT_DURATION)
		} else {
			return err
		}
	}

	go func() {
		ticker := time.NewTicker((HEARTBEAT_DURATION - 1) * time.Second)
		for {
			select {
			case <-stop:
				ticker.Stop()
				return
			case <-ticker.C:
				client.Update(key, value, HEARTBEAT_DURATION)
			}
		}
	}()

	return nil
}
