package service

import (
	"fmt"
)

type Host struct {
	Name     string `json:"name"`
	Port     string `json:"port"`
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
	Scheme   string `json:"scheme"`
}

func NewHost(params ...string) *Host {
	h := &Host{}
	if len(params) == 0 || len(params) > 5 || len(params) == 3 {
		return nil
	}

	h.Name = params[0]
	h.Port = params[1]
	if len(params) >= 4 {
		h.User = params[2]
		h.Password = params[3]
	}
	if len(params) == 5 {
		h.Scheme = params[4]
	} else {
		h.Scheme = "http"
	}
	return h
}

func (h *Host) Url(path string) string {
	if h.User != "" {
		return fmt.Sprintf("%s://%s:%s@%s:%s%s",
			h.Scheme, h.User, h.Password, h.Name, h.Port, path,
		)
	} else {
		return fmt.Sprintf("%s://%s:%s%s",
			h.Scheme, h.Name, h.Port, path,
		)
	}
}
