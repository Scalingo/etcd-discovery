package service

import (
	"context"

	"github.com/Scalingo/etcd-discovery/v9/service/etcdwrapper"

	etcdv2 "go.etcd.io/etcd/client/v2"
)

// ServiceResponse is the interface used to provide a response to the service.Get()
// and service.GetForShard() Methods.
//
// This interface provides a standard API used for method chaining like:
//
//	url, err := Get(ctx, "my-service").First(ctx).URL(ctx, "http", "/")
//	url, err := GetForShard(ctx, "my-service", "shard-0").First(ctx).URL(ctx, "http", "/")
//
// To provide such API, go errors need to be stored and sent at the last moment.
// To do so, each "final" method (like URL or All) will check if the Response is errored,
// before continuing to their own logic.
type ServiceResponse interface {
	// Err is the method used to check if the Response is errored.
	Err() error
	// Service returns the Service struct representing the requested service
	Service(ctx context.Context) (*Service, error)
	// One return a host of the service chosen randomly
	One(ctx context.Context) HostResponse
	// First return the first host of the service
	First(ctx context.Context) HostResponse
	// All returns all the hosts registered for this service
	All(ctx context.Context) (Hosts, error)
	// URL returns a valid url for this service
	URL(ctx context.Context, scheme, path string) (string, error)
}

// Get a service by its name. This method does not directly return the Service, but a ServiceResponse.
//
// This permits method chaining like:
//
//	url, err := Get(ctx, "my-service").First(ctx).URL(ctx, "http", "/")
//
// If there was an error during the acquisition of the service, this error will be stored in the
// ServiceResponse. Final methods will check for this error before doing actual logic.
//
// If the service is not found, we won't render an error but will return a service with minimal
// information. This is done to provide maximal backward compatibility since older versions do
// not register themselves to the "/services_infos" directory.
func Get(ctx context.Context, service string) ServiceResponse {
	res, err := etcdwrapper.KAPI().Get(ctx, "/services_infos/"+service, nil)
	if err != nil && !etcdv2.IsKeyNotFound(err) {
		return &GetServiceResponse{
			err:     err,
			service: nil,
		}
	} else if etcdv2.IsKeyNotFound(err) {
		res, err := etcdwrapper.KAPIV3().Get(ctx, "/services_infos/"+service)
		if err != nil {
			return &GetServiceResponse{
				err:     err,
				service: nil,
			}
		}

		if len(res.Kvs) == 0 {
			return &GetServiceResponse{
				err: nil,
				service: &Service{
					Name: service,
				},
			}
		}

		s, err := buildServiceFromNode(ctx, res.Kvs[0].Value)
		if err != nil {
			return &GetServiceResponse{
				err:     err,
				service: nil,
			}
		}
		return &GetServiceResponse{
			err:     nil,
			service: s,
		}
	}

	s, err := buildServiceFromNode(ctx, []byte(res.Node.Value))
	if err != nil {
		return &GetServiceResponse{
			err:     err,
			service: nil,
		}
	}
	return &GetServiceResponse{
		err:     nil,
		service: s,
	}
}

// GetForShard is similar to Get, but all host-based operations are filtered on the provided shard.
func GetForShard(ctx context.Context, serviceName, shard string) ServiceResponse {
	res := Get(ctx, serviceName)
	getRes, ok := res.(*GetServiceResponse)
	if !ok {
		return res
	}
	getRes.shard = shard
	return getRes
}

// GetServiceResponse is the implementation of the ServiceResponse interface used by the Get method.
//
// This only provides the error wrapping logic, the Service struct does all the actual logic for these methods.
type GetServiceResponse struct {
	service *Service
	err     error
	shard   string
}

// Err will return an error if an error happened when you've called the Get method,
// or nil if no error were detected.
func (q *GetServiceResponse) Err() error {
	return q.err
}

// Service will return the service returned by the Get method.
//
// If the service was not found, no error will be returned,
// but the service will only contain a Name field.
func (q *GetServiceResponse) Service(ctx context.Context) (*Service, error) {
	if q.err != nil {
		return nil, q.err
	}
	return q.service, nil
}

// All will return a slice of all the hosts registered to the service
func (q *GetServiceResponse) All(ctx context.Context) (Hosts, error) {
	if q.err != nil {
		return nil, q.err
	}

	opts := QueryOptions{Shard: q.shard}
	hosts, err := q.service.All(ctx, opts)
	if err != nil {
		return nil, err
	}

	return hosts, nil
}

// One will return a host chosen randomly in all the hosts of the service.
//
// If the ServiceResponse is errored, the errors will be passed to the HostResponse.
func (q *GetServiceResponse) One(ctx context.Context) HostResponse {
	if q.err != nil {
		return &GetHostResponse{
			err:  q.err,
			host: nil,
		}
	}

	opts := QueryOptions{Shard: q.shard}
	host, err := q.service.One(ctx, opts)
	if err != nil {
		return &GetHostResponse{
			err:  err,
			host: nil,
		}
	}
	return &GetHostResponse{
		err:  nil,
		host: host,
	}
}

// First will return the first host registered to the service.
//
// If the ServiceResponse is errored, the errors will be passed to the HostResponse.
func (q *GetServiceResponse) First(ctx context.Context) HostResponse {
	if q.err != nil {
		return &GetHostResponse{
			err:  q.err,
			host: nil,
		}
	}

	opts := QueryOptions{Shard: q.shard}
	host, err := q.service.First(ctx, opts)
	if err != nil {
		return &GetHostResponse{
			err:  err,
			host: nil,
		}
	}
	return &GetHostResponse{
		host: host,
		err:  nil,
	}
}

// URL build url for the specified service. If the service is not public,
// a random host will be chosen and a url will be generated.
func (q *GetServiceResponse) URL(ctx context.Context, scheme, path string) (string, error) {
	if q.err != nil {
		return "", q.err
	}

	opts := QueryOptions{Shard: q.shard}
	url, err := q.service.URL(ctx, scheme, path, opts)
	if err != nil {
		return "", err
	}
	return url, nil
}
