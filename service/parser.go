package service

import (
	"encoding/json"

	etcd "github.com/coreos/etcd/client"
)

func buildHostsFromNodes(nodes etcd.Nodes) Hosts {
	hosts := make(Hosts, len(nodes))
	for i, node := range nodes {
		hosts[i] = buildHostFromNode(node)
	}
	return hosts
}

func buildHostFromNode(node *etcd.Node) *Host {
	host := &Host{}
	err := json.Unmarshal([]byte(node.Value), &host)
	if err != nil {
		panic(err)
	}
	return host
}
