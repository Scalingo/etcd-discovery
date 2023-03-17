package service

import (
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestHostUrl(t *testing.T) {
	Convey("With a host without any password", t, func() {
		host := genHost("test")
		host.User = ""
		host.Password = ""

		url, err := host.URL("http", "/path")
		So(err, ShouldBeNil)
		So(url, ShouldEqual, "http://public.dev:10000/path")
	})

	Convey("With a host with a password", t, func() {
		host := genHost("test")
		url, err := host.URL("http", "/path")
		So(err, ShouldBeNil)
		So(url, ShouldEqual, "http://user:password@public.dev:10000/path")
	})

	Convey("When the port doesn't exists", t, func() {
		host := genHost("test")
		url, err := host.URL("htjp", "/path")
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "unknown scheme")
		So(len(url), ShouldEqual, 0)
	})

	Convey("When the scheme is not provided", t, func() {
		host := genHost("test")
		url, err := host.URL("", "/path")
		So(err, ShouldBeNil)
		So(url, ShouldEqual, "http://user:password@public.dev:10000/path")
	})
}

func TestHostPrivateUrl(t *testing.T) {
	Convey("With a host without any password", t, func() {
		host := genHost("test")
		host.User = ""
		host.Password = ""

		url, err := host.PrivateURL("http", "/path")
		So(err, ShouldBeNil)
		So(url, ShouldEqual, "http://test-private.dev:20000/path")
	})

	Convey("With a host with a password", t, func() {
		host := genHost("test")
		url, err := host.PrivateURL("http", "/path")
		So(err, ShouldBeNil)
		So(url, ShouldEqual, "http://user:password@test-private.dev:20000/path")
	})

	Convey("When the port doesn't exists", t, func() {
		host := genHost("test")
		url, err := host.PrivateURL("htjp", "/path")
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "unknown scheme")
		So(len(url), ShouldEqual, 0)
	})

	Convey("When the scheme is not provided", t, func() {
		host := genHost("test")
		url, err := host.PrivateURL("", "/path")
		So(err, ShouldBeNil)
		So(url, ShouldEqual, "http://user:password@test-private.dev:20000/path")
	})

	Convey("When the host does not support private urls, it should fall back to URL", t, func() {
		host := genHost("test")
		host.PrivateHostname = ""
		url, err := host.PrivateURL("http", "/path")
		So(err, ShouldBeNil)
		So(url, ShouldEqual, "http://user:password@public.dev:10000/path")
	})
}

func TestHostsString(t *testing.T) {
	Convey("With a list of two hosts", t, func() {
		host1 := genHost("test")
		host2 := genHost("test")
		host1.PrivateHostname = ""
		hosts := Hosts{&host1, &host2}
		So(hosts.String(), ShouldEqual, "public.dev, test-private.dev")
	})

	Convey("With an empty list", t, func() {
		hosts := Hosts{}
		So(hosts.String(), ShouldEqual, "")
	})
}

func TestGetHostResponse(t *testing.T) {
	Convey("With an errored response", t, func() {
		response := &GetHostResponse{
			err:  errors.New("TestError"),
			host: nil,
		}

		Convey("The err method should return an error", func() {
			So(response.Err(), ShouldNotBeNil)
			So(response.Err().Error(), ShouldEqual, "TestError")
		})

		Convey("The Host method should return an error", func() {
			host, err := response.Host()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "TestError")
			So(host, ShouldBeNil)
		})

		Convey("The URL method should return an error", func() {
			url, err := response.URL("http", "/path")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "TestError")
			So(url, ShouldEqual, "")
		})

		Convey("The PrivateURL should return an error", func() {
			url, err := response.PrivateURL("http", "/path")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "TestError")
			So(url, ShouldEqual, "")
		})
	})

	Convey("With a valid response", t, func() {
		host := genHost("test-service")
		response := &GetHostResponse{
			err:  nil,
			host: &host,
		}

		Convey("The err method should not return an error", func() {
			So(response.Err(), ShouldBeNil)
		})

		Convey("The Host method should return a valid host", func() {
			h, err := response.Host()
			So(err, ShouldBeNil)
			So(h, ShouldResemble, &host)
		})

		Convey("The URL method should return a valid url", func() {
			url, err := response.URL("http", "/path")
			So(err, ShouldBeNil)
			So(url, ShouldEqual, "http://user:password@public.dev:10000/path")
		})

		Convey("The Private URL should return a valid url", func() {
			url, err := response.PrivateURL("http", "/path")
			So(err, ShouldBeNil)
			So(url, ShouldEqual, "http://user:password@test-service-private.dev:20000/path")
		})
	})
}
