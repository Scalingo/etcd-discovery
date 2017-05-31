package service

import (
	etcd "github.com/coreos/etcd/client"
	"golang.org/x/net/context"
	errgo "gopkg.in/errgo.v1"
)

type ServiceResponse interface {
	Err() error
	Service() (*Service, error)
	One() HostResponse
	First() HostResponse
	All() (Hosts, error)
	Url(scheme, path string) (string, error)
}

func Get(service string) ServiceResponse {
	res, err := KAPI().Get(context.Background(), "/services_infos/"+service, nil)

	if err != nil {
		if etcd.IsKeyNotFound(err) {
			return &GetServiceResponse{
				err: nil,
				service: &Service{
					Name: service,
				},
			}
		} else {
			return &GetServiceResponse{
				err:     errgo.Mask(err),
				service: nil,
			}
		}
	} else {
		service, err := buildServiceFromNode(res.Node)
		if err != nil {
			return &GetServiceResponse{
				err:     errgo.Mask(err),
				service: nil,
			}
		}
		return &GetServiceResponse{
			err:     nil,
			service: service,
		}
	}
}

type GetServiceResponse struct {
	service *Service
	err     error
}

func (q *GetServiceResponse) Err() error {
	return q.err
}

func (q *GetServiceResponse) Service() (*Service, error) {
	if q.err != nil {
		return nil, errgo.Mask(q.err)
	}
	return q.service, nil
}

func (q *GetServiceResponse) All() (Hosts, error) {
	if q.err != nil {
		return nil, q.err
	}

	hosts, err := q.service.All()
	if err != nil {
		return nil, errgo.Mask(err)
	}
	return hosts, nil
}

func (q *GetServiceResponse) One() HostResponse {
	if q.err != nil {
		return &GetHostResponse{
			err:  errgo.Mask(q.err),
			host: nil,
		}
	}

	host, err := q.service.One()
	if err != nil {
		return &GetHostResponse{
			err:  errgo.Mask(err),
			host: nil,
		}
	}
	return &GetHostResponse{
		err:  nil,
		host: host,
	}
}

func (q *GetServiceResponse) First() HostResponse {
	if q.err != nil {
		return &GetHostResponse{
			err:  errgo.Mask(q.err),
			host: nil,
		}
	}
	host, err := q.service.First()
	if err != nil {
		return &GetHostResponse{
			err:  errgo.Mask(err),
			host: nil,
		}
	}
	return &GetHostResponse{
		host: host,
		err:  nil,
	}
}

func (q *GetServiceResponse) Url(scheme, path string) (string, error) {
	if q.err != nil {
		return "", errgo.Mask(q.err)
	}

	url, err := q.service.Url(scheme, path)
	if err != nil {
		return "", errgo.Mask(err)
	}
	return url, nil
}
