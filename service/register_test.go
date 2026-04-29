package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	etcdv2 "go.etcd.io/etcd/client/v2"

	"github.com/Scalingo/etcd-discovery/v9/service/etcdwrapper"
)

func TestRegister(t *testing.T) {
	t.Run("After registering service test", func(t *testing.T) {
		host := genHost("test-register")
		t.Run("It should be available with etcd", func(t *testing.T) {
			host.Name = "test_register"
			w := Register(t.Context(), "test_register", host)
			require.NoError(t, w.WaitRegistration(t.Context()))
			uuid := w.UUID()
			res, err := etcdwrapper.KAPI().Get(t.Context(), "/services/test_register/"+uuid, &etcdv2.GetOptions{})
			require.NoError(t, err)

			h := &Host{}
			err = json.Unmarshal([]byte(res.Node.Value), h)
			require.NoError(t, err)

			assert.Equal(t, uuid, path.Base(res.Node.Key))
			host.UUID = h.UUID
			assert.Equal(t, host, *h)
		})

		t.Run(fmt.Sprintf("And the ttl must be <= %s", etcdwrapper.HeartbeatDuration), func(t *testing.T) {
			w := Register(t.Context(), "test2_register", host)
			require.NoError(t, w.WaitRegistration(t.Context()))
			uuid := w.UUID()
			res, err := etcdwrapper.KAPI().Get(t.Context(),
				"/services/test2_register/"+uuid, &etcdv2.GetOptions{},
			)
			require.NoError(t, err)

			now := time.Now()
			duration := res.Node.Expiration.Sub(now)
			assert.LessOrEqual(t, duration, etcdwrapper.HeartbeatDuration*time.Second)
		})

		t.Run("And the serivce infos must be set", func(t *testing.T) {
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
			res, err := etcdwrapper.KAPI().Get(t.Context(), "/services_infos/test3_register", &etcdv2.GetOptions{})
			require.NoError(t, err)

			service := &Service{}
			err = json.Unmarshal([]byte(res.Node.Value), &service)
			require.NoError(t, err)
			assert.Equal(t, infos, service)
		})

		t.Run("And the shard must stay on the host when provided", func(t *testing.T) {
			hostWithShard := genHost("test-shard")
			hostWithShard.Shard = "shard-0"

			w := Register(t.Context(), "test5_register", hostWithShard)
			require.NoError(t, w.WaitRegistration(t.Context()))

			res, err := etcdwrapper.KAPI().Get(t.Context(), "/services_infos/test5_register", &etcdv2.GetOptions{})
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

			resService, err := etcdwrapper.KAPI().Get(t.Context(), "/services_infos/test6_register", &etcdv2.GetOptions{})
			require.NoError(t, err)
			assert.NotContains(t, resService.Node.Value, `"shard"`)

			service := &Service{}
			err = json.Unmarshal([]byte(resService.Node.Value), service)
			require.NoError(t, err)
			assert.Equal(t, "test6_register", service.Name)

			resHost, err := etcdwrapper.KAPI().Get(t.Context(), "/services/test6_register/"+w.UUID(), &etcdv2.GetOptions{})
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
				_, err := etcdwrapper.KAPI().Get(t.Context(), hostKey, &etcdv2.GetOptions{})
				return etcdv2.IsKeyNotFound(err)
			}, etcdwrapper.HeartbeatDuration+2*time.Second, 100*time.Millisecond)
		})

		t.Run("When the privatehostname is not set, it must take the node hostname", func(t *testing.T) {
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
				res, err := etcdwrapper.KAPI().Get(t.Context(), "/services/test-watcher/"+w.UUID(), &etcdv2.GetOptions{})
				require.NoError(t, err)

				h := &Host{}
				err = json.Unmarshal([]byte(res.Node.Value), &h)
				require.NoError(t, err)
				assert.Equal(t, "host2", h.User)
				assert.Equal(t, "password2", h.Password)
			}
		})
	})
}

