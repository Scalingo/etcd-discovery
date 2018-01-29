package service

import (
	etcd "github.com/coreos/etcd/client"
	"golang.org/x/net/context"
	errgo "gopkg.in/errgo.v1"
)

// ServiceResponse is the interface used to provide a response to the service.Get() Method.
// This interface provide a standard API used for method chaining like:
// 	url, err := Get("my-service").First().URL()
//
// To provide such API, go errors need to be stored and sent at the last moment.
// To do so, each "final" method (like Url or All), will check if the Response is errored, before
// continuing to their own logic.
type ServiceResponse interface {
	// Err is the method used to check if the Response is errored.
	Err() error
	// Service return the Service struct representing the requested service
	Service() (*Service, error)
	// One return a host of the service choosen randomly
	One() HostResponse
	// First return the first host of the serice
	First() HostResponse
	// All return all the hosts registred for this service
	All() (Hosts, error)
	// URL returns a valid url for this service
	URL(scheme, path string) (string, error)
}

// Get a service by its name. This method does not directly return the Service, but a ServiceResponse. This permit method chaining like:
// 	url, err := Get("my-service").First().URL()
//
// If there was an error during the acquisition of the service, this error will be stored in the ServiceResponse. Final methods will check for this error before doing actual logic.
// If the service is not found, we won't render an error, but will return a service with minimal informations. This is done to provide maximal backwerd compatibility since older versions does not register themself to the "/services_infos" directory.
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
		}
		return &GetServiceResponse{
			err:     errgo.Mask(err),
			service: nil,
		}
	}

	s, err := buildServiceFromNode(res.Node)
	if err != nil {
		return &GetServiceResponse{
			err:     errgo.Mask(err),
			service: nil,
		}
	}
	return &GetServiceResponse{
		err:     nil,
		service: s,
	}
}

// GetServiceResponse is the implementation of the ServiceResponse interface used by the Get method
// This only provide the error wrapping logic, all the actual logic for thsese mêthod are done by the Service struct.
type GetServiceResponse struct {
	service *Service
	err     error
}

// Err wil return an error if an error happend when you've called tre Get method, or nil if no error were detected.
func (q *GetServiceResponse) Err() error {
	return q.err
}

// Service will return the service returned by the Get method. If the service was not found, no error will be return but the service will only contains a Name field.
func (q *GetServiceResponse) Service() (*Service, error) {
	if q.err != nil {
		return nil, errgo.Mask(q.err)
	}
	return q.service, nil
}

// All will return a slice of all the hosts registred to the service
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

// One will return a host choosen randomly in all the hosts of the service
// If the ServiceResponse is errored, the errors will be passed to the HostResponse
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

// First will return the first host registred to the service
// If the ServiceResponse is errored, the errors will be passed to the HostResponse
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

// URL build url for the specified service. If the service is not public, a random host will be choosen and an url will be generated.
func (q *GetServiceResponse) URL(scheme, path string) (string, error) {
	if q.err != nil {
		return "", errgo.Mask(q.err)
	}

	url, err := q.service.URL(scheme, path)
	if err != nil {
		return "", errgo.Mask(err)
	}
	return url, nil
}
