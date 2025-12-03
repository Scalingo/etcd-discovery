package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetNoHost(t *testing.T) {
	t.Run("Without any service, Get should return an empty slice", func(t *testing.T) {
		hosts, err := Get("test_no_service").All()
		require.NoError(t, err)
		assert.Empty(t, hosts)
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
		w1.WaitRegistration()
		w2.WaitRegistration()

		hosts, err := Get("test_service_get").All()
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
			err:     errors.New(testErrorMsg),
			service: nil,
		}

		t.Run("Err should return an error", func(t *testing.T) {
			err := response.Err()
			require.EqualError(t, err, testErrorMsg)
		})

		t.Run("Service should return an error", func(t *testing.T) {
			service, err := response.Service()
			require.EqualError(t, err, testErrorMsg)
			assert.Nil(t, service)
		})

		t.Run("All should return an error", func(t *testing.T) {
			h, err := response.All()
			require.EqualError(t, err, testErrorMsg)
			assert.Empty(t, h)
		})

		t.Run("Url should return an error", func(t *testing.T) {
			url, err := response.URL("http", "/path")
			require.EqualError(t, err, testErrorMsg)
			assert.Empty(t, url)
		})

		t.Run("One should return an errored host response", func(t *testing.T) {
			response := response.One()
			require.NotNil(t, response)
			require.EqualError(t, response.Err(), testErrorMsg)
		})

		t.Run("First should return an errored host response", func(t *testing.T) {
			response := response.First()
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
			service, err := response.Service()
			require.NoError(t, err)
			assert.Equal(t, expectedService, service)
		})

		t.Run("All should return an empty array", func(t *testing.T) {
			hosts, err := response.All()
			require.NoError(t, err)
			assert.Empty(t, hosts)
		})

		t.Run("Url should return a valid url", func(t *testing.T) {
			url, err := response.URL("http", "/path")
			require.NoError(t, err)
			assert.Equal(t, "http://user:password@public.dev:80/path", url)
		})

		t.Run("One should pass the One error", func(t *testing.T) {
			r := response.One()
			require.EqualError(t, r.Err(), "no host found for this service")
		})

		t.Run("First should pass the First error", func(t *testing.T) {
			r := response.First()
			require.EqualError(t, r.Err(), "no host found for this service")
		})
	})
}
