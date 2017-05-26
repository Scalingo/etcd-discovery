package service

import (
	"context"
	"errors"
	"fmt"
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

func (s *Service) Url(scheme, path string) (string, error) {
	if len(s.PublicHostname) == 0 {
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

	var url string
	var port string
	var ok bool
	if port, ok = s.PublicPorts[scheme]; !ok {
		return "", errors.New("unknown scheme")
	}

	hostname = s.PublicHostname

	if s.User != "" {
		url = fmt.Sprintf("%s://%s:%s@%s:%s%s",
			scheme, s.User, s.Password, hostname, port, path,
		)
	} else {
		url = fmt.Sprintf("%s://%s:%s%s",
			scheme, hostname, port, path,
		)
	}
	return url, nil
}
