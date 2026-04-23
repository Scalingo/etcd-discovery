package service

import (
	stderrors "errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHostUrl(t *testing.T) {
	t.Run("With a host without any password", func(t *testing.T) {
		host := genHost("test")
		host.User = ""
		host.Password = ""

		url, err := host.URL(t.Context(), "http", "/path")
		require.NoError(t, err)
		assert.Equal(t, "http://public.dev:10000/path", url)
	})

	t.Run("With a host with a password", func(t *testing.T) {
		host := genHost("test")
		url, err := host.URL(t.Context(), "http", "/path")
		require.NoError(t, err)
		assert.Equal(t, "http://user:password@public.dev:10000/path", url)
	})

	t.Run("When the port doesn't exists", func(t *testing.T) {
		host := genHost("test")
		url, err := host.URL(t.Context(), "htjp", "/path")
		require.Error(t, err)
		require.ErrorIs(t, err, ErrUnknownScheme)
		assert.Equal(t, 0, len(url))
	})

	t.Run("When the scheme is not provided", func(t *testing.T) {
		host := genHost("test")
		url, err := host.URL(t.Context(), "", "/path")
		require.NoError(t, err)
		assert.Equal(t, "http://user:password@public.dev:10000/path", url)
	})
}

func TestHostPrivateUrl(t *testing.T) {
	t.Run("With a host without any password", func(t *testing.T) {
		host := genHost("test")
		host.User = ""
		host.Password = ""

		url, err := host.PrivateURL(t.Context(), "http", "/path")
		require.NoError(t, err)
		assert.Equal(t, "http://test-private.dev:20000/path", url)
	})

	t.Run("With a host with a password", func(t *testing.T) {
		host := genHost("test")
		url, err := host.PrivateURL(t.Context(), "http", "/path")
		require.NoError(t, err)
		assert.Equal(t, "http://user:password@test-private.dev:20000/path", url)
	})

	t.Run("When the port doesn't exists", func(t *testing.T) {
		host := genHost("test")
		url, err := host.PrivateURL(t.Context(), "htjp", "/path")
		require.Error(t, err)
		require.ErrorIs(t, err, ErrUnknownScheme)
		assert.Equal(t, 0, len(url))
	})

	t.Run("When the scheme is not provided", func(t *testing.T) {
		host := genHost("test")
		url, err := host.PrivateURL(t.Context(), "", "/path")
		require.NoError(t, err)
		assert.Equal(t, "http://user:password@test-private.dev:20000/path", url)
	})

	t.Run("When the host does not support private urls, it should fall back to URL", func(t *testing.T) {
		host := genHost("test")
		host.PrivateHostname = ""
		url, err := host.PrivateURL(t.Context(), "http", "/path")
		require.NoError(t, err)
		assert.Equal(t, "http://user:password@public.dev:10000/path", url)
	})
}

func TestHostsString(t *testing.T) {
	t.Run("With a list of two hosts", func(t *testing.T) {
		host1 := genHost("test")
		host2 := genHost("test")
		host1.PrivateHostname = ""
		hosts := Hosts{&host1, &host2}
		assert.Equal(t, "public.dev, test-private.dev", hosts.String())
	})

	t.Run("With an empty list", func(t *testing.T) {
		hosts := Hosts{}
		assert.Equal(t, "", hosts.String())
	})
}

func TestGetHostResponse(t *testing.T) {
	t.Run("With an errored response", func(t *testing.T) {
		response := &GetHostResponse{
			err:  stderrors.New("TestError"),
			host: nil,
		}

		t.Run("The err method should return an error", func(t *testing.T) {
			require.Error(t, response.Err())
			assert.Equal(t, "TestError", response.Err().Error())
		})

		t.Run("The Host method should return an error", func(t *testing.T) {
			host, err := response.Host(t.Context())
			require.Error(t, err)
			assert.Equal(t, "TestError", err.Error())
			assert.Nil(t, host)
		})

		t.Run("The URL method should return an error", func(t *testing.T) {
			url, err := response.URL(t.Context(), "http", "/path")
			require.Error(t, err)
			assert.Equal(t, "TestError", err.Error())
			assert.Equal(t, "", url)
		})

		t.Run("The PrivateURL should return an error", func(t *testing.T) {
			url, err := response.PrivateURL(t.Context(), "http", "/path")
			require.Error(t, err)
			assert.Equal(t, "TestError", err.Error())
			assert.Equal(t, "", url)
		})
	})

	t.Run("With a valid response", func(t *testing.T) {
		host := genHost("test-service")
		response := &GetHostResponse{
			err:  nil,
			host: &host,
		}

		t.Run("The err method should not return an error", func(t *testing.T) {
			require.NoError(t, response.Err())
		})

		t.Run("The Host method should return a valid host", func(t *testing.T) {
			h, err := response.Host(t.Context())
			require.NoError(t, err)
			assert.Equal(t, &host, h)
		})

		t.Run("The URL method should return a valid url", func(t *testing.T) {
			url, err := response.URL(t.Context(), "http", "/path")
			require.NoError(t, err)
			assert.Equal(t, "http://user:password@public.dev:10000/path", url)
		})

		t.Run("The Private URL should return a valid url", func(t *testing.T) {
			url, err := response.PrivateURL(t.Context(), "http", "/path")
			require.NoError(t, err)
			assert.Equal(t, "http://user:password@test-service-private.dev:20000/path", url)
		})
	})
}
