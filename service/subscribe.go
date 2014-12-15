package service

import (
	"path"

	"github.com/coreos/go-etcd/etcd"
)

func Subscribe(service string) (<-chan *etcd.Response, <-chan *etcd.EtcdError) {
	stop := make(chan bool)
	responses := make(chan *etcd.Response)
	errors := make(chan *etcd.EtcdError)
	go func() {
		_, err := Client().Watch("/services/"+service, 0, true, responses, stop)
		if err != nil {
			errors <- err.(*etcd.EtcdError)
			close(errors)
			close(stop)
			return
		}
	}()
	return responses, errors
}

func SubscribeDown(service string) (<-chan string, <-chan *etcd.EtcdError) {
	expirations := make(chan string)
	responses, errors := Subscribe(service)
	go func() {
		for response := range responses {
			if response.Action == "expire" || response.Action == "delete" {
				expirations <- path.Base(response.Node.Key)
			}
		}
	}()
	return expirations, errors
}

func SubscribeNew(service string) (<-chan *Host, <-chan *etcd.EtcdError) {
	hosts := make(chan *Host)
	responses, errors := Subscribe(service)
	go func() {
		for response := range responses {
			if response.Action == "create" || (response.PrevNode == nil && response.Action == "set") {
				hosts <- buildHostFromNode(response.Node)
			}
		}
	}()
	return hosts, errors
}

func SubscribeUpdate(service string) (<-chan *Host, <-chan *etcd.EtcdError) {
	hosts := make(chan *Host)
	responses, errors := Subscribe(service)
	go func() {
		for response := range responses {
			if response.Action == "update" || (response.PrevNode != nil && response.Action == "set") {
				hosts <- buildHostFromNode(response.Node)
			}
		}
	}()
	return hosts, errors
}
