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
			r, err := Register("test_register", host, nil, make(chan struct{}))
			So(err, ShouldBeNil)
			<-r

			res, err := KAPI().Get(context.Background(), "/services/test_register/"+host.Name, &etcd.GetOptions{})
			So(err, ShouldBeNil)

			h := &Host{}
			json.Unmarshal([]byte(res.Node.Value), &h)

			So(path.Base(res.Node.Key), ShouldEqual, host.Name)
			So(h, ShouldResemble, host)
		})

		Convey(fmt.Sprintf("And the ttl must be < %d", HEARTBEAT_DURATION), func() {
			r, _ := Register("test2_register", host, nil, make(chan struct{}))
			<-r
			res, err := KAPI().Get(context.Background(), "/services/test2_register/"+host.Name, &etcd.GetOptions{})
			So(err, ShouldBeNil)
			now := time.Now()
			duration := res.Node.Expiration.Sub(now)
			So(duration, ShouldBeLessThanOrEqualTo, HEARTBEAT_DURATION*time.Second)
		})

		Convey("After sending stop, the service should disappear", func() {
			stop := make(chan struct{})
			host := genHost("test-disappear")
			r, _ := Register("test3_register", host, nil, stop)
			<-r
			close(stop)
			time.Sleep(100 * time.Millisecond)
			_, err := KAPI().Get(context.Background(), "/services/test3_register/"+host.Name, &etcd.GetOptions{})
			So(etcd.IsKeyNotFound(err), ShouldBeTrue)
		})
	})
}
