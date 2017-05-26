package service

import (
	"errors"

	etcd "github.com/coreos/etcd/client"
	"golang.org/x/net/context"
	errgo "gopkg.in/errgo.v1"
)

func Get(service string) (*Service, error) {
	res, err := KAPI().Get(context.Background(), "/services_infos/"+service, nil)

	if err != nil {
		if !etcd.IsKeyNotFound(err) {
			return nil, err
		} else {
			return nil, errors.New("Service not found")
		}
	} else {
		service, err := buildServiceFromNode(res.Node)
		if err != nil {
			return nil, errgo.Mask(err)
		}
		return service, nil
	}
}
