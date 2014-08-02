package service

import (
	"testing"
	"time"
	. "github.com/smartystreets/goconvey/convey"
)

// Tests
func TestSubscribe(t *testing.T) {
	Convey("When we subscribe a service, we get all the notifications from it", t, func() {
		responses, _ := Subscribe("test_subs")
		time.Sleep(200 * time.Millisecond)
		Convey("When something happens about this service, the responses must be gathered in the channel", func() {
			_, err := Client().Create("/services/test_subs/key", "test", 0)
			So(err, ShouldBeNil)
			r := <-responses
			So(r, ShouldNotBeNil)
			So(r.Node.Key, ShouldEqual, "/services/test_subs/key")
			So(r.Action, ShouldEqual, "create")
			_, err = Client().Delete("/services/test_subs/key", false)
			So(err, ShouldBeNil)
			r = <-responses
			So(r, ShouldNotBeNil)
			So(r.Node.Key, ShouldEqual, "/services/test_subs/key")
			So(r.Action, ShouldEqual, "delete")
		})
	})
}

func TestSubscribeDown(t *testing.T) {
	stop := make(chan bool)
	defer close(stop)

	Convey("When the service 'test' is watched and a host expired", t, func() {
		Register("test_expiration", genHost("test-expiration"), stop)
		stop <- true
		hosts, _ := SubscribeDown("test_expiration")
		Convey("The name of the disappeared host should be returned", func() {
			host, ok := <-hosts
			So(host, ShouldEqual, "test-expiration")
			So(ok, ShouldBeTrue)
		})
	})
}

func TestSubscribeNew(t *testing.T) {
	stop := make(chan bool)
	defer close(stop)

	Convey("When the service 'test' is watched and a host registered", t, func() {
		hosts, _ := SubscribeNew("test_new")
		time.Sleep(200 * time.Millisecond)
		newHost := genHost("test-new")
		Register("test_new", newHost, stop)
		Convey("A host should be available in the channel", func() {
			host, ok := <-hosts
			So(host, ShouldResemble, newHost)
			So(ok, ShouldBeTrue)
		})
	})
}

func TestSubscribeUpdate(t *testing.T) {
	stop := make(chan bool)
	defer close(stop)

	Convey("When the service 'test' is watched and a host updates its data", t, func() {
		hosts, _ := SubscribeUpdate("test_upd")
		time.Sleep(200 * time.Millisecond)
		newHost := genHost("test-update")
		Register("test_upd", newHost, stop)
		stop <- true
		newHost.Password = "newpass"
		Register("test_upd", newHost, stop)

		Convey("A host should be available in the channel", func() {
			host, ok := <-hosts
			So(host, ShouldResemble, newHost)
			So(ok, ShouldBeTrue)
		})
	})
}
