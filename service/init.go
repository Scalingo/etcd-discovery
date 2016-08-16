package service

import (
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	etcd "github.com/coreos/etcd/client"
	"github.com/coreos/etcd/pkg/transport"
)

var (
	logger           *log.Logger
	clientSingleton  etcd.Client
	clientSingletonO = &sync.Once{}
	hostname         string
)

func KAPI() etcd.KeysAPI {
	return etcd.NewKeysAPI(Client())
}

func Client() etcd.Client {
	clientSingletonO.Do(func() {
		var err error

		hosts := []string{"http://localhost:2379"}
		if len(os.Getenv("ETCD_HOSTS")) != 0 {
			hosts = strings.Split(os.Getenv("ETCD_HOSTS"), ",")
		} else if len(os.Getenv("ETCD_HOST")) != 0 {
			hosts = []string{os.Getenv("ETCD_HOST")}
		} else if len(os.Getenv("ETCD_1_PORT_2379_TCP_ADDR")) != 0 {
			hosts = []string{
				"http://" +
					os.Getenv("ETCD_1_PORT_2379_TCP_ADDR") +
					":" + os.Getenv("ETCD_1_PORT_2379_TCP_PORT"),
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
			info := transport.TLSInfo{
				CertFile: tlscert, KeyFile: tlskey, TrustedCAFile: cacert,
			}
			tlsconfig, err := info.ClientConfig()
			if err != nil {
				panic(err)
			}

			transport := &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				Dial: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).Dial,
				TLSHandshakeTimeout: 10 * time.Second,
				TLSClientConfig:     tlsconfig,
			}
			c, err := etcd.New(etcd.Config{Endpoints: hosts, Transport: transport})
			if err != nil {
				panic(err)
			}

			clientSingleton = c
		} else {
			clientSingleton, err = etcd.New(etcd.Config{Endpoints: hosts, Transport: etcd.DefaultTransport})
			if err != nil {
				panic(err)
			}
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
