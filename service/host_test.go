package service

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestHostUrl(t *testing.T) {
	Convey("With a host without any password", t, func() {
		host := genHost("test")
		host.User = ""
		host.Password = ""

		url, err := host.Url("http", "/path")
		So(err, ShouldBeNil)
		So(url, ShouldEqual, "http://public.dev:10000/path")
	})

	Convey("With a host with a password", t, func() {
		host := genHost("test")
		url, err := host.Url("http", "/path")
		So(err, ShouldBeNil)
		So(url, ShouldEqual, "http://user:password@public.dev:10000/path")
	})

	Convey("When the port does'nt exists", t, func() {
		host := genHost("test")
		url, err := host.Url("htjp", "/path")
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "unknown scheme")
		So(len(url), ShouldEqual, 0)
	})
}

func TestHostPrivateUrl(t *testing.T) {
	Convey("With a host without any password", t, func() {
		host := genHost("test")
		host.User = ""
		host.Password = ""

		url, err := host.PrivateUrl("http", "/path")
		So(err, ShouldBeNil)
		So(url, ShouldEqual, "http://test-private.dev:20000/path")
	})

	Convey("With a host with a password", t, func() {
		host := genHost("test")
		url, err := host.PrivateUrl("http", "/path")
		So(err, ShouldBeNil)
		So(url, ShouldEqual, "http://user:password@test-private.dev:20000/path")
	})

	Convey("When the port does'nt exists", t, func() {
		host := genHost("test")
		url, err := host.PrivateUrl("htjp", "/path")
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "unknown scheme")
		So(len(url), ShouldEqual, 0)
	})

	Convey("When the host does not support private urls", t, func() {
		host := genHost("test")
		host.PrivateHostname = ""
		url, err := host.PrivateUrl("http", "/path")
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "This service does not support private urls")
		So(len(url), ShouldEqual, 0)
	})
}

func TestHostsString(t *testing.T) {
	Convey("With a list of two hosts", t, func() {
		host1 := genHost("test")
		host2 := genHost("test")
		host1.PrivateHostname = ""
		hosts := Hosts{host1, host2}
		So(hosts.String(), ShouldEqual, "public.dev, test-private.dev")
	})

	Convey("With an empty list", t, func() {
		hosts := Hosts{}
		So(hosts.String(), ShouldEqual, "")
	})
}
