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

		hosts, err := s.All(QueryOptions{})
		require.EqualError(t, err, ErrNoServiceFound.Error())
		assert.Nil(t, hosts)
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

		hosts, err := s.All(QueryOptions{})
		require.NoError(t, err)
		assert.Len(t, hosts, 2)
		if hosts[0].PrivateHostname == "test1-private.dev" {
			assert.Equal(t, "test2-private.dev", hosts[1].PrivateHostname)
		} else {
			assert.Equal(t, "test1-private.dev", hosts[1].PrivateHostname)
			assert.Equal(t, "test2-private.dev", hosts[0].PrivateHostname)
		}
	})

	t.Run("With shard filter", func(t *testing.T) {
		host1 := genHost("test-service-all-shard-1")
		host1.Shard = testShard1ID
		w1 := Register(t.Context(), "test-service-all-shard", host1)
		w1.WaitRegistration()

		host2 := genHost("test-service-all-shard-2")
		host2.Shard = testShard2ID
		w2 := Register(t.Context(), "test-service-all-shard", host2)
		w2.WaitRegistration()

		s, err := Get("test-service-all-shard").Service()
		require.NoError(t, err)

		hosts, err := s.All(QueryOptions{Shard: testShard2ID})
		require.NoError(t, err)
		require.Len(t, hosts, 1)
		assert.Equal(t, testShard2ID, hosts[0].Shard)
		assert.Equal(t, host2.PrivateHostname, hosts[0].PrivateHostname)
	})

	t.Run("With shard filter and no matching host", func(t *testing.T) {
		host := genHost("test-service-all-shard-no-match")
		host.Shard = testShard1ID
		w := Register(t.Context(), "test-service-all-shard-no-match", host)
		w.WaitRegistration()

		s, err := Get("test-service-all-shard-no-match").Service()
		require.NoError(t, err)

		hosts, err := s.All(QueryOptions{Shard: testShard2ID})
		require.EqualError(t, err, ErrNoHostFoundOnShard.Error())
		assert.Nil(t, hosts)
	})
}

func TestServiceFirst(t *testing.T) {
	t.Run("Without a service host, it should return ErrNoServiceFound", func(t *testing.T) {
		s, err := Get("test-service-first-empty").Service()
		require.NoError(t, err)

		host, err := s.First(QueryOptions{})
		require.EqualError(t, err, "fetch hosts: "+ErrNoServiceFound.Error())
		assert.Nil(t, host)
	})

	t.Run("With a service", func(t *testing.T) {
		host1 := genHost("test1")
		w := Register(t.Context(), "test-truc", host1)
		w.WaitRegistration()

		s, err := Get("test-truc").Service()
		require.NoError(t, err)

		host, err := s.First(QueryOptions{})
		require.NoError(t, err)
		assert.NotNil(t, host)
		assert.Equal(t, host1.PrivateHostname, host.PrivateHostname)
	})

	t.Run("With shard filter", func(t *testing.T) {
		host1 := genHost("test-service-first-shard-1")
		host1.Shard = testShard1ID
		w1 := Register(t.Context(), "test-service-first-shard", host1)
		w1.WaitRegistration()

		host2 := genHost("test-service-first-shard-2")
		host2.Shard = testShard2ID
		w2 := Register(t.Context(), "test-service-first-shard", host2)
		w2.WaitRegistration()

		s, err := Get("test-service-first-shard").Service()
		require.NoError(t, err)

		host, err := s.First(QueryOptions{Shard: testShard2ID})
		require.NoError(t, err)
		require.NotNil(t, host)
		assert.Equal(t, testShard2ID, host.Shard)
		assert.Equal(t, host2.PrivateHostname, host.PrivateHostname)
	})

	t.Run("With shard filter and no matching host, it should return ErrNoHostFoundOnShard", func(t *testing.T) {
		host1 := genHost("test-service-first-shard-no-match")
		host1.Shard = testShard1ID
		w := Register(t.Context(), "test-service-first-shard-no-match", host1)
		w.WaitRegistration()

		s, err := Get("test-service-first-shard-no-match").Service()
		require.NoError(t, err)

		host, err := s.First(QueryOptions{Shard: testShard2ID})
		require.EqualError(t, err, "fetch hosts: "+ErrNoHostFoundOnShard.Error())
		assert.Nil(t, host)
	})
}

