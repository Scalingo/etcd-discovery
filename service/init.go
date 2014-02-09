package service

import (
	"github.com/coreos/go-etcd/etcd"
	"log"
	"os"
)

var (
	logger   *log.Logger
	client   *etcd.Client
	hostname string
)

func init() {
	host := "http://localhost:4001"
	if len(os.Getenv("ETCD_HOST")) != 0 {
		host = os.Getenv("ETCD_HOST")
	}
	client = etcd.NewClient([]string{host})

	if len(os.Getenv("HOSTNAME")) != 0 {
		hostname = os.Getenv("HOSTNAME")
	} else {
		h, err := os.Hostname()
		if err != nil {
			panic(err)
		}
		hostname = h
	}

	logger = log.New(os.Stderr, "[etcd-discovery]", log.LstdFlags)
}
