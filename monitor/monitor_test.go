package monitor

import (
	"github.com/Appsdeck/etcd-discovery/service"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func start(service string) {
	go Start(service)
	time.Sleep(200 * time.Millisecond)
}

func TestStart(t *testing.T) {
	Convey("When monitoring a services 'test_start'", t, func() {
		Convey("Its slice must be defined", func() {
			start("test_start1")
			So(services["test_start1"], ShouldNotBeNil)
		})
		Convey("There must be no hosts", func() {
			start("test_start2")

			hosts, err := Hosts("test_start2")
			So(len(hosts), ShouldEqual, 0)
			So(err, ShouldBeNil)
		})
		Convey("After adding registering 2 hosts, the slice must be filled", func() {
			start("test_start3")

			stop1, stop2 := make(chan bool), make(chan bool)
			service.Register("test_start3", &service.Host{Name: "host_start3_1"}, stop1)
			service.Register("test_start3", &service.Host{Name: "host_start3_2"}, stop2)
			waitRegistration()

			hosts, err := Hosts("test_start3")
			So(len(hosts), ShouldEqual, 2)
			So(err, ShouldBeNil)
		})
		Convey("If a node is removed, one must left", func() {
			start("test_start4")

			stop1, stop2 := make(chan bool), make(chan bool)
			service.Register("test_start4", &service.Host{Name: "host_start4_1"}, stop1)
			service.Register("test_start4", &service.Host{Name: "host_start4_2"}, stop2)

			stop1 <- true
			waitExpiration()

			hosts, err := Hosts("test_start4")
			So(len(hosts), ShouldEqual, 1)
			So(err, ShouldBeNil)
		})
	})
}

func TestAttributes(t *testing.T) {
	host := &service.Host{Name: "host_attr", User: "user", Password: "password", Port: "1000"}
	Convey("When a host is added to the slice", t, func() {

		Convey("It should have all the attributes of the registered host", func() {
			start("test_attr1")

			stop := make(chan bool)
			service.Register("test_attr1", host, stop)
			waitRegistration()

			hosts, err := Hosts("test_attr1")
			So(err, ShouldBeNil)
			So(len(hosts), ShouldEqual, 1)
			So(hosts[0], ShouldResemble, host)
		})

		Convey("When the same host is removed and re-added, nothing should change", func() {
			start("test_attr2")

			stop := make(chan bool)
			service.Register("test_attr2", host, stop)

			stop <- true
			waitExpiration()

			service.Register("test_attr2", host, make(chan bool))
			waitRegistration()

			hosts, err := Hosts("test_attr2")
			So(err, ShouldBeNil)
			So(len(hosts), ShouldEqual, 1)
			So(hosts[0], ShouldResemble, host)
		})
	})
}

func TestNoService(t *testing.T) {
	Convey("When trying to get hosts from a non-monitored service", t, func() {
		_, err := Hosts("test_non_monitored")
		Convey("It must return an error", func() {
			So(err, ShouldEqual, NoSuchServiceError)
		})
	})
}

func TestUpdate(t *testing.T) {
	Convey("When a node is update, the data should be updated", t, func() {
		host1 := &service.Host{Name: "host_attr", User: "user", Password: "password", Port: "1000"}
		host2 := &service.Host{Name: "host_attr", User: "user", Password: "password2", Port: "1000"}
		start("test_upd")

		stop := make(chan bool)
		service.Register("test_upd", host1, stop)
		waitRegistration()
		stop <- true

		hosts, _ := Hosts("test_upd")
		So(len(hosts), ShouldEqual, 1)
		So(hosts[0].Password, ShouldEqual, host1.Password)

		service.Register("test_upd", host2, stop)
		waitRegistration()

		hosts, _ = Hosts("test_upd")
		So(len(hosts), ShouldEqual, 1)
		So(hosts[0].Password, ShouldEqual, host2.Password)
	})
}

func waitRegistration() {
	time.Sleep(200 * time.Millisecond)
}

func waitExpiration() {
	time.Sleep(service.HEARTBEAT_DURATION * 2 * time.Second)
}
