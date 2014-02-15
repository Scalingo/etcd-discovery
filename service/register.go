package service

import (
	"encoding/json"
	"fmt"
	"runtime/debug"
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
	err := createOrUpdate(key, value)
	if err != nil {
		return err
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Printf("Error when updating %v %v: %v\n", service, host, r)
				debug.PrintStack()
			}
		}()

		ticker := time.NewTicker((HEARTBEAT_DURATION - 1) * time.Second)
		for {
			select {
			case <-stop:
				ticker.Stop()
				return
			case <-ticker.C:
				err := updateOrCreate(key, value)
				if err != nil {
					panic(err)
				}
			}
		}
	}()

	return nil
}

func createOrUpdate(k, v string) error {
	_, err := client.Create(k, v, HEARTBEAT_DURATION)
	if err != nil {
		if IsKeyAlreadyExistError(err) {
			_, err = client.Update(k, v, HEARTBEAT_DURATION)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

func updateOrCreate(k, v string) error {
	_, err := client.Update(k, v, HEARTBEAT_DURATION)
	if err != nil {
		if IsKeyAlreadyExistError(err) {
			_, err = client.Create(k, v, HEARTBEAT_DURATION)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}
