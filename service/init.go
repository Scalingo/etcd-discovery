package service

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
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

// KAPI provide a etcd KeysAPI for a client provided by the Client() method
func KAPI() etcd.KeysAPI {
	return etcd.NewKeysAPI(Client())
}

// Client will generate a valid etcd client from the following environment variables:
//	* ETCD_HOSTS: a list of etcd hosts comma separated
//	* ETCD_HOST: a single etcd host
//	* ETCD_CACERT: The ca certificate
//	* ETCD_TLS_CERT: The client tls cert
// 	* ETCD_TLS_KEY: The client tls key
// 	* ETCD_TLS_INMEMORY: Is the tls configuration filename or raw certificates
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

			var tlsconfig *tls.Config
			if os.Getenv("ETCD_TLS_INMEMORY") == "true" {
				tlsconfig, err = tlsconfigFromMemory(tlscert, tlskey, cacert)
			} else {
				tlsconfig, err = tlsconfigFromFiles(tlscert, tlskey, cacert)
			}
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

func tlsconfigFromFiles(cert, key, ca string) (*tls.Config, error) {
	return transport.TLSInfo{
		CertFile: cert, KeyFile: key, TrustedCAFile: ca,
	}.ClientConfig()
}

func tlsconfigFromMemory(certb64, keyb64, cab64 string) (*tls.Config, error) {
	certpem, err := base64.StdEncoding.DecodeString(certb64)
	if err != nil {
		return nil, err
	}

	keypem, err := base64.StdEncoding.DecodeString(keyb64)
	if err != nil {
		return nil, err
	}

	capem, err := base64.StdEncoding.DecodeString(cab64)
	if err != nil {
		return nil, err
	}

	ca, _ := pem.Decode(capem)
	if ca == nil {
		return nil, errors.New("ca: invalid PEM")
	}

	certPool := x509.NewCertPool()
	caCert, err := x509.ParseCertificate(ca.Bytes)
	if err != nil {
		return nil, fmt.Errorf("ca: not a valid certificate %v", err)
	}
	certPool.AddCert(caCert)

	certkey, err := tls.X509KeyPair(certpem, keypem)
	if err != nil {
		return nil, fmt.Errorf("cert/key: invalid key pair/certificate %v", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{certkey},
		RootCAs:      certPool,
		MinVersion:   tls.VersionTLS12,
	}, nil
}
