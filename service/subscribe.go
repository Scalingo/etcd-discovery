package service

import (
	"context"
	"errors"
	"path"

	etcdv2 "go.etcd.io/etcd/client/v2"
)

var subscribeWatcher = Subscribe

// Subscribe to every event that happen to a service.
func Subscribe(service string) etcdv2.Watcher {
	return KAPI().Watcher("/services/"+service, &etcdv2.WatcherOptions{Recursive: true})
}

// SubscribeDown returns a channel that will notice you every time a host loses his etcd registration.
// The subscription lifetime is tied to ctx so callers can stop the blocking etcd watch cleanly.
func SubscribeDown(ctx context.Context, service string) (<-chan string, <-chan *etcdv2.Error) {
	expirations := make(chan string)
	errs := make(chan *etcdv2.Error, 1)
	watcher := subscribeWatcher(service)

	go func() {
		var (
			res *etcdv2.Response
			err error
		)

		for {
			// Watch with the caller context so this goroutine exits as soon as the subscription is canceled.
			res, err = watcher.Next(ctx)
			if err != nil {
				break
			}

			if res.Action == "expire" || res.Action == "delete" {
				expirations <- path.Base(res.Node.Key)
			}
		}

		etcdErr := subscriptionError(err)
		if etcdErr != nil {
			errs <- etcdErr
		}

		close(expirations)
		close(errs)
	}()
	return expirations, errs
}

// SubscribeNew returns a channel that will notice you every time a new host is registered.
// The subscription lifetime is tied to ctx so callers can stop the blocking etcd watch cleanly.
func SubscribeNew(ctx context.Context, service string) (<-chan *Host, <-chan *etcdv2.Error) {
	hosts := make(chan *Host)
	errs := make(chan *etcdv2.Error, 1)
	watcher := subscribeWatcher(service)

	go func() {
		var (
			res *etcdv2.Response
			err error
		)

		for {
			// Watch with the caller context so this goroutine exits as soon as the subscription is canceled.
			res, err = watcher.Next(ctx)
			if err != nil {
				break
			}

			if res.Action == "create" || (res.PrevNode == nil && res.Action == "set") {
				// etcd can report the first write for a key as "set" instead of
				// "create". A missing previous node still means a brand-new host.
				host, err := buildHostFromNode(ctx, res.Node)
				if err == nil {
					hosts <- host
				}
			}
		}

		etcdErr := subscriptionError(err)
		if etcdErr != nil {
			errs <- etcdErr
		}

		close(hosts)
		close(errs)
	}()
	return hosts, errs
}

// subscriptionError ignores context cancellation, forwards only etcd errors to
// the errs channel, and supports wrapped errors without panicking on unrelated
// error types.
func subscriptionError(err error) *etcdv2.Error {
	if err == nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return nil
	}

	var etcdErr *etcdv2.Error
	if errors.As(err, &etcdErr) {
		return etcdErr
	}

	var etcdErrValue etcdv2.Error
	if errors.As(err, &etcdErrValue) {
		// Some call paths return etcd.Error by value; copy it to preserve the
		// Subscribe* channel type without requiring callers to special-case it.
		return &etcdErrValue
	}

	return nil
}
