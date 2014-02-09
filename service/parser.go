package service

import (
	"encoding/json"
	"github.com/coreos/go-etcd/etcd"
)

func buildHostsFromNodes(nodes etcd.Nodes) []*Host {
	hosts := make([]*Host, len(nodes))
	for i, node := range nodes {
		hosts[i] = buildHostFromNode(&node)
	}
	return hosts
}

func buildHostFromNode(node *etcd.Node) *Host {
	host := &Host{}
	json.Unmarshal([]byte(node.Value), &host)
	return host
}
