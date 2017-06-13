package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"

	etcd "github.com/coreos/etcd/client"
	errgo "gopkg.in/errgo.v1"
)

// Service store all the informatiosn about a service. This is also used to marshal services present in the /services_infos/ directory.
type Service struct {
	Name     string `json:"name"`               // Name of the service
	Critical bool   `json:"critical"`           // Is the service critical to the infrastructure health?
	Hostname string `json:"hostname,omitempty"` // The service private hostname
	User     string `json:"user,omitempty"`     // The service username
	Password string `json:"password,omitempty"` // The service password
	Ports    Ports  `json:"ports,omitempty"`    // The service private ports
	Public   bool   `json:"public,omitempty"`   // Is the service public?
}

// Credentials store service credentials
type Credentials struct {
	User     string
	Password string
}

// All return all hosts associated to a service
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

// First return the first host of this service
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

// One return a random host from all the available hosts of this service.
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

// URL return the public url of this service. If this service do not have an public url, this will return an url to a random host.
func (s *Service) URL(scheme, path string) (string, error) {
	if !s.Public { // If the service is not public, fallback to a random node
		host, err := s.One()
		if err != nil {
			return "", errgo.Mask(err)
		}

		url, err := host.URL(scheme, path)
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
