package service

import (
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
	client.CreateDir(key, HEARTBEAT_DURATION)
	client.Create(key + "/user", host.User, 0)
	client.Create(key + "/password", host.Password, 0)
	client.Create(key + "/port", host.Port, 0)

	go func() {
		ticker := time.NewTimer((HEARTBEAT_DURATION - 1) * time.Second)
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				client.UpdateDir(key, HEARTBEAT_DURATION)
			}
		}
	}()
}
