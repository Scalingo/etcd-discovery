package service

import (
	. "github.com/smartystreets/goconvey/convey"
	"net/url"
	"testing"
)

func TestHostUrl(t *testing.T) {
	Convey("Given a specific host", t, func() {
		host := &Host{"host", "user", "password", "port"}
		hUrl := host.Url("/")
		Convey("It should return a valid URL", func() {
			u, err := url.Parse(hUrl)
			So(u, ShouldNotBeNil)
			So(err, ShouldBeNil)
		})
		Convey("should contain the correct data", func() {
			correctUrl := "http://user:password@host:port/"
			So(hUrl, ShouldEqual, correctUrl)
		})
	})
}
