package service

import (
	"encoding/json"
	"path"

	etcd "github.com/coreos/etcd/client"
)

func buildServiceFromNodes(nodes etcd.Nodes) *Service {
	infos := &Infos{}
	hosts := make(Hosts, 0)
	for _, node := range nodes {
		if path.Base(node.Key) == "service_infos" {
			infos = buildInfosFromNode(node)
		} else {
			hosts = append(hosts, buildHostFromNode(node))
		}
	}
	return &Service{
		Hosts: hosts,
		Infos: infos,
	}
}

func buildHostFromNode(node *etcd.Node) *Host {
	host := &Host{}
	err := json.Unmarshal([]byte(node.Value), &host)
	if err != nil {
		panic(err)
	}
	return host
}

func buildInfosFromNode(node *etcd.Node) *Infos {
	infos := &Infos{}
	err := json.Unmarshal([]byte(node.Value), &infos)
	if err != nil {
		panic(err)
	}
	return infos
}
