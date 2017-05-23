package service

import (
	"errors"
	"fmt"
	"strings"
)

type Ports map[string]string

type Hosts []*Host

func (hs Hosts) String() string {
	names := []string{}
	for _, h := range hs {
		names = append(names, h.Name)
	}
	return strings.Join(names, ", ")
}

type Host struct {
	Name           string `json:"name"`
	Ports          Ports  `json:"ports"`
	User           string `json:"user,omitempty"`
	Password       string `json:"password,omitempty"`
	PublicHostname string `json:"public_hostname,omitempty"`
}

func NewHost(hostname string, ports Ports, params ...string) (*Host, error) {
	h := &Host{Name: hostname, Ports: ports}
	if ports == nil || len(ports) == 0 {
		return nil, errors.New("ports is nil, an interface should be defined")
	}
	if len(params) == 1 {
		return nil, errors.New("if user is defined, password should be too")
	} else if len(params) == 2 {
		h.User = params[0]
		h.Password = params[1]
	}

	return h, nil
}

func (h *Host) Url(scheme, path string) (string, error) {
	var url string
	var port string
	var ok bool
	if port, ok = h.Ports[scheme]; !ok {
		return "", errors.New("unknown scheme")
	}

	hostname = h.Name

	if len(h.PublicHostname) != 0 {
		hostname = h.PublicHostname
	}
	if h.User != "" {
		url = fmt.Sprintf("%s://%s:%s@%s:%s%s",
			scheme, h.User, h.Password, hostname, port, path,
		)
	} else {
		url = fmt.Sprintf("%s://%s:%s%s",
			scheme, hostname, port, path,
		)
	}
	return url, nil
}
