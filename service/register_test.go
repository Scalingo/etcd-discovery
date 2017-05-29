package service

import (
	"encoding/json"
	"fmt"
	"path"
	"testing"
	"time"

	etcd "github.com/coreos/etcd/client"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/context"
)

func TestRegister(t *testing.T) {
	Convey("After registering service test", t, func() {
		host := genHost("test-register")
		Convey("It should be available with etcd", func() {
			c := Register("test_register", host, nil, make(chan struct{}))

			<-c

			res, err := KAPI().Get(context.Background(), "/services/test_register/"+host.Name, &etcd.GetOptions{})
			So(err, ShouldBeNil)

			h := &Host{}
			json.Unmarshal([]byte(res.Node.Value), &h)

			So(path.Base(res.Node.Key), ShouldEqual, host.Name)
			So(h, ShouldResemble, host)
		})

		Convey(fmt.Sprintf("And the ttl must be < %d", HEARTBEAT_DURATION), func() {
			r := Register("test2_register", host, nil, make(chan struct{}))
			<-r
			res, err := KAPI().Get(context.Background(), "/services/test2_register/"+host.Name, &etcd.GetOptions{})
			So(err, ShouldBeNil)
			now := time.Now()
			duration := res.Node.Expiration.Sub(now)
			So(duration, ShouldBeLessThanOrEqualTo, HEARTBEAT_DURATION*time.Second)
		})

		Convey("And the serivce infos must be set", func() {
			infos := &Infos{
				Critical: true,
			}
			r := Register("test3_register", host, infos, make(chan struct{}))
			<-r
			res, err := KAPI().Get(context.Background(), "/services_infos/test3_register", &etcd.GetOptions{})
			So(err, ShouldBeNil)

			i := &Infos{}

			json.Unmarshal([]byte(res.Node.Value), &i)

			So(i, ShouldResemble, infos)
		})

		Convey("After sending stop, the service should disappear", func() {
			stop := make(chan struct{})
			host := genHost("test-disappear")
			r := Register("test4_register", host, nil, stop)
			<-r
			close(stop)
			time.Sleep(100 * time.Millisecond)
			_, err := KAPI().Get(context.Background(), "/services/test4_register/"+host.Name, &etcd.GetOptions{})
			So(etcd.IsKeyNotFound(err), ShouldBeTrue)
		})
	})
}

func TestWatcher(t *testing.T) {
	Convey("With two instances of the same service", t, func() {
		host1 := genHost("test-watcher-1")
		host2 := genHost("test-watcher-1")

		c1 := Register("test-watcher", host1, &Infos{
			Critical: true,
			User:     "host1",
			Password: "password1",
		}, make(chan struct{}))

		cred1 := <-c1
		So(cred1.User, ShouldEqual, "host1")
		So(cred1.Password, ShouldEqual, "password1")

		c2 := Register("test-watcher", host2, &Infos{
			Critical: true,
			User:     "host2",
			Password: "password2",
		}, make(chan struct{}))

		cred2 := <-c2
		So(cred2.User, ShouldEqual, "host2")
		So(cred2.Password, ShouldEqual, "password2")

		cred1 = <-c1
		So(cred1.User, ShouldEqual, "host2")
		So(cred1.Password, ShouldEqual, "password2")

	})
}
