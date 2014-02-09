package service

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestGetNoHost(t *testing.T) {
	Convey("Without any service", t, func() {
		Convey("Get should return an empty slice", func() {
			hosts, err := Get("test_no_service")
			So(len(hosts), ShouldEqual, 0)
			So(err, ShouldBeNil)
		})
	})
}

func TestGet(t *testing.T) {
	Convey("Given two registered services", t, func() {
		stop1, stop2 := make(chan bool), make(chan bool)
		host1, host2 := genHost("host1"), genHost("host2")
		Register("test_service", host1, stop1)
		Register("test_service", host2, stop2)
		Convey("We should have 2 hosts", func() {
			hosts, err := Get("test_service")
			So(len(hosts), ShouldEqual, 2)
			So(hosts[0], ShouldResemble, host1)
			So(hosts[1], ShouldResemble, host2)
			So(err, ShouldBeNil)
		})
		stop1 <- true
		stop2 <- true
	})
}
