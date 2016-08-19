package service

import (
	etcd "github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

func Get(service string) (Hosts, error) {
	res, err := KAPI().Get(context.Background(), "/services/"+service, &etcd.GetOptions{Recursive: true})
	if err != nil {
		if etcd.IsKeyNotFound(err) {
			return []*Host{}, nil
		}
		return nil, err
	}

	return buildHostsFromNodes(res.Node.Nodes), nil
}
