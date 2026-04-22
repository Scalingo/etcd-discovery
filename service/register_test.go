package service

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	etcdv2 "go.etcd.io/etcd/client/v2"
)

func TestRegister(t *testing.T) {
	t.Run("After registering service test", func(t *testing.T) {
		host := genHost("test-register")
		t.Run("It should be available with etcd", func(t *testing.T) {
			host.Name = "test_register"
			w := Register(t.Context(), "test_register", host)
			require.NoError(t, w.WaitRegistration(t.Context()))
			uuid := w.UUID()
			res, err := KAPI().Get(t.Context(), "/services/test_register/"+uuid, &etcdv2.GetOptions{})
			require.NoError(t, err)

			h := &Host{}
			err = json.Unmarshal([]byte(res.Node.Value), h)
			require.NoError(t, err)

			assert.Equal(t, uuid, path.Base(res.Node.Key))
			host.UUID = h.UUID
			assert.Equal(t, host, *h)
		})

		t.Run(fmt.Sprintf("And the ttl must be < %d", HEARTBEAT_DURATION), func(t *testing.T) {
			w := Register(t.Context(), "test2_register", host)
			require.NoError(t, w.WaitRegistration(t.Context()))
			uuid := w.UUID()
			res, err := KAPI().Get(t.Context(),
				"/services/test2_register/"+uuid, &etcdv2.GetOptions{},
			)
			require.NoError(t, err)

			now := time.Now()
			duration := res.Node.Expiration.Sub(now)
			assert.LessOrEqual(t, duration, HEARTBEAT_DURATION*time.Second)
		})

		t.Run("And the service infos must be set", func(t *testing.T) {
			infos := &Service{
				Name:     "test3_register",
				Hostname: "public.dev",
				User:     "user",
				Password: "password",
				Ports: Ports{
					"http": "10000",
				},
				Public:   true,
				Critical: true,
			}
			w := Register(t.Context(), "test3_register", host)
			require.NoError(t, w.WaitRegistration(t.Context()))
			res, err := KAPI().Get(t.Context(), "/services_infos/test3_register", &etcdv2.GetOptions{})
			require.NoError(t, err)

			service := &Service{}
			err = json.Unmarshal([]byte(res.Node.Value), service)
			require.NoError(t, err)
			assert.Equal(t, infos, service)
		})

		t.Run("And the shard must stay on the host when provided", func(t *testing.T) {
			hostWithShard := genHost("test-shard")
			hostWithShard.Shard = "shard-0"

			w := Register(t.Context(), "test5_register", hostWithShard)
			require.NoError(t, w.WaitRegistration(t.Context()))

			res, err := KAPI().Get(t.Context(), "/services_infos/test5_register", &etcdv2.GetOptions{})
			require.NoError(t, err)
			assert.NotContains(t, res.Node.Value, `"shard"`)

			service := &Service{}
			err = json.Unmarshal([]byte(res.Node.Value), service)
			require.NoError(t, err)
			assert.Equal(t, "test5_register", service.Name)

			host, err := Get(t.Context(), "test5_register").First(t.Context()).Host(t.Context())
			require.NoError(t, err)
			assert.Equal(t, "shard-0", host.Shard)
		})

		t.Run("And the shard must be empty when not provided", func(t *testing.T) {
			hostWithoutShard := genHost("test-no-shard")
			hostWithoutShard.Shard = ""

			w := Register(t.Context(), "test6_register", hostWithoutShard)
			require.NoError(t, w.WaitRegistration(t.Context()))

			resService, err := KAPI().Get(t.Context(), "/services_infos/test6_register", &etcdv2.GetOptions{})
			require.NoError(t, err)
			assert.NotContains(t, resService.Node.Value, `"shard"`)

			service := &Service{}
			err = json.Unmarshal([]byte(resService.Node.Value), service)
			require.NoError(t, err)
			assert.Equal(t, "test6_register", service.Name)

			resHost, err := KAPI().Get(t.Context(), "/services/test6_register/"+w.UUID(), &etcdv2.GetOptions{})
			require.NoError(t, err)
			assert.NotContains(t, resHost.Node.Value, `"shard"`)

			storedHost := &Host{}
			err = json.Unmarshal([]byte(resHost.Node.Value), storedHost)
			require.NoError(t, err)
			assert.Empty(t, storedHost.Shard)
		})

		t.Run("After cancelling context, the service should disappear", func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			host := genHost("test-disappear")
			w := Register(ctx, "test4_register", host)
			require.NoError(t, w.WaitRegistration(t.Context()))
			hostKey := "/services/test4_register/" + w.UUID()
			cancel()
			assert.Eventually(t, func() bool {
				_, err := KAPI().Get(t.Context(), hostKey, &etcdv2.GetOptions{})
				return etcdv2.IsKeyNotFound(err)
			}, (HEARTBEAT_DURATION+2)*time.Second, 100*time.Millisecond)
		})

		t.Run("When the private_hostname is not set, it must take the node hostname", func(t *testing.T) {
			host := genHost("HelloWorld")
			host.PrivateHostname = ""
			w := Register(t.Context(), "hello_world", host)
			assert.True(t, strings.HasSuffix(w.UUID(), hostname))
		})
		t.Run("When the private ports is not set and the service is private, it should take the public_ports", func(t *testing.T) {
			host := genHost("HelloWorld2")
			host.Public = false
			host.PrivatePorts = Ports{}
			w := Register(t.Context(), "hello_world2", host)
			require.NoError(t, w.WaitRegistration(t.Context()))
			h, err := Get(t.Context(), "hello_world2").First(t.Context()).Host(t.Context())
			require.NoError(t, err)

			assert.Len(t, h.PrivatePorts, 1)
		})
	})
}

func TestWatcher(t *testing.T) {
	t.Run("With two instances of the same service", func(t *testing.T) {
		host1 := genHost("test-watcher-1")
		host2 := genHost("test-watcher-1")

		host1.User = "host1"
		host1.Password = "password1"

		host2.User = "host2"
		host2.Password = "password2"

		w1 := Register(t.Context(), "test-watcher", host1)
		require.NoError(t, w1.WaitRegistration(t.Context()))

		cred1, err := w1.Credentials()
		require.NoError(t, err)
		assert.Equal(t, "host1", cred1.User)
		assert.Equal(t, "password1", cred1.Password)

		w2 := Register(t.Context(), "test-watcher", host2)
		require.NoError(t, w2.WaitRegistration(t.Context()))

		t.Run("it should send the new passwords", func(t *testing.T) {
			cred2, err := w2.Credentials()
			require.NoError(t, err)
			assert.Equal(t, "host2", cred2.User)
			assert.Equal(t, "password2", cred2.Password)

			time.Sleep(1 * time.Second)
			cred1, err := w1.Credentials()
			require.NoError(t, err)
			assert.Equal(t, "host2", cred1.User)
			assert.Equal(t, "password2", cred1.Password)
		})

		t.Run("it should update the host key", func(t *testing.T) {
			for _, w := range []*Registration{w1, w2} {
				res, err := KAPI().Get(t.Context(), "/services/test-watcher/"+w.UUID(), &etcdv2.GetOptions{})
				require.NoError(t, err)

				h := &Host{}
				err = json.Unmarshal([]byte(res.Node.Value), h)
				require.NoError(t, err)
				assert.Equal(t, "host2", h.User)
				assert.Equal(t, "password2", h.Password)
			}
		})
	})
}
