package service

import (
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestHostUrl(t *testing.T) {
	Convey("Given a specific host", t, func() {
		host, err := NewHost("host", Ports{"http": "111"}, "user", "password")
		So(host, ShouldNotBeNil)
		So(err, ShouldBeNil)
		hUrl, err := host.Url("http", "/")
		Convey("the error should bi nil", func() {
			So(err, ShouldBeNil)
		})
		Convey("It should return a valid URL", func() {
			u, err := url.Parse(hUrl)
			So(u, ShouldNotBeNil)
			So(err, ShouldBeNil)
		})
		Convey("should contain the correct data", func() {
			correctUrl := "http://user:password@host:111/"
			So(hUrl, ShouldEqual, correctUrl)
		})
	})

	Convey("Given a host with only a user", t, func() {
		host, err := NewHost("host", Ports{"http": "111"}, "user")
		Convey("it should returns an error", func() {
			So(host, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "password should be too")
		})
	})

	Convey("Given a host with a nil interface", t, func() {
		host, err := NewHost("host", nil)
		Convey("it should returns an error", func() {
			So(host, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "an interface should be defined")
		})
	})

	Convey("Given a host without any interface", t, func() {
		host, err := NewHost("host", Ports{})
		Convey("it should returns an error", func() {
			So(host, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "an interface should be defined")
		})
	})

	Convey("Given a host without credentials", t, func() {
		host, err := NewHost("host", Ports{"http": "111"})
		So(host, ShouldNotBeNil)
		So(err, ShouldBeNil)
		hUrl, err := host.Url("http", "/")
		Convey("the error should be nil", func() {
			So(err, ShouldBeNil)
		})
		Convey("The URL should be valid", func() {
			u, err := url.Parse(hUrl)
			So(u, ShouldNotBeNil)
			So(err, ShouldBeNil)
		})
		Convey("It should return the correct URL", func() {
			correctUrl := "http://host:111/"
			So(hUrl, ShouldEqual, correctUrl)
		})
	})
}
