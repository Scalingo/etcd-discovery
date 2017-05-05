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
	res, err := KAPI().Get(context.Background(), "/services/"+service, &etcd.GetOptions{Recursive: true})
	if err != nil {
		if etcd.IsKeyNotFound(err) {
			return &Service{}, nil
		}
		return nil, err
	}

	return buildServiceFromNodes(res.Node.Nodes), nil
}
