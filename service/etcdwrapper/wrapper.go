package etcdwrapper

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	etcd "go.etcd.io/etcd/client/v2"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"gopkg.in/errgo.v1"
)

const (
	// HEARTBEAT_DURATION time in second between two registration. The host will
	// be deleted if etcd didn't received any new registration in those 5 seocnds
	HEARTBEAT_DURATION = 5
)

func Delete(ctx context.Context, hostKey string) error {
	_, err := KAPI().Delete(ctx, hostKey, &etcd.DeleteOptions{Recursive: false})
	if err != nil {
		return err
	}

	_, err = KAPIV3().Delete(ctx, hostKey)
	if err != nil {
		return err
	}

	return nil
}

func Endpoints() ([]string, []string) {
	return client().Endpoints(), clientV3().Endpoints()
}

func Set(ctx context.Context, key, value string, withLease bool) (uint64, int64, error) {
	var opts *etcd.SetOptions
	var opOptions []etcdv3.OpOption
	if withLease {
		opts = &etcd.SetOptions{TTL: HEARTBEAT_DURATION * time.Second}

		lease, err := leaseV3().Grant(ctx, int64(HEARTBEAT_DURATION))
		if err != nil {
			return 0, 0, errgo.Notef(err, "Unable to create lease for host registration")
		}
		opOptions = append(opOptions, etcdv3.WithLease(lease.ID))
	}

	resV2, err := KAPI().Set(ctx, key, value, opts)
	if err != nil {
		return 0, 0, err
	}

	resV3, err := KAPIV3().Put(ctx, key, value, opOptions...)
	if err != nil {
		return 0, 0, err
	}

	return resV2.Node.ModifiedIndex, resV3.Header.GetRevision(), nil
}

type watchResponse struct {
	User     string
	Password string
}

func Watch(logger *log.Logger, ctx context.Context, key string, idxV2 uint64, idxV3 int64) <-chan watchResponse {
	watchChan := make(chan watchResponse)

	go watchV2(logger, ctx, key, idxV2, watchChan)
	go watchV3(logger, ctx, key, idxV3, watchChan)

	return watchChan
}

func watchV2(logger *log.Logger, ctx context.Context, serviceKey string, id uint64, watchChan chan watchResponse) {
	// id is the index of the last modification made to the key. The watcher will
	// start watching for modifications done after this index. This will prevent
	// packet or modification lost.

	for {
		watcher := KAPI().Watcher(serviceKey, &etcd.WatcherOptions{
			AfterIndex: id,
		})
		resp, err := watcher.Next(ctx)
		if err == context.Canceled {
			return
		}

		if err != nil {
			// We've lost the connexion to etcd. Sleep 1s and retry
			logger.Printf("lost watcher of '%v': '%v' (v2: %v)", serviceKey, err, client().Endpoints())
			id = 0
			time.Sleep(1 * time.Second)
			continue
		}
		var serviceInfos service
		err = json.Unmarshal([]byte(resp.Node.Value), &serviceInfos)
		if err != nil {
			logger.Printf("error while getting service key '%v': '%v' (v2: %v)", serviceKey, err, client().Endpoints())
			time.Sleep(1 * time.Second)
			continue
		}
		// We've got the modification, send it to the register agent
		id = resp.Node.ModifiedIndex
		watchChan <- watchResponse{
			User:     serviceInfos.User,
			Password: serviceInfos.Password,
		}
	}
}

func watchV3(logger *log.Logger, ctx context.Context, serviceKey string, rev int64, watchChan chan watchResponse) {
	// rev is the revision of the last modification made to the key. The watcher will
	// start watching for modifications done after this revision. This will prevent
	// packet or modification lost.

	// We always watch from rev+1 (i.e., strictly after the known write)
	startRev := rev + 1

	for {
		rch := watcherV3().Watch(ctx, serviceKey, etcdv3.WithRev(startRev))
		for resp := range rch {
			if resp.Canceled {
				// We've lost the connection to etcd or the context was canceled.
				if errors.Is(resp.Err(), context.Canceled) || errors.Is(ctx.Err(), context.Canceled) {
					return
				}
				logger.Printf("lost watcher of '%v': '%v' (v2: %v)", serviceKey, resp.Err(), clientV3().Endpoints())
				time.Sleep(1 * time.Second)
				// Restart from last known startRev
				break
			}

			for _, ev := range resp.Events {
				// We only care about PUT updates on the single key
				if ev.Kv == nil || ev.Type != etcdv3.EventTypePut {
					continue
				}
				var serviceInfos service
				if err := json.Unmarshal(ev.Kv.Value, &serviceInfos); err != nil {
					logger.Printf("error while getting service key '%v': '%v' (%v)", serviceKey, err, clientV3().Endpoints())
					time.Sleep(1 * time.Second)
					continue
				}
				// We've got the modification, send it to the register agent
				startRev = ev.Kv.ModRevision + 1
				watchChan <- watchResponse{
					User:     serviceInfos.User,
					Password: serviceInfos.Password,
				}
			}
		}

		// Channel closed not due to context cancel: short backoff and retry.
		if errors.Is(ctx.Err(), context.Canceled) {
			return
		}
		time.Sleep(1 * time.Second)
	}
}

// service is a partial copy of `service.go` to unmarshal the values stored in etcd.
type service struct {
	User     string `json:"user,omitempty"`     // The service username
	Password string `json:"password,omitempty"` // The service password
}

func ListValuesForService(ctx context.Context, name string) ([][]byte, error) {
	resv2, err := KAPI().Get(ctx, "/services/"+name, &etcd.GetOptions{
		Recursive: true,
	})
	if err != nil && !etcd.IsKeyNotFound(err) {
		return nil, errgo.Notef(err, "Unable to fetch v2 services")
	}

	resv3, err := KAPIV3().Get(ctx, "/services/"+name)
	if err != nil {
		return nil, errgo.Notef(err, "Unable to fetch v3 services")
	}

	res := [][]byte{}
	for _, node := range resv2.Node.Nodes {
		res = append(res, []byte(node.Value))
	}

	for _, node := range resv3.Kvs {
		res = append(res, node.Value)
	}

	return res, nil
}
