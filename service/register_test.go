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
			host.Name = "test_register"
			r, err := Register("test_register", host)
			So(err, ShouldBeNil)
			r.WaitRegistration()
			Reset(func() { So(r.Stop(), ShouldBeNil) })

			uuid := r.UUID()
			res, err := KAPI().Get(context.Background(), "/services/test_register/"+uuid, &etcd.GetOptions{})
			So(err, ShouldBeNil)

			h := &Host{}
			json.Unmarshal([]byte(res.Node.Value), &h)

			So(path.Base(res.Node.Key), ShouldEqual, uuid)
			host.UUID = h.UUID
			So(h, ShouldResemble, &host)
		})

		Convey(fmt.Sprintf("And the ttl must be < %d", HEARTBEAT_DURATION), func() {
			r, err := Register("test2_register", host)
			So(err, ShouldBeNil)
			r.WaitRegistration()
			Reset(func() { So(r.Stop(), ShouldBeNil) })

			uuid := r.UUID()
			res, err := KAPI().Get(context.Background(), "/services/test2_register/"+uuid, &etcd.GetOptions{})
			So(err, ShouldBeNil)
			now := time.Now()
			duration := res.Node.Expiration.Sub(now)
			So(duration, ShouldBeLessThanOrEqualTo, HEARTBEAT_DURATION*time.Second)
		})

		Convey("And the serivce infos must be set", func() {
			infos := &Service{
				Name:     "test3_register",
				Hostname: "public.dev",
				User:     "user",
				Password: "password",
				Ports: Ports{
					"http": "10000",
				},
				Public:   true,
				Critical: true,
			}

			r, err := Register("test3_register", host)
			So(err, ShouldBeNil)
			r.WaitRegistration()
			Reset(func() { So(r.Stop(), ShouldBeNil) })

			res, err := KAPI().Get(context.Background(), "/services_infos/test3_register", &etcd.GetOptions{})
			So(err, ShouldBeNil)

			service := &Service{}
			json.Unmarshal([]byte(res.Node.Value), &service)

			So(service, ShouldResemble, infos)
		})

		Convey("After cancelling context, the service should disappear", func() {
			host := genHost("test-disappear")
			r, err := Register("test4_register", host)
			So(err, ShouldBeNil)
			r.WaitRegistration()
			So(r.Stop(), ShouldBeNil)

			time.Sleep(100 * time.Millisecond)
			_, err = KAPI().Get(context.Background(), "/services/test4_register/"+host.Name, &etcd.GetOptions{})
			So(etcd.IsKeyNotFound(err), ShouldBeTrue)
		})

		Convey("When the privatehostname is not set, it must take the node hostname", func() {
			host := genHost("HelloWorld")
			host.PrivateHostname = ""
			r, err := Register("hello_world", host)
			So(err, ShouldBeNil)
			So(r.UUID(), ShouldEndWith, hostname)
			So(r.Stop(), ShouldBeNil)
		})
		Convey("When the private ports is not set and the service is private, it should take the public_ports", func() {
			host := genHost("HelloWorld2")
			host.Public = false
			host.PrivatePorts = Ports{}

			r, err := Register("hello_world2", host)
			So(err, ShouldBeNil)
			r.WaitRegistration()
			Reset(func() { So(r.Stop(), ShouldBeNil) })

			h, err := Get("hello_world2").First().Host()
			So(err, ShouldBeNil)
			So(len(h.PrivatePorts), ShouldEqual, 1)
		})
	})
}

func TestWatcher(t *testing.T) {
	Convey("With two instances of the same service", t, func() {
		host1 := genHost("test-watcher-1")
		host2 := genHost("test-watcher-1")

		host1.User = "host1"
		host1.Password = "password1"

		host2.User = "host2"
		host2.Password = "password2"

		r1, err := Register("test-watcher", host1)
		So(err, ShouldBeNil)
		r1.WaitRegistration()
		Reset(func() { So(r1.Stop(), ShouldBeNil) })

		cred1, _ := r1.Credentials()
		So(cred1.User, ShouldEqual, "host1")
		So(cred1.Password, ShouldEqual, "password1")

		r2, err := Register("test-watcher", host2)
		So(err, ShouldBeNil)
		r2.WaitRegistration()
		Reset(func() { So(r2.Stop(), ShouldBeNil) })

		Convey("it should send the new passwords", func() {
			cred2, _ := r2.Credentials()
			So(cred2.User, ShouldEqual, "host2")
			So(cred2.Password, ShouldEqual, "password2")

			time.Sleep(1 * time.Second)
			cred1, _ = r1.Credentials()
			So(cred1.User, ShouldEqual, "host2")
			So(cred1.Password, ShouldEqual, "password2")
		})

		Convey("it should update the host key", func() {
			for _, r := range []*Registration{r1, r2} {
				res, err := KAPI().Get(context.Background(), "/services/test-watcher/"+r.UUID(), &etcd.GetOptions{})
				So(err, ShouldBeNil)

				h := &Host{}
				json.Unmarshal([]byte(res.Node.Value), &h)
				So(h.User, ShouldEqual, "host2")
				So(h.Password, ShouldEqual, "password2")
			}

		})
	})
}
