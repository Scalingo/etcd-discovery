package service

import (
	"log"
	"os"
	"strings"
	"sync"

	"github.com/coreos/go-etcd/etcd"
)

var (
	logger           *log.Logger
	clientSingleton  *etcd.Client
	clientSingletonO = &sync.Once{}
	hostname         string
)

func Client() *etcd.Client {
	clientSingletonO.Do(func() {
		hosts := []string{"http://localhost:4001"}
		if len(os.Getenv("ETCD_HOSTS")) != 0 {
			hosts = strings.Split(os.Getenv("ETCD_HOSTS"), ",")
		} else if len(os.Getenv("ETCD_HOST")) != 0 {
			hosts = []string{os.Getenv("ETCD_HOST")}
		} else if len(os.Getenv("ETCD_1_PORT_4001_TCP_ADDR")) != 0 {
			hosts = []string{
				"http://" +
					os.Getenv("ETCD_1_PORT_4001_TCP_ADDR") +
					":" + os.Getenv("ETCD_1_PORT_4001_TCP_PORT"),
			}
		}

		cacert := os.Getenv("ETCD_CACERT")
		tlskey := os.Getenv("ETCD_TLS_KEY")
		tlscert := os.Getenv("ETCD_TLS_CERT")
		if len(cacert) != 0 && len(tlskey) != 0 && len(tlscert) != 0 {
			for i, host := range hosts {
				if !strings.Contains(host, "https://") {
					hosts[i] = strings.Replace(host, "http", "https", 1)
				}
			}
			c, err := etcd.NewTLSClient(hosts, tlscert, tlskey, cacert)
			if err != nil {
				panic(err)
			}
			clientSingleton = c
		} else {
			clientSingleton = etcd.NewClient(hosts)
		}
	})
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

	logger = log.New(os.Stderr, "[etcd-discovery] ", log.LstdFlags)
}
