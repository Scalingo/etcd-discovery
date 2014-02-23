package monitor

import (
	"errors"
	"github.com/Appsdeck/etcd-discovery/service"
	"log"
	"os"
	"time"
)

var (
	logger             = log.New(os.Stdout, "[etcd-discovery] ", log.LstdFlags)
	services           = map[string][]*service.Host{}
	NoSuchServiceError = errors.New("no such service")
)

func Start(name string) error {
	hosts, err := service.Get(name)
	if err != nil {
		return err
	}

	currentHosts := make([]*service.Host, len(hosts))
	copy(currentHosts, hosts)
	services[name] = currentHosts

	newHosts, errsNew := service.SubscribeNew(name)
	updateHosts, errsUpd := service.SubscribeUpdate(name)
	deadHosts, errsDown := service.SubscribeDown(name)
	for {
		select {
		case err := <-errsNew:
			logger.Println("Wait subscribe new", name, "etcd (", err.ErrorCode, ")")
			time.Sleep(1 * time.Second)
			newHosts, errsNew = service.SubscribeNew(name)
		case err := <-errsUpd:
			logger.Println("Wait subscribe update", name, "etcd (", err.ErrorCode, ")")
			time.Sleep(1 * time.Second)
			updateHosts, errsUpd = service.SubscribeUpdate(name)
		case err := <-errsDown:
			logger.Println("Wait subscribe down", name, "etcd (", err.ErrorCode, ")")
			time.Sleep(1 * time.Second)
			deadHosts, errsDown = service.SubscribeDown(name)
		case h := <-newHosts:
			currentHosts = append(currentHosts, h)
		case h := <-updateHosts:
			for i, host := range currentHosts {
				if host.Name == h.Name {
					currentHosts[i] = h
				}
			}
		case name := <-deadHosts:
			for i, s := range currentHosts {
				if s.Name == name {
					currentHosts = append(currentHosts[:i], currentHosts[i+1:]...)
				}
			}
		}
		// We update the global state after each operation on the local slice
		services[name] = currentHosts
	}
	return nil
}

func Hosts(service string) ([]*service.Host, error) {
	if _, ok := services[service]; !ok {
		return nil, NoSuchServiceError
	}
	return services[service], nil
}
