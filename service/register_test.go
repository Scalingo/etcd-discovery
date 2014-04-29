package service

import (
	"encoding/json"
	"fmt"
	"path"
	"testing"
	"time"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRegister(t *testing.T) {
	Convey("After registering service test", t, func() {
		host := genHost()
		Convey("It should be available with etcd", func() {
			err := Register("test_register", host, make(chan bool))
			waitRegistration()
			So(err, ShouldBeNil)

			res, err := Client().Get("/services/test_register/"+host.Name, false, false)
			So(err, ShouldBeNil)

			h := &Host{}
			json.Unmarshal([]byte(res.Node.Value), &h)

			So(path.Base(res.Node.Key), ShouldEqual, host.Name)
			So(h, ShouldResemble, host)
		})

		Convey(fmt.Sprintf("And the ttl must be < %d", HEARTBEAT_DURATION), func() {
			Register("test2_register", host, make(chan bool))
			waitRegistration()
			res, err := Client().Get("/services/test2_register/"+host.Name, false, false)
			So(err, ShouldBeNil)
			now := time.Now()
			duration := res.Node.Expiration.Sub(now)
			So(duration, ShouldBeLessThanOrEqualTo, HEARTBEAT_DURATION*time.Second)
		})

		Convey("After sending stop, the service should disappear", func() {
			stop := make(chan bool)
			host := genHost()
			Register("test3_register", host, stop)
			waitRegistration()
			stop <- true
			time.Sleep(HEARTBEAT_DURATION * 2 * time.Second)
			_, err := Client().Get("/services/test3_register/"+host.Name, false, false)
			So(err, ShouldNotBeNil)
		})
	})
}
