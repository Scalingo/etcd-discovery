package service

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestServiceAll(t *testing.T) {
	Convey("With no services", t, func() {
		s, err := Get("service-test-get-1").Service()
		So(err, ShouldBeNil)

		hosts, err := s.All()
		So(err, ShouldBeNil)
		So(len(hosts), ShouldEqual, 0)
	})

	Convey("With two services", t, func() {
		host1 := genHost("test1")
		host2 := genHost("test2")
		r1, err := Register("test-get-222", host1)
		So(err, ShouldBeNil)
		Reset(func() { So(r1.Stop(), ShouldBeNil) })
		r2, err := Register("test-get-222", host2)
		So(err, ShouldBeNil)
		Reset(func() { So(r2.Stop(), ShouldBeNil) })

		r1.WaitRegistration()
		r2.WaitRegistration()

		s, err := Get("test-get-222").Service()
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
		s, err := Get("service-test-1").Service()
		So(err, ShouldBeNil)
		host, err := s.First()
		So(err, ShouldNotBeNil)
		So(host, ShouldBeNil)
		So(err.Error(), ShouldEqual, "No host found for this service")
	})

	Convey("With a service", t, func() {
		host1 := genHost("test1")

		r, err := Register("test-truc", host1)
		So(err, ShouldBeNil)
		Reset(func() { So(r.Stop(), ShouldBeNil) })
		r.WaitRegistration()

		s, err := Get("test-truc").Service()
		So(err, ShouldBeNil)
		host, err := s.First()
		So(err, ShouldBeNil)
		So(host, ShouldNotBeNil)
		So(host.PrivateHostname, ShouldEqual, host1.PrivateHostname)
	})
}

func TestServiceOne(t *testing.T) {
	Convey("With no services", t, func() {
		s, err := Get("service-test-1").Service()
		So(err, ShouldBeNil)
		host, err := s.One()
		So(err, ShouldNotBeNil)
		So(host, ShouldBeNil)
		So(err.Error(), ShouldEqual, "No host found for this service")
	})

	Convey("With a service", t, func() {
		host1 := genHost("test1")
		r, err := Register("test-truc", host1)
		So(err, ShouldBeNil)
		Reset(func() { So(r.Stop(), ShouldBeNil) })

		r.WaitRegistration()

		s, err := Get("test-truc").Service()
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
			host := genHost("test")
			host.User = ""
			host.Password = ""

			r, err := Register("service-url-1", host)
			So(err, ShouldBeNil)
			Reset(func() { So(r.Stop(), ShouldBeNil) })

			r.WaitRegistration()

			s, err := Get("service-url-1").Service()
			So(err, ShouldBeNil)
			url, err := s.URL("http", "/path")
			So(err, ShouldBeNil)
			So(url, ShouldEqual, "http://public.dev:10000/path")
		})

		Convey("With a host with a password", func() {
			host := genHost("test")

			r, err := Register("service-url-3", host)
			So(err, ShouldBeNil)
			Reset(func() { So(r.Stop(), ShouldBeNil) })

			r.WaitRegistration()

			s, err := Get("service-url-3").Service()
			So(err, ShouldBeNil)
			url, err := s.URL("http", "/path")
			So(err, ShouldBeNil)
			So(url, ShouldEqual, "http://user:password@public.dev:10000/path")
		})

		Convey("When the port does'nt exists", func() {
			host := genHost("test")

			r, err := Register("service-url-4", host)
			So(err, ShouldBeNil)
			Reset(func() { So(r.Stop(), ShouldBeNil) })

			r.WaitRegistration()

			s, err := Get("service-url-4").Service()
			So(err, ShouldBeNil)
			url, err := s.URL("htjp", "/path")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "unknown scheme")
			So(len(url), ShouldEqual, 0)
		})
	})
}
