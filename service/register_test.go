package service

import (
	"encoding/json"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"path"
	"testing"
	"time"
)

func TestRegister(t *testing.T) {
	Convey("After registering service test", t, func() {
		host := genHost()
		Convey("It should be available with etcd", func() {
			err := Register("test_register", host, make(chan bool))
			So(err, ShouldBeNil)

			res, err := client.Get("/services/test_register/"+host.Name, false, false)

			h := &Host{}
			json.Unmarshal([]byte(res.Node.Value), &h)

			So(path.Base(res.Node.Key), ShouldEqual, host.Name)
			So(h, ShouldResemble, host)
			So(err, ShouldBeNil)
		})

		Convey(fmt.Sprintf("And the ttl must be < %d", HEARTBEAT_DURATION), func() {
			Register("test2_register", host, make(chan bool))
			res, _ := client.Get("/services/test2_register/"+host.Name, false, false)
			now := time.Now()
			duration := res.Node.Expiration.Sub(now)
			So(duration, ShouldBeLessThanOrEqualTo, HEARTBEAT_DURATION*time.Second)
		})

		Convey("After sending stop, the service should disappear", func() {
			stop := make(chan bool)
			host := genHost()
			Register("test3_register", host, stop)
			stop<-true
			time.Sleep(HEARTBEAT_DURATION * 2 * time.Second)
			_, err := client.Get("/services/test3_register/"+host.Name, false, false)
			So(err, ShouldNotBeNil)
		})
	})
}
