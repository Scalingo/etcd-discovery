package service

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	etcd "go.etcd.io/etcd/client/v2"

	"github.com/Scalingo/etcd-discovery/v8/service/etcdwrapper"
)

type resAndErr struct {
	Response *etcd.Response
	error    error
}

// Tests
func TestSubscribe(t *testing.T) {
	Convey("When we subscribe a service, we get all the notifications from it", t, func() {
		watcher := Subscribe("test_subs")
		Convey("When something happens about this service, the responses must be gathered in the channel", func() {
			responsesChan := make(chan resAndErr)
			go func() {
				for {
					r, err := watcher.Next(context.Background())
					responsesChan <- resAndErr{r, err}
				}
			}()

			time.Sleep(100 * time.Millisecond)
			_, err := etcdwrapper.KAPI().Create(context.Background(), "/services/test_subs/key", "test")
			So(err, ShouldBeNil)

			response := <-responsesChan
			r := response.Response
			err = response.error

			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)
			So(r.Node.Key, ShouldEqual, "/services/test_subs/key")
			So(r.Action, ShouldEqual, "create")

			_, err = etcdwrapper.KAPI().Delete(context.Background(), "/services/test_subs/key", &etcd.DeleteOptions{})
			So(err, ShouldBeNil)

			response = <-responsesChan
			r = response.Response
			err = response.error

			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)
			So(r.Node.Key, ShouldEqual, "/services/test_subs/key")
			So(r.Action, ShouldEqual, "delete")
		})
	})
}

func TestSubscribeDown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	Convey("When the service 'test' is watched and a host expired", t, func() {
		w := Register(ctx, "test_expiration", genHost("test-expiration"))
		hosts, _ := SubscribeDown("test_expiration")
		w.WaitRegistration()

		cancel()
		Convey("The name of the disappeared host should be returned", func() {
			select {
			case host, ok := <-hosts:
				So(ok, ShouldBeTrue)
				So(host, ShouldEqual, w.UUID())
			}
		})
	})
}

func TestSubscribeNew(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	Convey("When the service 'test' is watched and a host registered", t, func() {
		hosts, _ := SubscribeNew("test_new")
		time.Sleep(200 * time.Millisecond)
		newHost := genHost("test-new")
		Register(ctx, "test_new", newHost)
		newHost.Name = "test_new"
		Convey("A host should be available in the channel", func() {
			host, ok := <-hosts
			So(ok, ShouldBeTrue)
			newHost.UUID = host.UUID
			So(host, ShouldResemble, &newHost)
		})
	})
}
