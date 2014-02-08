package service

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestGet(t *testing.T) {
	Convey("Given two registered services", t, func() {
		stop1, stop2 := make(chan bool), make(chan bool)
		Register("test_service", genHost("host1"), stop1)
		Register("test_service", genHost("host2"), stop2)
		Convey("We should have 2 hosts", func() {
			hosts, err := Get("test_service")
			So(len(hosts), ShouldEqual, 2)
			So(err, ShouldBeNil)
		})
		stop1 <- true
		stop2 <- true
	})
}
