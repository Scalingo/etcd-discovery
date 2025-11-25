package service

import (
	"encoding/json"

	etcdv2 "go.etcd.io/etcd/client/v2"
	"gopkg.in/errgo.v1"
)

func buildHostsFromNodes(nodes etcdv2.Nodes) (Hosts, error) {
	hosts := make(Hosts, len(nodes))
	for i, node := range nodes {
		host, err := buildHostFromNode(node)
		if err != nil {
			return nil, errgo.Mask(err)
		}
		hosts[i] = host
	}
	return hosts, nil
}

func buildHostFromNode(node *etcdv2.Node) (*Host, error) {
	host := &Host{}
	err := json.Unmarshal([]byte(node.Value), &host)
	if err != nil {
		return nil, errgo.Notef(err, "Unable to unmarshal host")
	}
	return host, nil
}

func buildServiceFromNode(node *etcdv2.Node) (*Service, error) {
	service := &Service{}
	err := json.Unmarshal([]byte(node.Value), service)
	if err != nil {
		return nil, errgo.Notef(err, "Unable to unmarshal service")
	}
	return service, nil
}
