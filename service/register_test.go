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
		stop := make(chan bool)
		host := genHost()
		Register("test_register", host, stop)

		Convey("It should be available with etcd", func() {
			res, err := client.Get("/services/test_register/"+hostname, false, false)

			h := &Host{}
			json.Unmarshal([]byte(res.Node.Value), &h)

			So(path.Base(res.Node.Key), ShouldEqual, host.Name)
			So(h, ShouldResemble, host)
			So(err, ShouldBeNil)
		})
		stop <- true

		Register("test2_register", genHost(), stop)
		Convey(fmt.Sprintf("And the ttl must be < %d", HEARTBEAT_DURATION), func() {
			res, _ := client.Get("/services/test2_register/"+hostname, false, false)
			now := time.Now()
			duration := res.Node.Expiration.Sub(now)
			So(duration, ShouldBeLessThanOrEqualTo, HEARTBEAT_DURATION*time.Second)
		})
		stop <- true

		Convey("After sending stop, the service should disappear", func() {
			time.Sleep(HEARTBEAT_DURATION * 2 * time.Second)
			_, err := client.Get("/services/test2_register/"+hostname, false, false)
			So(err, ShouldNotBeNil)
		})
	})
}
