package service

import (
	"fmt"

	uuid "github.com/nu7hatch/gouuid"
)

const (
	// HEARTBEAT_DURATION time in second between two registration. The host will
	// be deleted if etcd didn't received any new registration in those 5 seocnds
	HEARTBEAT_DURATION = 5
)

// Register a host with a service name and a host description. The last chan is
// a stop method. If something is written on this channel, any goroutines
// launch by this method will stop.
//
// This service will launch two go routines. The first one will maintain the
// registration every 5 seconds and the second one will check if the service
// credentials don't change and notify otherwise
func Register(serviceName string, host Host) (*Registration, error) {
	if !host.Public && len(host.PrivateHostname) == 0 {
		host.PrivateHostname = host.Hostname
	}

	if len(host.PrivateHostname) == 0 {
		host.PrivateHostname = hostname
	}
	host.Name = serviceName

	if len(host.PrivateHostname) != 0 && len(host.PrivatePorts) == 0 {
		host.PrivatePorts = host.Ports
	}

	uuid, _ := uuid.NewV4()

	hostUuid := fmt.Sprintf("%s-%s", uuid.String(), host.PrivateHostname)
	host.UUID = hostUuid

	service := Service{
		Name:     serviceName,
		Critical: host.Critical,
		Public:   host.Public,
	}

	if host.Public {
		service.Hostname = host.Hostname
		service.Ports = host.Ports
		service.Password = host.Password
		service.User = host.User
	}

	registration, err := NewRegistration(host, service)
	if err != nil {
		return nil, fmt.Errorf("fail to create new registration: %v", err)
	}

	return registration, nil
}
