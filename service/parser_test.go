package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	etcdv2 "go.etcd.io/etcd/client/v2"
)

var (
	sampleNode = &etcdv2.Node{
		Key: "/services/test/example.org",
		Value: `
		{
			"name": "public.dev",
			"service_name": "test-service",
			"User": "user",
			"password": "password",
			"ports": {
				"http": "10000"
			},
			"public": true,
			"critical": true,
			"private_hostname": "test-private.dev",
			"private_ports": {
				"http": "20000"
			},
			"uuid": "1234"
		}
		`,
	}
	sampleInfoNode = &etcdv2.Node{
		Key: "/services_infos/test",
		Value: `
		{
			"critical": true
		}
		`,
	}
	sampleNodes = etcdv2.Nodes{sampleNode, sampleNode}
)

var (
	sampleResult = genHost("test")
)

func TestBuildHostsFromNodes(t *testing.T) {
	t.Run("Given a sample response with 2 nodes, we got 2 hosts", func(t *testing.T) {
		hosts, err := buildHostsFromNodes(sampleNodes)
		require.NoError(t, err)
		assert.Len(t, hosts, 2)
		assert.Equal(t, sampleResult, *hosts[0])
		assert.Equal(t, sampleResult, *hosts[1])
	})
}

func TestBuildHostFromNode(t *testing.T) {
	t.Run("Given a sample response, we got a filled Host", func(t *testing.T) {
		host, err := buildHostFromNode(sampleNode)
		require.NoError(t, err)
		assert.Equal(t, sampleResult, *host)
	})
}

func TestBuildServiceFromNode(t *testing.T) {
	t.Run("Given a sample response, we got a filled Infos", func(t *testing.T) {
		infos, err := buildServiceFromNode(sampleInfoNode)
		require.NoError(t, err)
		assert.True(t, infos.Critical)
	})
}
