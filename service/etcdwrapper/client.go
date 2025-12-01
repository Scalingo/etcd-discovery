package etcdwrapper

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"go.etcd.io/etcd/client/pkg/v3/transport"
	etcdv2 "go.etcd.io/etcd/client/v2"
	etcdv3 "go.etcd.io/etcd/client/v3"
)

var (
	clientSingleton    etcdv2.Client
	clientSingletonO   = &sync.Once{}
	clientV3Singleton  *etcdv3.Client
	clientV3SingletonO = &sync.Once{}
)

// KAPI provide a etcd KeysAPI for a client provided by the Client() method
func KAPI() etcdv2.KeysAPI {
	return etcdv2.NewKeysAPI(client())
}

// KAPIV3 provide a etcd KeysAPI for a client provided by the ClientV3() method
func KAPIV3() etcdv3.KV {
	return etcdv3.NewKV(clientV3())
}

func leaseV3() etcdv3.Lease {
	return etcdv3.NewLease(clientV3())
}

func watcherV3() etcdv3.Watcher {
	return etcdv3.NewWatcher(clientV3())
}

// client generates a valid etcd client from the following environment variables:
//   - ETCD_HOSTS: a list of etcd hosts comma separated
//   - ETCD_HOST: a single etcd host
//   - ETCD_CACERT: The CA certificate
//   - ETCD_TLS_CERT: The client TLS cert
//   - ETCD_TLS_KEY: The client TLS key
//   - ETCD_TLS_INMEMORY: Is the TLS configuration filename or raw certificates
func client() etcdv2.Client {
	clientSingletonO.Do(func() {
		var err error

		hosts := []string{"http://localhost:2379"}
		switch {
		case os.Getenv("ETCD_HOSTS") != "":
			hosts = strings.Split(os.Getenv("ETCD_HOSTS"), ",")
		case os.Getenv("ETCD_HOST") != "":
			hosts = []string{os.Getenv("ETCD_HOST")}
		case os.Getenv("ETCD_1_PORT_2379_TCP_ADDR") != "":
			hosts = []string{
				"http://" +
					os.Getenv("ETCD_1_PORT_2379_TCP_ADDR") +
					":" + os.Getenv("ETCD_1_PORT_2379_TCP_PORT"),
			}
		}

		cacert := os.Getenv("ETCD_CACERT")
		tlskey := os.Getenv("ETCD_TLS_KEY")
		tlscert := os.Getenv("ETCD_TLS_CERT")
		if cacert != "" && tlskey != "" && tlscert != "" { //nolint: nestif
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
			c, err := etcdv2.New(etcdv2.Config{Endpoints: hosts, Transport: transport})
			if err != nil {
				panic(err)
			}

			clientSingleton = c
		} else {
			clientSingleton, err = etcdv2.New(etcdv2.Config{Endpoints: hosts, Transport: etcdv2.DefaultTransport})
			if err != nil {
				panic(err)
			}
		}
	})
	return clientSingleton
}

// clientV3 generates a valid etcd client from the following environment variables:
//   - ETCD_HOSTS: a list of etcd hosts comma separated
//   - ETCD_HOST: a single etcd host
//   - ETCD_CACERT: The CA certificate
//   - ETCD_TLS_CERT: The client TLS cert
//   - ETCD_TLS_KEY: The client TLS key
//   - ETCD_TLS_INMEMORY: Is the TLS configuration filename or raw certificates
func clientV3() *etcdv3.Client {
	clientV3SingletonO.Do(func() {
		var err error

		hosts := []string{"http://localhost:2379"}
		switch {
		case os.Getenv("ETCD_HOSTS") != "":
			hosts = strings.Split(os.Getenv("ETCD_HOSTS"), ",")
		case os.Getenv("ETCD_HOST") != "":
			hosts = []string{os.Getenv("ETCD_HOST")}
		case os.Getenv("ETCD_1_PORT_2379_TCP_ADDR") != "":
			hosts = []string{
				"http://" +
					os.Getenv("ETCD_1_PORT_2379_TCP_ADDR") +
					":" + os.Getenv("ETCD_1_PORT_2379_TCP_PORT"),
			}
		}

		cacert := os.Getenv("ETCD_CACERT")
		tlskey := os.Getenv("ETCD_TLS_KEY")
		tlscert := os.Getenv("ETCD_TLS_CERT")
		if cacert != "" && tlskey != "" && tlscert != "" { //nolint: nestif
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

			c, err := etcdv3.New(etcdv3.Config{Endpoints: hosts, TLS: tlsconfig})
			if err != nil {
				panic(err)
			}

			clientV3Singleton = c
		} else {
			clientV3Singleton, err = etcdv3.New(etcdv3.Config{Endpoints: hosts})
			if err != nil {
				panic(err)
			}
		}
	})

	return clientV3Singleton
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
