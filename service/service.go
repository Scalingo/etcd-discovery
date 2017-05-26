package service

import (
	"context"
	"errors"
	"math/rand"

	etcd "github.com/coreos/etcd/client"
	errgo "gopkg.in/errgo.v1"
)

type Service struct {
	Name           string `json:"name"`
	Critical       bool   `json:"critical"`                  // Is the service critical to the infrastructure health?
	PublicHostname string `json:"public_hostname,omitempty"` // The service public hostname
	User           string `json:"user,omitempty"`            // The service username
	Password       string `json:"password,omitempty"`        // The service password
	PublicPorts    Ports  `json:"public_ports,omitempty"`    // The service public ports
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
