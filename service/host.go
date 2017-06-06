package service

import (
	"errors"
	"fmt"
	"strings"

	errgo "gopkg.in/errgo.v1"
)

// Ports is a representation of the ports exposed by a host or a service.
// The key is the protocol name and the value is the port used for this protocol.
// Typical usage is:
// 	Ports{
//		"http":"80",
//		"https": "443",
//  }
type Ports map[string]string

// Hosts will represent a slice of hosts
type Hosts []*Host

func (hs Hosts) String() string {
	names := []string{}
	for _, h := range hs {
		if len(h.PrivateHostname) != 0 {
			names = append(names, h.PrivateHostname)
		} else {
			names = append(names, h.Hostname)
		}
	}
	return strings.Join(names, ", ")
}

// Host store all the host informations. This is also used to store the host in the etcd services directory.
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
	// UUID is the service UUID, this must have the following pattern: uuid-PrivateHostname
	UUID string `json:"uuid, omitempty"`
}

// URL will return a valid url to contact this service on the specific protocol provided by the scheme parameter
func (h *Host) URL(scheme, path string) (string, error) {
	var url string
	var port string
	var ok bool
	if port, ok = h.Ports[scheme]; !ok {
		return "", errors.New("unknown scheme")
	}

	if h.User != "" {
		url = fmt.Sprintf("%s://%s:%s@%s:%s%s",
			scheme, h.User, h.Password, h.Hostname, port, path,
		)
	} else {
		url = fmt.Sprintf("%s://%s:%s%s",
			scheme, h.Hostname, port, path,
		)
	}
	return url, nil
}

// PrivateURL will provide a valid url to contact this service on the Private network
// this method will fallback to the URL method if the host does not provide any PrivateURL
func (h *Host) PrivateURL(scheme, path string) (string, error) {
	if len(h.PrivateHostname) == 0 {
		url, err := h.URL(scheme, path)
		if err != nil {
			return "", errgo.Mask(err)
		} else {
			return url, nil
		}
	}
	var url, port string
	var ok bool
	if port, ok = h.PrivatePorts[scheme]; !ok {
		return "", errors.New("unknown scheme")
	}

	if len(h.User) != 0 {
		url = fmt.Sprintf("%s://%s:%s@%s:%s%s",
			scheme, h.User, h.Password, h.PrivateHostname, port, path,
		)
	} else {
		url = fmt.Sprintf("%s://%s:%s%s",
			scheme, h.PrivateHostname, port, path)
	}
	return url, nil
}

// HostResponse is the interface used to provide a single host response.
// This interface provide a standard API used fot method chaining like:
// 	Get("my-service").First().Url()
//
// To provide such API errores need to be stored and sent at the last moment.
// To do so each "final" methods (like URL or Host) will check if the Response is errored before
// continuing to their own logic.
type HostResponse interface {
	Err() error
	Host() (*Host, error)
	URL(scheme, path string) (string, error)
	PrivateURL(scheme, path string) (string, error)
}

// GetHostResponse is the HostResponse Implementation of the HostResponse interface used by the the Service methods.
// This only provide the error wrapping logic, all the actaul logic for these method are done by the Host struct.
type GetHostResponse struct {
	err  error
	host *Host
}

// Err will return an error if an error happened on any previous steps.
func (q *GetHostResponse) Err() error {
	return q.err
}

// Host will return the host represented by this error.
func (q *GetHostResponse) Host() (*Host, error) {
	if q.err != nil {
		return nil, errgo.Mask(q.err)
	}

	return q.host, nil
}

// URL will return the public URL of this host
func (q *GetHostResponse) URL(scheme, path string) (string, error) {
	if q.err != nil {
		return "", errgo.Mask(q.err)
	}
	url, err := q.host.URL(scheme, path)
	if err != nil {
		return "", errgo.Mask(err)
	}
	return url, nil
}

// PrivateURL will return the private URL of this host
func (q *GetHostResponse) PrivateURL(scheme, path string) (string, error) {
	if q.err != nil {
		return "", errgo.Mask(q.err)
	}
	url, err := q.host.PrivateURL(scheme, path)
	if err != nil {
		return "", errgo.Mask(err)
	}
	return url, nil
}
