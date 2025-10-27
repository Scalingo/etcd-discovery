package service

import (
	"context"
	"path"

	etcd "go.etcd.io/etcd/client/v2"

	"github.com/Scalingo/etcd-discovery/v8/service/etcdwrapper"
)

// Subscribe to every event that happen to a service
func Subscribe(service string) etcd.Watcher {
	return etcdwrapper.KAPI().Watcher("/services/"+service, &etcd.WatcherOptions{Recursive: true})
}

// SubscribeDown return a channel that will notice you everytime a host loose his etcd registration
func SubscribeDown(service string) (<-chan string, <-chan *etcd.Error) {
	expirations := make(chan string)
	errs := make(chan *etcd.Error)
	watcher := Subscribe(service)
	go func() {
		var (
			res *etcd.Response
			err error
		)

		for {
			res, err = watcher.Next(context.Background())
			if err != nil {
				break
			}

			if res.Action == "expire" || res.Action == "delete" {
				expirations <- path.Base(res.Node.Key)
			}
		}
		if err != nil {
			errs <- err.(*etcd.Error)
		}
		close(expirations)
		close(errs)
	}()
	return expirations, errs
}

// SubscribeNew return a channel that will notice you everytime a new host is registred.
func SubscribeNew(service string) (<-chan *Host, <-chan *etcd.Error) {
	hosts := make(chan *Host)
	errs := make(chan *etcd.Error)
	watcher := Subscribe(service)
	go func() {
		var (
			res *etcd.Response
			err error
		)

		for {
			res, err = watcher.Next(context.Background())
			if err != nil {
				break
			}

			if res.Action == "create" || (res.PrevNode == nil && res.Action == "set") {
				host, err := buildHostFromNode(res.Node)
				if err == nil {
					hosts <- host
				}
			}
		}
		if err != nil {
			errs <- err.(*etcd.Error)
		}
		close(hosts)
		close(errs)
	}()
	return hosts, errs
}
