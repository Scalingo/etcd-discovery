package service

import (
	"github.com/coreos/go-etcd/etcd"
	"path"
)

func buildHostsFromNodes(nodes etcd.Nodes) []*Host {
	hosts := make([]*Host, len(nodes))
	for i, node := range nodes {
		hosts[i] = buildHostFromNode(&node)
	}
	return hosts
}

func buildHostFromNode(node *etcd.Node) *Host {
	nodes := node.Nodes
	host := &Host{}
	host.Name = path.Base(node.Key)
	for _, n := range nodes {
		switch(path.Base(n.Key)) {
		case "user":
			host.User = n.Value
		case "password":
			host.Password = n.Value
		case "port":
			host.Port = n.Value
		}
	}
	return host
}
