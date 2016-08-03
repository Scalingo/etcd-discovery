package service

import (
	"path"

	"golang.org/x/net/context"

	etcd "github.com/coreos/etcd/client"
)

func Subscribe(service string) etcd.Watcher {
	return KAPI().Watcher("/services/"+service, &etcd.WatcherOptions{Recursive: true})
}

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
				hosts <- buildHostFromNode(res.Node)
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

func SubscribeUpdate(service string) (<-chan *Host, <-chan *etcd.Error) {
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
			if res.Action == "update" || (res.PrevNode != nil && res.Action == "set") {
				hosts <- buildHostFromNode(res.Node)
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
