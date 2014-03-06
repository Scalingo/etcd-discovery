package service

import (
	"fmt"
)

type Host struct {
	Name     string
	User     string
	Password string
	Port     string
}

func (h *Host) Url(path string) string {
	return fmt.Sprintf("http://%s:%s@%s:%s%s",
		h.User, h.Password, h.Name, h.Port, path,
	)
}
