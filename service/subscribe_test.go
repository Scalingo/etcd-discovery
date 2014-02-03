package service

import (
	"github.com/coreos/go-etcd/etcd"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

// Tests
func TestSubscribe(t *testing.T) {
	Convey("When we subscribe a service, we get all the notifications from it", t, func() {
		responses := Subscribe("test")
		Convey("When something happens about this service, the responses must be gathered in the channel", func() {
			client.Create("/services/test/key", "test", 0)
			r := <-responses
			So(r, ShouldNotBeNil)
			So(r.Node.Key, ShouldEqual, "/services/test/key")
			So(r.Action, ShouldEqual,"create")
			client.Delete("/services/test/key", false)
			r = <-responses
			So(r, ShouldNotBeNil)
			So(r.Node.Key, ShouldEqual, "/services/test/key")
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

	Convey("When the service 'test' is watched and a host registered", t, func() {
		hosts := SubscribeNew("test")
		Register("test", genHost(), stop)
		Convey("A host should be available in the channel", func() {
			host, ok := <-hosts
			So(host, ShouldNotBeNil)
			So(ok, ShouldBeTrue)
		})
		stop <- true
	})
}


var sampleResponse = &etcd.Response{
	Node: &etcd.Node{
		Key: "/services/test/example.org",
		Nodes: etcd.Nodes{
			etcd.Node{
				Key: "/services/test/example.org/user",
				Value: "user",
			},
			etcd.Node{
				Key: "/services/test/example.org/password",
				Value: "password",
			},
			etcd.Node{
				Key: "/services/test/example.org/port",
				Value: "port",
			},
		},
	},
}

func TestBuildHostFromResponse(t *testing.T) {
	host := buildHostFromResponse(sampleResponse)
	Convey("Given a sample response, we got a filled Host", t, func() {
		So(host.User, ShouldEqual, "user")
		So(host.Password, ShouldEqual, "password")
		So(host.Port, ShouldEqual, "port")
	})
}
