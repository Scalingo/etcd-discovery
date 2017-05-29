package service

import (
	"encoding/json"

	errgo "gopkg.in/errgo.v1"

	etcd "github.com/coreos/etcd/client"
)

func buildHostsFromNodes(nodes etcd.Nodes) (Hosts, error) {
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

func buildHostFromNode(node *etcd.Node) (*Host, error) {
	host := &Host{}
	err := json.Unmarshal([]byte(node.Value), &host)
	if err != nil {
		return nil, errgo.Notef(err, "Unable to unmarshal host")
	}
	return host, nil
}

func buildServiceFromNode(node *etcd.Node) (*Service, error) {
	service := &Service{}
	err := json.Unmarshal([]byte(node.Value), service)
	if err != nil {
		return nil, errgo.Notef(err, "Unable to unmarshal service")
	}
	return service, nil
}
