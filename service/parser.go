package service

import (
	"encoding/json"

	"gopkg.in/errgo.v1"
)

func buildHostsFromNodes(nodeValues [][]byte) (Hosts, error) {
	hosts := make(Hosts, len(nodeValues))
	for i, node := range nodeValues {
		host, err := buildHostFromNode(node)
		if err != nil {
			return nil, errgo.Mask(err)
		}
		hosts[i] = host
	}
	return hosts, nil
}

func buildHostFromNode(nodeValue []byte) (*Host, error) {
	host := &Host{}
	err := json.Unmarshal(nodeValue, &host)
	if err != nil {
		return nil, errgo.Notef(err, "Unable to unmarshal host")
	}
	return host, nil
}

func buildServiceFromNode(val []byte) (*Service, error) {
	service := &Service{}
	err := json.Unmarshal(val, service)
	if err != nil {
		return nil, errgo.Notef(err, "Unable to unmarshal service")
	}
	return service, nil
}