func TestServiceOne(t *testing.T) {
	t.Run("Without a service host, it should return ErrNoServiceFound", func(t *testing.T) {
		s, err := Get("test-service-one-empty").Service()
		require.NoError(t, err)

		host, err := s.One(QueryOptions{})
		require.EqualError(t, err, ErrNoServiceFound.Error())
		assert.Nil(t, host)
	})

	t.Run("With a service", func(t *testing.T) {
		host1 := genHost("test1")
		w := Register(t.Context(), "test-truc", host1)
		w.WaitRegistration()

		s, err := Get("test-truc").Service()
		require.NoError(t, err)

		host, err := s.One(QueryOptions{})
		require.NoError(t, err)
		assert.NotNil(t, host)
		assert.Equal(t, host1.PrivateHostname, host.PrivateHostname)
	})

	t.Run("With shard filter", func(t *testing.T) {
		host1 := genHost("test-shard-1")
		host1.Shard = testShard1ID
		w1 := Register(t.Context(), "test-shard-truc", host1)
		w1.WaitRegistration()

		host2 := genHost("test-shard-2")
		host2.Shard = testShard2ID
		w2 := Register(t.Context(), "test-shard-truc", host2)
		w2.WaitRegistration()

		s, err := Get("test-shard-truc").Service()
		require.NoError(t, err)

		host, err := s.One(QueryOptions{Shard: testShard2ID})
		require.NoError(t, err)
		assert.Equal(t, testShard2ID, host.Shard)
	})

	t.Run("With shard filter and no matching host, it should return ErrNoHostFoundOnShard", func(t *testing.T) {
		host1 := genHost("test-shard-no-match")
		host1.Shard = testShard1ID
		w := Register(t.Context(), "test-shard-no-match", host1)
		w.WaitRegistration()

		s, err := Get("test-shard-no-match").Service()
		require.NoError(t, err)

		host, err := s.One(QueryOptions{Shard: testShard2ID})
		require.EqualError(t, err, ErrNoHostFoundOnShard.Error())
		assert.Nil(t, host)
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

			url, err := s.URL("http", "/path", QueryOptions{})
			require.NoError(t, err)
			assert.Equal(t, "http://public.dev:10000/path", url)
		})

		t.Run("With a host with a password", func(t *testing.T) {
			host := genHost("test")
			w := Register(t.Context(), "service-url-3", host)
			w.WaitRegistration()

			s, err := Get("service-url-3").Service()
			require.NoError(t, err)

			url, err := s.URL("http", "/path", QueryOptions{})
			require.NoError(t, err)
			assert.Equal(t, "http://user:password@public.dev:10000/path", url)
		})

		t.Run("When the port does'nt exists", func(t *testing.T) {
			host := genHost("test")
			w := Register(t.Context(), "service-url-4", host)
			w.WaitRegistration()

			s, err := Get("service-url-4").Service()
			require.NoError(t, err)

			url, err := s.URL("htjp", "/path", QueryOptions{})
			require.EqualError(t, err, "unknown scheme")
			assert.Empty(t, url)
		})

		t.Run("With a shard filter, it should resolve the URL from the matching host", func(t *testing.T) {
			host1 := genHost("service-url-shard-1")
			host1.Shard = testShard1ID
			host1.Hostname = "service-url-shard-1.dev"
			w1 := Register(t.Context(), "service-url-shard", host1)
			w1.WaitRegistration()

			host2 := genHost("service-url-shard-2")
			host2.Shard = testShard2ID
			host2.Hostname = "service-url-shard-2.dev"
			w2 := Register(t.Context(), "service-url-shard", host2)
			w2.WaitRegistration()

			s, err := Get("service-url-shard").Service()
			require.NoError(t, err)

			url, err := s.URL("http", "/path", QueryOptions{Shard: testShard2ID})
			require.NoError(t, err)
			assert.Equal(t, "http://user:password@service-url-shard-2.dev:10000/path", url)
		})

		t.Run("With a shard filter and no matching host, it should return ErrNoHostFoundOnShard", func(t *testing.T) {
			host := genHost("service-url-shard-no-match")
			host.Shard = testShard1ID
			host.Hostname = "service-url-shard-no-match.dev"
			w := Register(t.Context(), "service-url-shard-no-match", host)
			w.WaitRegistration()

			s, err := Get("service-url-shard-no-match").Service()
			require.NoError(t, err)

			url, err := s.URL("http", "/path", QueryOptions{Shard: testShard2ID})
			require.EqualError(t, err, ErrNoHostFoundOnShard.Error())
			assert.Empty(t, url)
		})
	})
}
