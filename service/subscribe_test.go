package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	etcdv2 "go.etcd.io/etcd/client/v2"
)

type resAndErr struct {
	Response *etcdv2.Response
	error    error
}

func TestSubscribe(t *testing.T) {
	t.Run("When we subscribe a service, we get all the notifications from it", func(t *testing.T) {
		watcher := Subscribe("test_subs")
		t.Run("When something happens about this service, the responses must be gathered in the channel", func(t *testing.T) {
			responsesChan := make(chan resAndErr)
			go func() {
				for {
					r, err := watcher.Next(t.Context())
					responsesChan <- resAndErr{r, err}
				}
			}()

			time.Sleep(100 * time.Millisecond)
			_, err := KAPI().Create(t.Context(), "/services/test_subs/key", "test")
			require.NoError(t, err)

			response := <-responsesChan
			r := response.Response
			err = response.error

			require.NoError(t, err)
			assert.NotNil(t, r)
			assert.Equal(t, "/services/test_subs/key", r.Node.Key)
			assert.Equal(t, "create", r.Action)

			_, err = KAPI().Delete(t.Context(),
				"/services/test_subs/key", &etcdv2.DeleteOptions{},
			)
			require.NoError(t, err)

			response = <-responsesChan
			r = response.Response
			err = response.error

			require.NoError(t, err)
			assert.NotNil(t, r)
			assert.Equal(t, "/services/test_subs/key", r.Node.Key)
			assert.Equal(t, "delete", r.Action)
		})
	})
}

func TestSubscribeDown(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())

	t.Run("When the service 'test' is watched and a host expired", func(t *testing.T) {
		w := Register(ctx, "test_expiration", genHost("test-expiration"))
		hosts, _ := SubscribeDown("test_expiration")
		w.WaitRegistration()

		cancel()
		t.Run("The name of the disappeared host should be returned", func(t *testing.T) {
			select {
			case host, ok := <-hosts:
				assert.True(t, ok)
				assert.Equal(t, host, w.UUID())
			}
		})
	})
}

func TestSubscribeNew(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	t.Run("When the service 'test' is watched and a host registered", func(t *testing.T) {
		hosts, _ := SubscribeNew("test_new")
		time.Sleep(200 * time.Millisecond)
		newHost := genHost("test-new")
		Register(ctx, "test_new", newHost)
		newHost.Name = "test_new"
		t.Run("A host should be available in the channel", func(t *testing.T) {
			host, ok := <-hosts
			assert.True(t, ok)
			newHost.UUID = host.UUID
			assert.Equal(t, *host, newHost)
		})
	})
}
