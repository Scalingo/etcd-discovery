package service

import (
	"log"
	"os"
)

var (
	logger   *log.Logger
	hostname string
)

func init() {
	if len(os.Getenv("HOSTNAME")) != 0 {
		hostname = os.Getenv("HOSTNAME")
	} else {
		h, err := os.Hostname()
		if err != nil {
			panic(err)
		}
		hostname = h
	}

	logger = log.New(os.Stderr, "[etcd-discovery] ", log.LstdFlags)
}