func TestWithDefaultRegistrationTimeout(t *testing.T) {
	t.Run("It adds a default deadline when the parent context has none", func(t *testing.T) {
		ctx, cancel := withDefaultRegistrationTimeout(t.Context())
		t.Cleanup(cancel)

		deadline, ok := ctx.Deadline()
		require.True(t, ok)
		assert.WithinDuration(t, time.Now().Add(defaultRegistrationTimeout), deadline, 2*time.Second)
	})

	t.Run("It keeps the parent deadline when one already exists", func(t *testing.T) {
		parent, parentCancel := context.WithTimeout(t.Context(), 30*time.Second)
		t.Cleanup(parentCancel)

		ctx, cancel := withDefaultRegistrationTimeout(parent)
		t.Cleanup(cancel)

		parentDeadline, ok := parent.Deadline()
		require.True(t, ok)
		deadline, ok := ctx.Deadline()
		require.True(t, ok)
		assert.Equal(t, parentDeadline, deadline)
	})
}

func TestEnsureInitialHostRegistration(t *testing.T) {
	t.Run("It registers the host with the heartbeat TTL", func(t *testing.T) {
		called := false
		useFakeEtcdServer(t, func(w http.ResponseWriter, r *http.Request) {
			called = true

			assert.Equal(t, http.MethodPut, r.Method)
			assert.Equal(t, "/v2/keys/services/test-initial/host-1", r.URL.Path)
			assert.NoError(t, r.ParseForm())
			assert.Equal(t, "{}", r.Form.Get("value"))
			assert.Equal(t, "5", r.Form.Get("ttl"))

			w.Header().Set("Content-Type", "application/json")
			_, err := w.Write([]byte(`{"action":"set","node":{"key":"/services/test-initial/host-1","value":"{}","createdIndex":1,"modifiedIndex":1}}`))
			assert.NoError(t, err)
		})

		err := ensureInitialHostRegistration(
			t.Context(),
			"test-initial",
			"/services/test-initial/host-1",
			"{}",
			false,
		)

		require.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("It honors the parent context deadline", func(t *testing.T) {
		useFakeEtcdServer(t, func(w http.ResponseWriter, _ *http.Request) {
			writeEtcdError(t, w)
		})

		ctx, cancel := context.WithTimeout(t.Context(), 20*time.Millisecond)
		defer cancel()

		err := ensureInitialHostRegistration(
			ctx,
			"test-initial-timeout",
			"/services/test-initial-timeout/host-1",
			"{}",
			false,
		)

		require.ErrorIs(t, err, context.DeadlineExceeded)
	})
}

func TestEnsureHostRegistrationWaitsForCallerContext(t *testing.T) {
	firstRequest := make(chan struct{})
	useFakeEtcdServer(t, func(w http.ResponseWriter, _ *http.Request) {
		select {
		case firstRequest <- struct{}{}:
		default:
		}
		writeEtcdError(t, w)
	})

	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)

	go func() {
		done <- ensureHostRegistration(
			ctx,
			"test-heartbeat",
			"/services/test-heartbeat/host-1",
			"{}",
			false,
		)
	}()

	select {
	case <-firstRequest:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for the first host registration attempt")
	}

	select {
	case err := <-done:
		t.Fatalf("ensureHostRegistration returned before caller context was canceled: %v", err)
	case <-time.After(100 * time.Millisecond):
	}

	cancel()

	select {
	case err := <-done:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for ensureHostRegistration to stop after context cancellation")
	}
}

func useFakeEtcdServer(t *testing.T, handler http.HandlerFunc) {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	t.Setenv("ETCD_HOSTS", server.URL)

	previousClient := clientSingleton
	previousOnce := clientSingletonO
	clientSingleton = nil
	clientSingletonO = &sync.Once{}

	t.Cleanup(func() {
		clientSingleton = previousClient
		clientSingletonO = previousOnce
	})
}

func writeEtcdError(t *testing.T, w http.ResponseWriter) {
	t.Helper()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	_, err := w.Write([]byte(`{"errorCode":300,"message":"boom","cause":"test","index":1}`))
	require.NoError(t, err)
}
