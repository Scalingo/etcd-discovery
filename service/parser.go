package service

import (
	"context"
	"encoding/json"

	"github.com/Scalingo/go-utils/errors/v3"
)

func buildHostsFromNodes(ctx context.Context, nodeValues [][]byte) (Hosts, error) {
	hosts := make(Hosts, len(nodeValues))
	for i, node := range nodeValues {
		host, err := buildHostFromNode(ctx, node)
		if err != nil {
			return nil, errors.Wrap(ctx, err, "build host from node")
		}
		hosts[i] = host
	}
	return hosts, nil
}

func buildHostFromNode(ctx context.Context, nodeValue []byte) (*Host, error) {
	host := &Host{}
	err := json.Unmarshal(nodeValue, host)
	if err != nil {
		return nil, errors.Wrap(ctx, err, "unmarshal host")
	}
	return host, nil
}

func buildServiceFromNode(ctx context.Context, val []byte) (*Service, error) {
	service := &Service{}
	err := json.Unmarshal(val, service)
	if err != nil {
		return nil, errors.Wrap(ctx, err, "unmarshal service")
	}
	return service, nil
}
