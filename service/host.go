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
		if len(h.PrivateHostname) != 0 {
			names = append(names, h.PrivateHostname)
		} else {
			names = append(names, h.Hostname)
		}
	}
	return strings.Join(names, ", ")
}

type Host struct {
	Hostname        string `json:"name"`
	Name            string `json:"service_name"` // Will be overwritten by the register function
	Ports           Ports  `json:"ports"`
	User            string `json:"user,omitempty"`
	Password        string `json:"password,omitempty"`
	Public          bool   `json:"public,omitempty"`
	PrivateHostname string `json:"private_hostname,omitempty"` // Will defaults to Hostname
	PrivatePorts    Ports  `json:"private_ports,omitempty"`    // Will defaults to Port
	Critical        bool   `json:"critical,omitempty"`
	Uuid            string `json:"uuid, omitempty"` // Will be overwritten by the register function
}

func (h *Host) Url(scheme, path string) (string, error) {
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

func (h *Host) PrivateUrl(scheme, path string) (string, error) {
	if len(h.PrivateHostname) == 0 {
		return "", errors.New("This service does not support private urls")
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
