package service

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestServiceAll(t *testing.T) {
	Convey("With no services", t, func() {
		ctx := context.Background()
		s, err := Get(ctx, "service-test-get-1").Service()
		So(err, ShouldBeNil)

		hosts, err := s.All()
		So(err, ShouldBeNil)
		So(len(hosts), ShouldEqual, 0)
	})

	Convey("With two services", t, func() {
		ctx := context.Background()
		host1 := genHost("test1")
		host2 := genHost("test2")
		w1 := Register(ctx, "test-get-222", host1)
		w2 := Register(ctx, "test-get-222", host2)

		w1.WaitRegistration()
		w2.WaitRegistration()

		s, err := Get(ctx, "test-get-222").Service()
		hosts, err := s.All()
		So(err, ShouldBeNil)
		So(len(hosts), ShouldEqual, 2)
		if hosts[0].PrivateHostname == "test1-private.dev" {
			So(hosts[1].PrivateHostname, ShouldEqual, "test2-private.dev")
		} else {
			So(hosts[1].PrivateHostname, ShouldEqual, "test1-private.dev")
			So(hosts[0].PrivateHostname, ShouldEqual, "test2-private.dev")
		}
	})
}

func TestServiceFirst(t *testing.T) {
	Convey("With no services", t, func() {
		ctx := context.Background()
		s, err := Get(ctx, "service-test-1").Service()
		So(err, ShouldBeNil)
		host, err := s.First()
		So(err, ShouldNotBeNil)
		So(host, ShouldBeNil)
		So(err.Error(), ShouldEqual, "No host found for this service")
	})

	Convey("With a service", t, func() {
		ctx := context.Background()
		host1 := genHost("test1")
		w := Register(ctx, "test-truc", host1)
		w.WaitRegistration()

		s, err := Get(ctx, "test-truc").Service()
		So(err, ShouldBeNil)
		host, err := s.First()
		So(err, ShouldBeNil)
		So(host, ShouldNotBeNil)
		So(host.PrivateHostname, ShouldEqual, host1.PrivateHostname)
	})
}

func TestServiceOne(t *testing.T) {
	Convey("With no services", t, func() {
		ctx := context.Background()
		s, err := Get(ctx, "service-test-1").Service()
		So(err, ShouldBeNil)
		host, err := s.One()
		So(err, ShouldNotBeNil)
		So(host, ShouldBeNil)
		So(err.Error(), ShouldEqual, "No host found for this service")
	})

	Convey("With a service", t, func() {
		ctx := context.Background()
		host1 := genHost("test1")
		w := Register(ctx, "test-truc", host1)
		w.WaitRegistration()

		s, err := Get(ctx, "test-truc").Service()
		So(err, ShouldBeNil)
		host, err := s.One()
		So(err, ShouldBeNil)
		So(host, ShouldNotBeNil)
		So(host.PrivateHostname, ShouldEqual, host1.PrivateHostname)
	})
}

func TestServiceUrl(t *testing.T) {
	Convey("With a public service", t, func() {
		Convey("With a service without any password", func() {
			ctx := context.Background()
			host := genHost("test")
			host.User = ""
			host.Password = ""
			w := Register(ctx, "service-url-1", host)
			w.WaitRegistration()

			s, err := Get(ctx, "service-url-1").Service()
			So(err, ShouldBeNil)
			url, err := s.URL("http", "/path")
			So(err, ShouldBeNil)
			So(url, ShouldEqual, "http://public.dev:10000/path")
		})

		Convey("With a host with a password", func() {
			ctx := context.Background()
			host := genHost("test")
			w := Register(ctx, "service-url-3", host)
			w.WaitRegistration()

			s, err := Get(ctx, "service-url-3").Service()
			So(err, ShouldBeNil)
			url, err := s.URL("http", "/path")
			So(err, ShouldBeNil)
			So(url, ShouldEqual, "http://user:password@public.dev:10000/path")
		})

		Convey("When the port does'nt exists", func() {
			ctx := context.Background()
			host := genHost("test")
			w := Register(ctx, "service-url-4", host)
			w.WaitRegistration()

			s, err := Get(ctx, "service-url-4").Service()
			So(err, ShouldBeNil)
			url, err := s.URL("htjp", "/path")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "unknown scheme")
			So(len(url), ShouldEqual, 0)
		})
	})
}
