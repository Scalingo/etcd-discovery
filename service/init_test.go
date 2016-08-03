package service

import (
	"log"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestInit(t *testing.T) {
	Convey("After initialization", t, func() {
		Convey("client should be set", func() {
			So(Client(), ShouldNotBeNil)
		})
		Convey("hostname should be set", func() {
			So(hostname, ShouldNotBeNil)
		})
		Convey("the logger should be correctly parameterised", func() {
			So(logger.Prefix(), ShouldEqual, "[etcd-discovery] ")
			So(logger.Flags(), ShouldEqual, log.LstdFlags)
		})
	})
}
