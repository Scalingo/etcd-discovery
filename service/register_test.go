package service

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestRegister(t *testing.T) {
	Convey("After registering service test", t, func() {
		stop := make(chan bool)
		Register("test", genHost(), stop)

		Convey("It should be available with etcd", func() {
			_, err := client.Get("/services/test/"+hostname, false, false)
			So(err, ShouldBeNil)
			_, err = client.Get("/services/test/"+hostname+"/user", false, false)
			So(err, ShouldBeNil)
			_, err = client.Get("/services/test/"+hostname+"/password", false, false)
			So(err, ShouldBeNil)
			_, err = client.Get("/services/test/"+hostname+"/port", false, false)
			So(err, ShouldBeNil)
		})
		stop <- true

		Register("test2", genHost(), stop)
		Convey(fmt.Sprintf("And the ttl must be < %d", HEARTBEAT_DURATION), func() {
			res, _ := client.Get("/services/test2/"+hostname, false, false)
			now := time.Now()
			duration := res.Node.Expiration.Sub(now)
			So(duration, ShouldBeLessThanOrEqualTo, HEARTBEAT_DURATION*time.Second)
		})
		stop <- true

		Convey("After sending stop, the service should disappear", func() {
			time.Sleep(HEARTBEAT_DURATION * time.Second)
			_, err := client.Get("/services/test/"+hostname, false, false)
			So(err, ShouldNotBeNil)
		})
	})
}
