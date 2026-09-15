package service

import (
	"context"
	stderrors "errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetNoHost(t *testing.T) {
	t.Run("Without any service, Get should return ErrNoHostFound and a nil slice", func(t *testing.T) {
		hosts, err := Get(t.Context(), "test_no_service").All(t.Context())
		require.EqualError(t, err, ErrNoHostFound.Error())
		assert.Nil(t, hosts)
	})

	t.Run("Without any service, Get().One().Host() should return ErrNoHostFound", func(t *testing.T) {
		host, err := Get(t.Context(), "test_no_service").One(t.Context()).Host(t.Context())
		require.ErrorIs(t, err, ErrNoHostFound)
		assert.Nil(t, host)
	})
}

func TestGet(t *testing.T) {
	t.Run("With registered services, we should have 2 hosts", func(t *testing.T) {
		ctx1, cancel1 := context.WithCancel(t.Context())
		defer cancel1()
		ctx2, cancel2 := context.WithCancel(t.Context())
		defer cancel2()

		host1 := genHost("host1")
		host2 := genHost("host2")
		host1.Name = "test_service_get"
		host2.Name = "test_service_get"
		w1 := Register(ctx1, "test_service_get", host1)
		w2 := Register(ctx2, "test_service_get", host2)
		require.NoError(t, w1.WaitRegistration(t.Context()))
		require.NoError(t, w2.WaitRegistration(t.Context()))

		hosts, err := Get(t.Context(), "test_service_get").All(t.Context())
		require.NoError(t, err)
		assert.Len(t, hosts, 2)

		if hosts[0].UUID == w1.UUID() {
			host1.UUID = hosts[0].UUID
			host2.UUID = hosts[1].UUID
			assert.Equal(t, host1, *hosts[0])
			assert.Equal(t, host2, *hosts[1])
		} else {
			host1.UUID = hosts[1].UUID
			host2.UUID = hosts[0].UUID
			assert.Equal(t, host1, *hosts[1])
			assert.Equal(t, host2, *hosts[0])
		}
	})
}

func TestGetServiceResponse(t *testing.T) {
	t.Run("With an errored Response", func(t *testing.T) {
		testErrorMsg := "my test error"
		response := &GetServiceResponse{
			err:     stderrors.New(testErrorMsg),
			service: nil,
		}

		t.Run("Err should return an error", func(t *testing.T) {
			err := response.Err()
			require.EqualError(t, err, testErrorMsg)
		})

		t.Run("Service should return an error", func(t *testing.T) {
			service, err := response.Service(t.Context())
			require.EqualError(t, err, testErrorMsg)
			assert.Nil(t, service)
		})

		t.Run("All should return an error", func(t *testing.T) {
			h, err := response.All(t.Context())
			require.EqualError(t, err, testErrorMsg)
			assert.Empty(t, h)
		})

		t.Run("URL should return an error", func(t *testing.T) {
			url, err := response.URL(t.Context(), "http", "/path")
			require.EqualError(t, err, testErrorMsg)
			assert.Empty(t, url)
		})

		t.Run("One should return an errored host response", func(t *testing.T) {
			response := response.One(t.Context())
			require.NotNil(t, response)
			require.EqualError(t, response.Err(), testErrorMsg)
		})

		t.Run("First should return an errored host response", func(t *testing.T) {
			response := response.First(t.Context())
			require.NotNil(t, response)
			require.EqualError(t, response.Err(), testErrorMsg)
		})
	})

	t.Run("With a valid response", func(t *testing.T) {
		expectedService := genService("test-service-11122233444")
		response := &GetServiceResponse{
			err:     nil,
			service: expectedService,
		}

		t.Run("Err should be nil", func(t *testing.T) {
			require.NoError(t, response.Err())
		})

		t.Run("Service should respond a valid service", func(t *testing.T) {
			service, err := response.Service(t.Context())
			require.NoError(t, err)
			assert.Equal(t, expectedService, service)
		})

		t.Run("All should return ErrNoHostFound when the backing service has no hosts key", func(t *testing.T) {
			hosts, err := response.All(t.Context())
			require.EqualError(t, err, ErrNoHostFound.Error())
			assert.Nil(t, hosts)
		})

		t.Run("URL should return a valid url", func(t *testing.T) {
			url, err := response.URL(t.Context(), "http", "/path")
			require.NoError(t, err)
			assert.Equal(t, "http://user:password@public.dev:80/path", url)
		})

		t.Run("One should pass the One error", func(t *testing.T) {
			r := response.One(t.Context())
			require.ErrorIs(t, r.Err(), ErrNoHostFound)
		})

		t.Run("First should pass the First error", func(t *testing.T) {
			r := response.First(t.Context())
			require.ErrorIs(t, r.Err(), ErrNoHostFound)
		})
	})
}

