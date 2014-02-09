package service

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	HEARTBEAT_DURATION = 5
)

func Register(service string, host *Host, stop chan bool) {
	if len(host.Name) == 0 {
		host.Name = hostname
	}

	key := fmt.Sprintf("/services/%s/%s", service, host.Name)
	hostJson, _ := json.Marshal(&host)

	client.Create(key, string(hostJson), HEARTBEAT_DURATION)

	go func() {
		ticker := time.NewTicker((HEARTBEAT_DURATION - 1) * time.Second)
		for {
			select {
			case <-stop:
				ticker.Stop()
				return
			case <-ticker.C:
				client.UpdateDir(key, HEARTBEAT_DURATION)
			}
		}
	}()
}
