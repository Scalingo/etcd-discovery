package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gopkg.in/errgo.v1"
)

// Ports is a representation of the ports exposed by a host or a service.
// The key is the protocol name, and the value is the port used for this protocol.
// Typical usage is:
//
//	Ports{
//		"http":"80",
//		"https": "443",
//	}
type Ports map[string]string

// Hosts will represent a slice of hosts
type Hosts []*Host

func (hs Hosts) String() string {
	names := []string{}
	for _, h := range hs {
		if h.PrivateHostname != "" {
			names = append(names, h.PrivateHostname)
		} else {
			names = append(names, h.Hostname)
		}
	}
	return strings.Join(names, ", ")
}

// Host stores all the host information.
// This is also used to store the host in the etcd services directory.
type Host struct {
	// Hostname  must be a publicly available IP/FQDN if the service is public or a private IP/FQDN id the service is private.
	Hostname string `json:"name"`
	// Name of the service which this host store. This will be overwritten by the Register Function
	Name string `json:"service_name"`
	// Ports is the ports accessible on a public network if the service is Public or the ports accessible on a private network if the service is private
	Ports Ports `json:"ports"`
	// User name used to authenticate to this service. The Register function can override this at any time if the service is public
	User string `json:"user,omitempty"`
	// Password used to authenticate to this service. The Register function can override this at any time if the service is public
	Password string `json:"password,omitempty"`
	// Public is set to true if the service is public
	Public bool `json:"public,omitempty"`
	// PrivateHostname is the private IP/FQDN used to communicate with this service on the private network
	// This will defaults to Hostname if empty
	PrivateHostname string `json:"private_hostname,omitempty"`
	// PrivatePorts is the private ports used to communicate with this service on the private network.
	// This will defaults to Ports if empty
	PrivatePorts Ports `json:"private_ports,omitempty"`
	// Critical will be set to true is the service is critical
	Critical bool `json:"critical,omitempty"`
	// Shard identifies instances of this service deployed on a shard.
	//
	// This field is an empty string if the service is not sharded.
	Shard string `json:"shard,omitempty"`
	// UUID is the service UUID, this must have the following pattern: uuid-PrivateHostname
	UUID string `json:"uuid,omitempty"`
}

// URL will return a valid url to contact this service on the specific protocol provided by the scheme parameter
func (h *Host) URL(ctx context.Context, scheme, path string) (string, error) {
	var (
		url  string
		port string
		ok   bool
		err  error
	)

	if scheme != "" {
		if port, ok = h.Ports[scheme]; !ok {
			return "", errors.New("unknown scheme")
		}
	} else {
		scheme, port, err = h.findSchemeAndPort(h.Ports)
		if err != nil {
			return "", errgo.Notef(err, "find scheme and port")
		}
	}

	if h.User != "" {
		url = fmt.Sprintf("%s://%s:%s@%s:%s%s", scheme, h.User, h.Password, h.Hostname, port, path)
	} else {
		url = fmt.Sprintf("%s://%s:%s%s", scheme, h.Hostname, port, path)
	}
	return url, nil
}

// PrivateURL will provide a valid url to contact this service on the Private network
// this method will fall back to the URL method if the host does not provide any PrivateURL
func (h *Host) PrivateURL(ctx context.Context, scheme, path string) (string, error) {
	if h.PrivateHostname == "" {
		url, err := h.URL(ctx, scheme, path)
		if err != nil {
			return "", errgo.Mask(err)
		}
		return url, nil
	}

	var (
		url  string
		port string
		ok   bool
		err  error
	)

	if scheme != "" {
		if port, ok = h.PrivatePorts[scheme]; !ok {
			return "", errors.New("unknown scheme")
		}
	} else {
		scheme, port, err = h.findSchemeAndPort(h.PrivatePorts)
		if err != nil {
			return "", errgo.Notef(err, "find scheme and port")
		}
	}

	if h.User != "" {
		url = fmt.Sprintf("%s://%s:%s@%s:%s%s", scheme, h.User, h.Password, h.PrivateHostname, port, path)
	} else {
		url = fmt.Sprintf("%s://%s:%s%s", scheme, h.PrivateHostname, port, path)
	}
	return url, nil
}

func (h *Host) findSchemeAndPort(ports Ports) (string, string, error) {
	schemes := [2]string{"http", "https"}
	for _, scheme := range schemes {
		if port, ok := ports[scheme]; ok {
			return scheme, port, nil
		}
	}
	return "", "", errors.New("scheme not found")
}

// HostResponse is the interface used to provide a single host response.
// This interface provides a standard API used for method chaining like:
//
//	url, err := Get(ctx, "my-service").First(ctx).URL(ctx, "http", "/")
//
// To provide such API errors need to be stored and sent at the last moment.
// To do so, each "final" method (like URL or Host) will check if the Response is errored before
// continuing to their own logic.
type HostResponse interface {
	Err() error
	Host(ctx context.Context) (*Host, error)
	URL(ctx context.Context, scheme, path string) (string, error)
	PrivateURL(ctx context.Context, scheme, path string) (string, error)
}

// GetHostResponse is the HostResponse Implementation of the HostResponse interface used by the the Service methods.
// This only provide the error wrapping logic, all the actaul logic for these method are done by the Host struct.
type GetHostResponse struct {
	err  error
	host *Host
}

// Err returns an error if an error happened on any previous steps.
func (q *GetHostResponse) Err() error {
	return q.err
}

// Host will return the host represented by this error.
func (q *GetHostResponse) Host(ctx context.Context) (*Host, error) {
	if q.err != nil {
		return nil, q.err
	}

	return q.host, nil
}

// URL will return the public URL of this host
func (q *GetHostResponse) URL(ctx context.Context, scheme, path string) (string, error) {
	if q.err != nil {
		return "", q.err
	}
	url, err := q.host.URL(ctx, scheme, path)
	if err != nil {
		return "", err
	}
	return url, nil
}

// PrivateURL will return the private URL of this host
func (q *GetHostResponse) PrivateURL(ctx context.Context, scheme, path string) (string, error) {
	if q.err != nil {
		return "", q.err
	}
	url, err := q.host.PrivateURL(ctx, scheme, path)
	if err != nil {
		return "", err
	}
	return url, nil
}
