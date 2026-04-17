package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"

	etcdv2 "go.etcd.io/etcd/client/v2"
	"gopkg.in/errgo.v1"
)

var (
	ErrNoServiceFound     = errors.New("service not found")
	ErrNoHostFound        = errors.New("no host found for this service")
	ErrNoHostFoundOnShard = errors.New("no host found for this service on this shard")
	ErrUnknownScheme      = errors.New("unknown scheme")
)

// Service stores all the information about a service.
// This is also used to marshal services present in the /services_infos/ directory.
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

// QueryOptions allows optional filtering for service queries.
type QueryOptions struct {
	Shard string
}

// All returns all hosts associated with a service
func (s *Service) All(ctx context.Context, queryOpts QueryOptions) (Hosts, error) {
	res, err := KAPI().Get(ctx, "/services/"+s.Name, &etcdv2.GetOptions{
		Recursive: true,
	})

	if err != nil {
		if etcdv2.IsKeyNotFound(err) {
			return nil, ErrNoServiceFound
		}
		return nil, errgo.Notef(err, "Unable to fetch services")
	}

	hosts, err := buildHostsFromNodes(ctx, res.Node.Nodes)
	if err != nil {
		return nil, errgo.Mask(err)
	}

	if len(hosts) == 0 {
		return nil, ErrNoHostFound
	}

	// If no shard is specified, return all hosts
	if queryOpts.Shard == "" {
		return hosts, nil
	}

	// If shard is specified, filter hosts by shard
	filteredHosts := make(Hosts, 0, len(hosts))
	for _, host := range hosts {
		if host.Shard == queryOpts.Shard {
			filteredHosts = append(filteredHosts, host)
		}
	}

	if len(filteredHosts) == 0 {
		return nil, ErrNoHostFoundOnShard
	}

	return filteredHosts, nil
}

// First returns the first host of this service
func (s *Service) First(ctx context.Context, queryOpts QueryOptions) (*Host, error) {
	hosts, err := s.All(ctx, queryOpts)
	if err != nil {
		return nil, errgo.Notef(err, "fetch hosts")
	}

	return hosts[0], nil
}

// One returns a random host from all the available hosts of this service.
func (s *Service) One(ctx context.Context, queryOpts QueryOptions) (*Host, error) {
	hosts, err := s.All(ctx, queryOpts)
	if err != nil {
		return nil, errgo.Mask(err)
	}

	return hosts[rand.Int()%len(hosts)], nil
}

// URL returns the public url of this service.
//
// If this service do not have a public url, this will return a url to a random host.
func (s *Service) URL(ctx context.Context, scheme, path string, queryOpts QueryOptions) (string, error) {
	// If the service is not public, fallback to a random host.
	// If a shard is requested, always resolve the URL from a host in that shard.
	//
	// The public metadata stored under /services_infos/<name> is shared across
	// all shards, so it cannot identify the public host for a specific shard.
	if !s.Public || queryOpts.Shard != "" {
		host, err := s.One(ctx, queryOpts)
		if err != nil {
			return "", errgo.Mask(err)
		}

		url, err := host.URL(ctx, scheme, path)
		if err != nil {
			return "", errgo.Mask(err)
		}
		return url, nil
	}

	// If the service is public, take the service node.
	var url string
	var port string
	var ok bool
	if port, ok = s.Ports[scheme]; !ok {
		return "", ErrUnknownScheme
	}

	if s.User != "" {
		url = fmt.Sprintf("%s://%s:%s@%s:%s%s", scheme, s.User, s.Password, s.Hostname, port, path)
	} else {
		url = fmt.Sprintf("%s://%s:%s%s", scheme, s.Hostname, port, path)
	}
	return url, nil
}
