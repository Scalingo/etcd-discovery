package service

import (
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetNoHost(t *testing.T) {
	Convey("Without any service", t, func() {
		Convey("Get should return an empty slice", func() {
			hosts, err := Get("test_no_service").All()
			So(err, ShouldBeNil)
			So(len(hosts), ShouldEqual, 0)
		})
	})
}

func TestGet(t *testing.T) {
	Convey("With registred services", t, func() {
		stop1, stop2 := make(chan struct{}), make(chan struct{})
		host1, host2 := genHost("host1"), genHost("host2")
		host1.Name = "test_service_get"
		host2.Name = "test_service_get"
		w1 := Register("test_service_get", host1, stop1)
		w2 := Register("test_service_get", host2, stop2)
		w1.WaitRegistration()
		w2.WaitRegistration()
		Convey("We should have 2 hosts", func() {
			hosts, err := Get("test_service_get").All()
			So(err, ShouldBeNil)
			So(len(hosts), ShouldEqual, 2)
			if hosts[0].UUID == w1.UUID() {
				host1.UUID = hosts[0].UUID
				host2.UUID = hosts[1].UUID
				So(hosts[0], ShouldResemble, &host1)
				So(hosts[1], ShouldResemble, &host2)
			} else {
				host1.UUID = hosts[1].UUID
				host2.UUID = hosts[0].UUID
				So(hosts[1], ShouldResemble, &host1)
				So(hosts[0], ShouldResemble, &host2)
			}
			So(err, ShouldBeNil)
		})
		close(stop1)
		close(stop2)
	})
}

func TestGetServiceResponse(t *testing.T) {
	Convey("With an errored Response", t, func() {
		response := &GetServiceResponse{
			err:     errors.New("TestError"),
			service: nil,
		}

		Convey("Err should return an error", func() {
			err := response.Err()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "TestError")
		})

		Convey("Service should return an error", func() {
			service, err := response.Service()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "TestError")
			So(service, ShouldBeNil)
		})

		Convey("All should return an error", func() {
			h, err := response.All()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "TestError")
			So(len(h), ShouldEqual, 0)
		})

		Convey("Url should return an error", func() {
			url, err := response.URL("http", "/path")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "TestError")
			So(url, ShouldEqual, "")
		})

		Convey("One should return an errored host response", func() {
			response := response.One()
			So(response, ShouldNotBeNil)
			So(response.Err().Error(), ShouldEqual, "TestError")
		})

		Convey("First should return an errored host response", func() {
			response := response.First()
			So(response, ShouldNotBeNil)
			So(response.Err().Error(), ShouldEqual, "TestError")
		})
	})

	Convey("With a valid response", t, func() {
		service := genService("test-service-11122233444")
		response := &GetServiceResponse{
			err:     nil,
			service: service,
		}

		Convey("Err should be nil", func() {
			So(response.Err(), ShouldBeNil)
		})

		Convey("Service should respond a valid service", func() {
			s, err := response.Service()
			So(err, ShouldBeNil)
			So(s, ShouldResemble, service)
		})

		Convey("All should return an empty array", func() {
			hosts, err := response.All()
			So(err, ShouldBeNil)
			So(len(hosts), ShouldEqual, 0)
		})

		Convey("Url should return a valid url", func() {
			url, err := response.URL("http", "/path")
			So(err, ShouldBeNil)
			So(url, ShouldEqual, "http://user:password@public.dev:80/path")
		})

		Convey("One should pass the One error", func() {
			r := response.One()
			So(r.Err(), ShouldNotBeNil)
			So(r.Err().Error(), ShouldEqual, "No host found for this service")
		})

		Convey("First should pass the First error", func() {
			r := response.First()
			So(r.Err(), ShouldNotBeNil)
			So(r.Err().Error(), ShouldEqual, "No host found for this service")
		})
	})
}
