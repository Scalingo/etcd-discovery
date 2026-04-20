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

type fakeWatcher struct {
	results []resAndErr
	index   int
}

func (w *fakeWatcher) Next(context.Context) (*etcdv2.Response, error) {
	if w.index >= len(w.results) {
		return nil, context.Canceled
	}

	result := w.results[w.index]
	w.index++
	return result.Response, result.error
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

func TestSubscribeDownClosesDataChannelWhenErrUnread(t *testing.T) {
	subscribeWatcher = func(string) etcdv2.Watcher {
		return &fakeWatcher{
			results: []resAndErr{
				{error: &etcdv2.Error{Code: 500, Message: "boom"}},
			},
		}
	}
	t.Cleanup(func() {
		subscribeWatcher = Subscribe
	})

	hosts, errs := SubscribeDown(t.Context(), "test_expiration")

	select {
	case host, ok := <-hosts:
		assert.False(t, ok)
		assert.Empty(t, host)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for hosts channel to close")
	}

	err, ok := <-errs
	require.True(t, ok)
	require.Equal(t, 500, err.Code)
	require.Equal(t, "boom", err.Message)
}

func TestSubscribeDown(t *testing.T) {
	registrationCtx, cancelRegistration := context.WithCancel(t.Context())
	subscriptionCtx, cancelSubscription := context.WithCancel(t.Context())
	defer cancelSubscription()

	t.Run("When the service 'test' is watched and a host expired", func(t *testing.T) {
		w := Register(registrationCtx, "test_expiration", genHost("test-expiration"))
		hosts, _ := SubscribeDown(subscriptionCtx, "test_expiration")
		require.NoError(t, w.WaitRegistration(t.Context()))

		cancelRegistration()
		t.Run("The name of the disappeared host should be returned", func(t *testing.T) {
			select {
			case host, ok := <-hosts:
				assert.True(t, ok)
				assert.Equal(t, host, w.UUID())
			}
		})
	})
}

func TestSubscribeNewClosesDataChannelWhenErrUnread(t *testing.T) {
	subscribeWatcher = func(string) etcdv2.Watcher {
		return &fakeWatcher{
			results: []resAndErr{
				{error: &etcdv2.Error{Code: 500, Message: "boom"}},
			},
		}
	}
	t.Cleanup(func() {
		subscribeWatcher = Subscribe
	})

	hosts, errs := SubscribeNew(t.Context(), "test_new")

	select {
	case host, ok := <-hosts:
		assert.False(t, ok)
		assert.Nil(t, host)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for hosts channel to close")
	}

	err, ok := <-errs
	require.True(t, ok)
	require.Equal(t, 500, err.Code)
	require.Equal(t, "boom", err.Message)
}

func TestSubscribeNew(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	t.Run("When the service 'test' is watched and a host registered", func(t *testing.T) {
		hosts, _ := SubscribeNew(t.Context(), "test_new")
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
