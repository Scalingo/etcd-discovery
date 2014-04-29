package service

import (
	"log"
	"os"
	"strings"

	"github.com/coreos/go-etcd/etcd"
)

var (
	logger          *log.Logger
	clientSingleton *etcd.Client
	hostname        string
)

func Client() *etcd.Client {
	if clientSingleton == nil {
		host := "http://localhost:4001"
		if len(os.Getenv("ETCD_HOST")) != 0 {
			host = os.Getenv("ETCD_HOST")
		}

		cacert := os.Getenv("ETCD_CACERT")
		tlskey := os.Getenv("ETCD_TLS_KEY")
		tlscert := os.Getenv("ETCD_TLS_CERT")
		if len(cacert) != 0 && len(tlskey) != 0 && len(tlscert) != 0 {
			if !strings.Contains(host, "https://") {
				host = strings.Replace(host, "http", "https", 1)
			}
			c, err := etcd.NewTLSClient([]string{host}, tlscert, tlskey, cacert)
			if err != nil {
				panic(err)
			}
			clientSingleton = c
		} else {
			clientSingleton = etcd.NewClient([]string{host})
		}
	}
	return clientSingleton
}

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

	logger = log.New(os.Stderr, "[etcd-discovery]", log.LstdFlags)
}
