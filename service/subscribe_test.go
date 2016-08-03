package service

import (
	"testing"
	"time"

	etcd "github.com/coreos/etcd/client"
	"golang.org/x/net/context"

	. "github.com/smartystreets/goconvey/convey"
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
			_, err := KAPI().Create(context.Background(), "/services/test_subs/key", "test")
			So(err, ShouldBeNil)

			response := <-responsesChan
			r := response.Response
			err = response.error

			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)
			So(r.Node.Key, ShouldEqual, "/services/test_subs/key")
			So(r.Action, ShouldEqual, "create")

			_, err = KAPI().Delete(context.Background(), "/services/test_subs/key", &etcd.DeleteOptions{})
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
	stop := make(chan struct{})

	Convey("When the service 'test' is watched and a host expired", t, func() {
		Register("test_expiration", genHost("test-expiration"), stop)
		hosts, _ := SubscribeDown("test_expiration")
		close(stop)
		Convey("The name of the disappeared host should be returned", func() {
			select {
			case host, ok := <-hosts:
				So(ok, ShouldBeTrue)
				So(host, ShouldEqual, "test-expiration")
			}
		})
	})
}

func TestSubscribeNew(t *testing.T) {
	stop := make(chan struct{})
	defer close(stop)

	Convey("When the service 'test' is watched and a host registered", t, func() {
		hosts, _ := SubscribeNew("test_new")
		time.Sleep(200 * time.Millisecond)
		newHost := genHost("test-new")
		Register("test_new", newHost, stop)
		Convey("A host should be available in the channel", func() {
			host, ok := <-hosts
			So(ok, ShouldBeTrue)
			So(host, ShouldResemble, newHost)
		})
	})
}

func TestSubscribeUpdate(t *testing.T) {
	stop1 := make(chan struct{})
	stop2 := make(chan struct{})
	defer close(stop2)

	Convey("When the service 'test' is watched and a host updates its data", t, func() {
		hosts, _ := SubscribeUpdate("test_upd")
		newHost := genHost("test-update")
		Register("test_upd", newHost, stop1)
		close(stop1)
		newHost.Password = "newpass"
		Register("test_upd", newHost, stop2)

		Convey("A host should be available in the channel", func() {
			host, ok := <-hosts
			So(host, ShouldResemble, newHost)
			So(ok, ShouldBeTrue)
		})
	})
}
