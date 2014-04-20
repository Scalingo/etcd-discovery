package service

import (
	"log"
	"os"
	"strings"

	"github.com/coreos/go-etcd/etcd"
)

var (
	logger   *log.Logger
	Client   *etcd.Client
	hostname string

	cacert  = os.Getenv("ETCD_CACERT")
	tlskey  = os.Getenv("ETCD_TLS_KEY")
	tlscert = os.Getenv("ETCD_TLS_CERT")
)

func init() {
	host := "http://localhost:4001"
	if len(os.Getenv("ETCD_HOST")) != 0 {
		host = os.Getenv("ETCD_HOST")
	}
	if len(cacert) != 0 && len(tlskey) != 0 && len(tlscert) != 0 {
		if !strings.Contains(host, "https://") {
			host = strings.Replace(host, "http", "https", 1)
		}
		Client = newTLSClient([]string{host})
	} else {
		Client = etcd.NewClient([]string{host})
	}

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

func newTLSClient(hosts []string) *etcd.Client {
	c, err := etcd.NewTLSClient(hosts, tlscert, tlskey, cacert)
	if err != nil {
		panic(err)
	}
	return c
}
