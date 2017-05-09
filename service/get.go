package service

import (
	etcd "github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

type Service struct {
	Infos *Infos
	Hosts Hosts
}

func Get(service string) (*Service, error) {
	var hosts Hosts
	var infos *Infos

	res, err := KAPI().Get(context.Background(), "/services/"+service, &etcd.GetOptions{Recursive: true})
	if err != nil {
		if !etcd.IsKeyNotFound(err) {
			return nil, err
		}
	} else {
		hosts = buildHostsFromNodes(res.Node.Nodes)
	}

	res, err = KAPI().Get(context.Background(), "/services_infos/"+service, nil)

	if err != nil {
		if !etcd.IsKeyNotFound(err) {
			return nil, err
		}
	} else {
		infos = buildInfosFromNode(res.Node)
	}

	return &Service{
		Hosts: hosts,
		Infos: infos,
	}, nil
}
