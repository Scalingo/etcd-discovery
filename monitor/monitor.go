package monitor

import (
	"errors"
	"github.com/Appsdeck/etcd-discovery/service"
	"log"
	"os"
	"time"
)

type Service struct {
	Name  string
	Hosts []*service.Host
}
type Services map[string]*Service

var (
	services           = Services{}
	logger             = log.New(os.Stdout, "[etcd-discovery] ", log.LstdFlags)
	NoSuchServiceError = errors.New("no such service")
)

func Start(name string) error {
	services[name] = &Service{name, nil}
	s := services[name]
	s.Populate()

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
			s.Save(h)
		case h := <-updateHosts:
			s.Save(h)
		case hostname := <-deadHosts:
			s.Remove(hostname)
		}
	}
	return nil
}

func Hosts(service string) ([]*service.Host, error) {
	if _, ok := services[service]; !ok {
		return nil, NoSuchServiceError
	}
	return services[service].Hosts, nil
}

func (s *Service) Populate() error {
	hosts, err := service.Get(s.Name)
	if err != nil {
		return err
	}
	for _, h := range hosts {
		s.Save(h)
	}
	return nil
}

func (s *Service) Save(newHost *service.Host) {
	for _, h := range s.Hosts {
		if h.Name == newHost.Name {
			h.Port = newHost.Port
			h.Password = newHost.Password
			h.User = newHost.User
			return
		}
	}
	s.Hosts = append(s.Hosts, newHost)
}

func (s *Service) Remove(hostname string) {
	for i, h := range s.Hosts {
		if h.Name == hostname {
			s.Hosts = append(s.Hosts[:i], s.Hosts[i+1:]...)
		}
	}
}
