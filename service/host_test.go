package service

import (
	"errors"
	"testing"

	c "github.com/smartystreets/goconvey/convey"
)

func TestHostUrl(t *testing.T) {
	c.Convey("With a host without any password", t, func() {
		host := genHost("test")
		host.User = ""
		host.Password = ""

		url, err := host.URL("http", "/path")
		c.So(err, c.ShouldBeNil)
		c.So(url, c.ShouldEqual, "http://public.dev:10000/path")
	})

	c.Convey("With a host with a password", t, func() {
		host := genHost("test")
		url, err := host.URL("http", "/path")
		c.So(err, c.ShouldBeNil)
		c.So(url, c.ShouldEqual, "http://user:password@public.dev:10000/path")
	})

	c.Convey("When the port doesn't exists", t, func() {
		host := genHost("test")
		url, err := host.URL("htjp", "/path")
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldEqual, "unknown scheme")
		c.So(len(url), c.ShouldEqual, 0)
	})

	c.Convey("When the scheme is not provided", t, func() {
		host := genHost("test")
		url, err := host.URL("", "/path")
		c.So(err, c.ShouldBeNil)
		c.So(url, c.ShouldEqual, "http://user:password@public.dev:10000/path")
	})
}

func TestHostPrivateUrl(t *testing.T) {
	c.Convey("With a host without any password", t, func() {
		host := genHost("test")
		host.User = ""
		host.Password = ""

		url, err := host.PrivateURL("http", "/path")
		c.So(err, c.ShouldBeNil)
		c.So(url, c.ShouldEqual, "http://test-private.dev:20000/path")
	})

	c.Convey("With a host with a password", t, func() {
		host := genHost("test")
		url, err := host.PrivateURL("http", "/path")
		c.So(err, c.ShouldBeNil)
		c.So(url, c.ShouldEqual, "http://user:password@test-private.dev:20000/path")
	})

	c.Convey("When the port doesn't exists", t, func() {
		host := genHost("test")
		url, err := host.PrivateURL("htjp", "/path")
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldEqual, "unknown scheme")
		c.So(len(url), c.ShouldEqual, 0)
	})

	c.Convey("When the scheme is not provided", t, func() {
		host := genHost("test")
		url, err := host.PrivateURL("", "/path")
		c.So(err, c.ShouldBeNil)
		c.So(url, c.ShouldEqual, "http://user:password@test-private.dev:20000/path")
	})

	c.Convey("When the host does not support private urls, it should fall back to URL", t, func() {
		host := genHost("test")
		host.PrivateHostname = ""
		url, err := host.PrivateURL("http", "/path")
		c.So(err, c.ShouldBeNil)
		c.So(url, c.ShouldEqual, "http://user:password@public.dev:10000/path")
	})
}

func TestHostsString(t *testing.T) {
	c.Convey("With a list of two hosts", t, func() {
		host1 := genHost("test")
		host2 := genHost("test")
		host1.PrivateHostname = ""
		hosts := Hosts{&host1, &host2}
		c.So(hosts.String(), c.ShouldEqual, "public.dev, test-private.dev")
	})

	c.Convey("With an empty list", t, func() {
		hosts := Hosts{}
		c.So(hosts.String(), c.ShouldEqual, "")
	})
}

func TestGetHostResponse(t *testing.T) {
	c.Convey("With an errored response", t, func() {
		response := &GetHostResponse{
			err:  errors.New("TestError"),
			host: nil,
		}

		c.Convey("The err method should return an error", func() {
			c.So(response.Err(), c.ShouldNotBeNil)
			c.So(response.Err().Error(), c.ShouldEqual, "TestError")
		})

		c.Convey("The Host method should return an error", func() {
			host, err := response.Host()
			c.So(err, c.ShouldNotBeNil)
			c.So(err.Error(), c.ShouldEqual, "TestError")
			c.So(host, c.ShouldBeNil)
		})

		c.Convey("The URL method should return an error", func() {
			url, err := response.URL("http", "/path")
			c.So(err, c.ShouldNotBeNil)
			c.So(err.Error(), c.ShouldEqual, "TestError")
			c.So(url, c.ShouldEqual, "")
		})

		c.Convey("The PrivateURL should return an error", func() {
			url, err := response.PrivateURL("http", "/path")
			c.So(err, c.ShouldNotBeNil)
			c.So(err.Error(), c.ShouldEqual, "TestError")
			c.So(url, c.ShouldEqual, "")
		})
	})

	c.Convey("With a valid response", t, func() {
		host := genHost("test-service")
		response := &GetHostResponse{
			err:  nil,
			host: &host,
		}

		c.Convey("The err method should not return an error", func() {
			c.So(response.Err(), c.ShouldBeNil)
		})

		c.Convey("The Host method should return a valid host", func() {
			h, err := response.Host()
			c.So(err, c.ShouldBeNil)
			c.So(h, c.ShouldResemble, &host)
		})

		c.Convey("The URL method should return a valid url", func() {
			url, err := response.URL("http", "/path")
			c.So(err, c.ShouldBeNil)
			c.So(url, c.ShouldEqual, "http://user:password@public.dev:10000/path")
		})

		c.Convey("The Private URL should return a valid url", func() {
			url, err := response.PrivateURL("http", "/path")
			c.So(err, c.ShouldBeNil)
			c.So(url, c.ShouldEqual, "http://user:password@test-service-private.dev:20000/path")
		})
	})
}
