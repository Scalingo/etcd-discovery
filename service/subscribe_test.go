package service

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

// Tests
func TestSubscribe(t *testing.T) {
	Convey("When we subscribe a service, we get all the notifications from it", t, func() {
		responses := Subscribe("test_subs")
		time.Sleep(200 * time.Millisecond)
		Convey("When something happens about this service, the responses must be gathered in the channel", func() {
			_, err := client.Create("/services/test_subs/key", "test", 0)
			So(err, ShouldBeNil)
			r := <-responses
			So(r, ShouldNotBeNil)
			So(r.Node.Key, ShouldEqual, "/services/test_subs/key")
			So(r.Action, ShouldEqual, "create")
			_, err = client.Delete("/services/test_subs/key", false)
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
		Register("test_expiration", genHost(), stop)
		stop <- true
		hosts := SubscribeDown("test_expiration")
		Convey("The name of the disappeared host should be returned", func() {
			host, ok := <-hosts
			So(host, ShouldEqual, hostname)
			So(ok, ShouldBeTrue)
		})
	})
}

func TestSubscribeNew(t *testing.T) {
	stop := make(chan bool)
	defer close(stop)

	Convey("When the service 'test' is watched and a host registered", t, func() {
		hosts := SubscribeNew("test_new")
		time.Sleep(200 * time.Millisecond)
		newHost := genHost()
		Register("test_new", newHost, stop)
		Convey("A host should be available in the channel", func() {
			host, ok := <-hosts
			So(host, ShouldResemble, newHost)
			So(ok, ShouldBeTrue)
		})
	})
}