func TestGetForShard(t *testing.T) {
	const testShardID = "shard-0"

	t.Run("With two shards, it should only return hosts from requested shard", func(t *testing.T) {
		ctx1, cancel1 := context.WithCancel(t.Context())
		defer cancel1()
		ctx2, cancel2 := context.WithCancel(t.Context())
		defer cancel2()

		hostShard0 := genHost("host-shard-0")
		hostShard0.Shard = testShardID
		hostShard0.Hostname = "host-shard-0.dev"

		hostShard1 := genHost("host-shard-1")
		hostShard1.Shard = testShard1ID
		hostShard1.Hostname = "host-shard-1.dev"

		w1 := Register(ctx1, "test_service_get_for_shard", hostShard0)
		w2 := Register(ctx2, "test_service_get_for_shard", hostShard1)
		require.NoError(t, w1.WaitRegistration(t.Context()))
		require.NoError(t, w2.WaitRegistration(t.Context()))

		hosts, err := GetForShard(t.Context(), "test_service_get_for_shard", testShardID).All(t.Context())
		require.NoError(t, err)
		require.Len(t, hosts, 1)
		assert.Equal(t, testShardID, hosts[0].Shard)

		host, err := GetForShard(t.Context(), "test_service_get_for_shard", testShard1ID).First(t.Context()).Host(t.Context())
		require.NoError(t, err)
		assert.Equal(t, testShard1ID, host.Shard)
		assert.Equal(t, "host-shard-1.dev", host.Hostname)

		url, err := GetForShard(t.Context(), "test_service_get_for_shard", testShard1ID).URL(t.Context(), "http", "/path")
		require.NoError(t, err)
		assert.Equal(t, "http://user:password@host-shard-1.dev:10000/path", url)
	})

	t.Run("When no host matches the shard, it should return shard-specific no-host errors", func(t *testing.T) {
		host := genHost("host-no-match")
		host.Shard = testShardID
		w := Register(t.Context(), "test_service_get_for_shard_no_match", host)
		require.NoError(t, w.WaitRegistration(t.Context()))

		hosts, err := GetForShard(t.Context(), "test_service_get_for_shard_no_match", testShard1ID).All(t.Context())
		require.EqualError(t, err, ErrNoHostFoundOnShard.Error())
		assert.Nil(t, hosts)

		emptyHost, err := GetForShard(t.Context(), "test_service_get_for_shard_no_match", testShard1ID).First(t.Context()).Host(t.Context())
		require.ErrorIs(t, err, ErrNoHostFoundOnShard)
		assert.Nil(t, emptyHost)

		oneHost, err := GetForShard(t.Context(), "test_service_get_for_shard_no_match", testShard1ID).One(t.Context()).Host(t.Context())
		require.ErrorIs(t, err, ErrNoHostFoundOnShard)
		assert.Nil(t, oneHost)

		url, err := GetForShard(t.Context(), "test_service_get_for_shard_no_match", testShard1ID).URL(t.Context(), "http", "/path")
		require.ErrorIs(t, err, ErrNoHostFoundOnShard)
		assert.Empty(t, url)
	})
}
