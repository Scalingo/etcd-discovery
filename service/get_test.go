package service

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetNoHost(t *testing.T) {
	Convey("Without any service", t, func() {
		Convey("Get should return an empty slice", func() {
			service, err := Get("test_no_service")
			So(len(service.Hosts), ShouldEqual, 0)
			So(err, ShouldBeNil)
		})
	})
}

func TestGet(t *testing.T) {
	Convey("Given two registered services", t, func() {
		stop1, stop2 := make(chan struct{}), make(chan struct{})
		host1, host2 := genHost("host1"), genHost("host2")
		infos := &Infos{
			Critical: true,
		}
		r1 := Register("test_service_get", host1, infos, stop1)
		r2 := Register("test_service_get", host2, infos, stop2)
		<-r1
		<-r2
		Convey("We should have 2 hosts", func() {
			service, err := Get("test_service_get")
			hosts := service.Hosts
			So(service.Infos, ShouldResemble, infos)
			So(len(hosts), ShouldEqual, 2)
			if hosts[0].Name == host1.Name {
				So(hosts[0], ShouldResemble, host1)
				So(hosts[1], ShouldResemble, host2)
			} else {
				So(hosts[1], ShouldResemble, host1)
				So(hosts[0], ShouldResemble, host2)
			}
			So(err, ShouldBeNil)
		})
		close(stop1)
		close(stop2)
	})
}
