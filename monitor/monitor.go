package monitor

import (
	"errors"
	"github.com/Appsdeck/etcd-discovery/service"
)

var (
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

	newHosts := service.SubscribeNew(name)
	updateHosts := service.SubscribeUpdate(name)
	deadHosts := service.SubscribeDown(name)
	for {
		select {
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
