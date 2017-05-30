package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"

	etcd "github.com/coreos/etcd/client"
	errgo "gopkg.in/errgo.v1"
)

type Service struct {
	Name     string `json:"name"`
	Critical bool   `json:"critical"`           // Is the service critical to the infrastructure health?
	Hostname string `json:"hostname,omitempty"` // The service private hostname
	User     string `json:"user,omitempty"`     // The service username
	Password string `json:"password,omitempty"` // The service password
	Ports    Ports  `json:"ports,omitempty"`    // The service private ports
	Public   bool   `json:"public,omitempty"`   // Is the service public?
}

type Credentials struct {
	User     string
	Password string
}

func (s *Service) All() (Hosts, error) {
	res, err := KAPI().Get(context.Background(), "/services/"+s.Name, &etcd.GetOptions{
		Recursive: true,
	})

	if err != nil {
		if etcd.IsKeyNotFound(err) {
			return Hosts{}, nil
		}
		return nil, errgo.Notef(err, "Unable to fetch services")
	}

	hosts, err := buildHostsFromNodes(res.Node.Nodes)
	if err != nil {
		return nil, errgo.Mask(err)
	}

	return hosts, nil
}

func (s *Service) First() (*Host, error) {
	hosts, err := s.All()
	if err != nil {
		return nil, errgo.Mask(err)
	}

	if len(hosts) == 0 {
		return nil, errors.New("No host found for this service")
	}

	return hosts[0], nil
}

func (s *Service) One() (*Host, error) {
	hosts, err := s.All()

	if err != nil {
		return nil, errgo.Mask(err)
	}

	if len(hosts) == 0 {
		return nil, errors.New("No host found for this service")
	}

	return hosts[rand.Int()%len(hosts)], nil
}

func (s *Service) Url(scheme, path string) (string, error) {
	log.Println(s.Public)
	if !s.Public { // If the service is not public, fallback to a random node
		host, err := s.One()
		if err != nil {
			return "", errgo.Mask(err)
		}

		url, err := host.Url(scheme, path)
		if err != nil {
			return "", errgo.Mask(err)
		}
		return url, nil
	}

	// If the service IS public, take the service node.

	var url string
	var port string
	var ok bool
	if port, ok = s.Ports[scheme]; !ok {
		return "", errors.New("unknown scheme")
	}

	if s.User != "" {
		url = fmt.Sprintf("%s://%s:%s@%s:%s%s",
			scheme, s.User, s.Password, s.Hostname, port, path,
		)
	} else {
		url = fmt.Sprintf("%s://%s:%s%s",
			scheme, s.Hostname, port, path,
		)
	}
	return url, nil
}
