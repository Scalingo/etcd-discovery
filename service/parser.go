package service

import (
	"context"
	"encoding/json"

	etcdv2 "go.etcd.io/etcd/client/v2"

	"github.com/Scalingo/go-utils/errors/v3"
)

func buildHostsFromNodes(nodes etcdv2.Nodes) (Hosts, error) {
	hosts := make(Hosts, len(nodes))
	for i, node := range nodes {
		host, err := buildHostFromNode(node)
		if err != nil {
			return nil, errors.Wrap(context.Background(), err, "build host from node")
		}
		hosts[i] = host
	}
	return hosts, nil
}

func buildHostFromNode(node *etcdv2.Node) (*Host, error) {
	host := &Host{}
	err := json.Unmarshal([]byte(node.Value), &host)
	if err != nil {
		return nil, errors.Wrap(context.Background(), err, "unmarshal host")
	}
	return host, nil
}

func buildServiceFromNode(node *etcdv2.Node) (*Service, error) {
	service := &Service{}
	err := json.Unmarshal([]byte(node.Value), service)
	if err != nil {
		return nil, errors.Wrap(context.Background(), err, "unmarshal service")
	}
	return service, nil
}
