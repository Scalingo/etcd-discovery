package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceAll(t *testing.T) {
	t.Run("With no services", func(t *testing.T) {
		s, err := Get("service-test-get-1").Service()
		require.NoError(t, err)

		hosts, err := s.All()
		require.NoError(t, err)
		assert.Empty(t, hosts)
	})

	t.Run("With two services", func(t *testing.T) {
		host1 := genHost("test1")
		host2 := genHost("test2")
		w1 := Register(t.Context(), "test-get-222", host1)
		w2 := Register(t.Context(), "test-get-222", host2)

		w1.WaitRegistration()
		w2.WaitRegistration()

		s, err := Get("test-get-222").Service()
		require.NoError(t, err)
		hosts, err := s.All()
		require.NoError(t, err)
		assert.Len(t, hosts, 2)
		if hosts[0].PrivateHostname == "test1-private.dev" {
			assert.Equal(t, "test2-private.dev", hosts[1].PrivateHostname)
		} else {
			assert.Equal(t, "test1-private.dev", hosts[1].PrivateHostname)
			assert.Equal(t, "test2-private.dev", hosts[0].PrivateHostname)
		}
	})
}

func TestServiceFirst(t *testing.T) {
	t.Run("With no services", func(t *testing.T) {
		s, err := Get("service-test-1").Service()
		require.NoError(t, err)
		host, err := s.First()
		require.EqualError(t, err, "no host found for this service")
		assert.Nil(t, host)
	})

	t.Run("With a service", func(t *testing.T) {
		host1 := genHost("test1")
		w := Register(t.Context(), "test-truc", host1)
		w.WaitRegistration()

		s, err := Get("test-truc").Service()
		require.NoError(t, err)

		host, err := s.First()
		require.NoError(t, err)
		assert.NotNil(t, host)
		assert.Equal(t, host1.PrivateHostname, host.PrivateHostname)
	})
}

func TestServiceOne(t *testing.T) {
	t.Run("With no services", func(t *testing.T) {
		s, err := Get("service-test-1").Service()
		require.NoError(t, err)

		host, err := s.One()
		require.EqualError(t, err, "no host found for this service")
		assert.Nil(t, host)
	})

	t.Run("With a service", func(t *testing.T) {
		host1 := genHost("test1")
		w := Register(t.Context(), "test-truc", host1)
		w.WaitRegistration()

		s, err := Get("test-truc").Service()
		require.NoError(t, err)

		host, err := s.One()
		require.NoError(t, err)
		assert.NotNil(t, host)
		assert.Equal(t, host1.PrivateHostname, host.PrivateHostname)
	})
}

func TestServiceUrl(t *testing.T) {
	t.Run("With a public service", func(t *testing.T) {
		t.Run("With a service without any password", func(t *testing.T) {
			host := genHost("test")
			host.User = ""
			host.Password = ""
			w := Register(t.Context(), "service-url-1", host)
			w.WaitRegistration()

			s, err := Get("service-url-1").Service()
			require.NoError(t, err)

			url, err := s.URL("http", "/path")
			require.NoError(t, err)
			assert.Equal(t, "http://public.dev:10000/path", url)
		})

		t.Run("With a host with a password", func(t *testing.T) {
			host := genHost("test")
			w := Register(t.Context(), "service-url-3", host)
			w.WaitRegistration()

			s, err := Get("service-url-3").Service()
			require.NoError(t, err)

			url, err := s.URL("http", "/path")
			require.NoError(t, err)
			assert.Equal(t, "http://user:password@public.dev:10000/path", url)
		})

		t.Run("When the port does'nt exists", func(t *testing.T) {
			host := genHost("test")
			w := Register(t.Context(), "service-url-4", host)
			w.WaitRegistration()

			s, err := Get("service-url-4").Service()
			require.NoError(t, err)

			url, err := s.URL("htjp", "/path")
			require.EqualError(t, err, "unknown scheme")
			assert.Empty(t, url)
		})
	})
}
