package service

import (
	. "github.com/smartystreets/goconvey/convey"
	"net/url"
	"testing"
)

func TestHostUrl(t *testing.T) {
	Convey("Given a specific host", t, func() {
		host := NewHost("host", "port", "user", "password", "http")
		So(host, ShouldNotBeNil)
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

	Convey("Given a host without scheme", t, func() {
		host := NewHost("host", "port", "user", "password")
		So(host, ShouldNotBeNil)
		hUrl := host.Url("/")
		Convey("The URL should be valid", func() {
			u, err := url.Parse(hUrl)
			So(u, ShouldNotBeNil)
			So(err, ShouldBeNil)
		})
		Convey("The scheme should be HTTP", func() {
			u, _ := url.Parse(hUrl)
			So(u.Scheme, ShouldEqual, "http")
		})
		Convey("It should return the correct URL", func() {
			correctUrl := "http://user:password@host:port/"
			So(hUrl, ShouldEqual, correctUrl)
		})
	})

	Convey("Given a host without credentials", t, func() {
		host := NewHost("host", "port")
		So(host, ShouldNotBeNil)
		hUrl := host.Url("/")
		Convey("The URL should be valid", func() {
			u, err := url.Parse(hUrl)
			So(u, ShouldNotBeNil)
			So(err, ShouldBeNil)
		})
		Convey("The scheme should be HTTP", func() {
			u, _ := url.Parse(hUrl)
			So(u.Scheme, ShouldEqual, "http")
		})
		Convey("It should return the correct URL", func() {
			correctUrl := "http://host:port/"
			So(hUrl, ShouldEqual, correctUrl)
		})
	})
}
