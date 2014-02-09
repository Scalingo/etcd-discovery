package service

import (
	"github.com/coreos/go-etcd/etcd"
)

func Get(service string) ([]*Host, error) {
	res, err := client.Get("/services/" + service, false, true)
	if err != nil {
		// If the service does not exist
		if etcdErr, ok := err.(*etcd.EtcdError); ok && etcdErr.ErrorCode == 100 {
			return []*Host{}, nil
		}
		return nil, err
	}

	return buildHostsFromNodes(res.Node.Nodes), nil
}
