package service

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRegistration_Ready(t *testing.T) {
	host := genHost("test-registration-ready")
	service := Service{
		Name:     host.Name,
		Hostname: "public.dev",
	}

	Convey("When a new registration is created", t, func() {
		r, err := NewRegistration(host, service)
		So(err, ShouldBeNil)
		Reset(func() { So(r.Stop(), ShouldBeNil) })

		Convey("it must send false", func() {
			So(r.Ready(), ShouldBeFalse)
		})
	})

	// Convey("After a service registration", t, func() {
	// 	r, err := NewRegistration(host, service)
	// 	So(err, ShouldBeNil)
	// 	Reset(func() { So(r.Stop(), ShouldBeNil) })

	// 	Convey("it must send true", func() {
	// 		So(r.Ready(), ShouldBeTrue)
	// 	})
	// })
}

func TestRegistration_WaitRegistration(t *testing.T) {
	host := genHost("test-wait-registration")
	service := Service{
		Name:     host.Name,
		Hostname: "public.dev",
		User:     host.User,
		Password: host.Password,
	}

	Convey("It must wait for a service registration", t, func() {
		r, err := NewRegistration(host, service)
		So(err, ShouldBeNil)
		Reset(func() { So(r.Stop(), ShouldBeNil) })

		creds, err := r.Credentials()
		So(err, ShouldNotBeNil)

		r.WaitRegistration()
		creds, err = r.Credentials()
		So(err, ShouldBeNil)
		So(creds.User, ShouldEqual, host.User)
		So(creds.Password, ShouldEqual, host.Password)
	})
}

func TestRegistration_UUID(t *testing.T) {
	host := genHost("test-registration-uuid")
	service := Service{
		Name:     host.Name,
		Hostname: "public.dev",
	}

	Convey("It must send the original UUID", t, func() {
		r, err := NewRegistration(host, service)
		So(err, ShouldBeNil)
		So(r.UUID(), ShouldEqual, host.UUID)
		So(r.Stop(), ShouldBeNil)
	})
}

func TestRegistration_Credentials(t *testing.T) {
	host := genHost("test-registration-credentials")
	service := Service{
		Name:     host.Name,
		Hostname: "public.dev",
		User:     host.User,
		Password: host.Password,
	}
	host2 := genHost("test-registration-credentials")

	// First usecase has been tested in TestRegistration_WaitRegistration

	Convey("After a credential update", t, func() {
		r, err := NewRegistration(host, service)
		So(err, ShouldBeNil)
		Reset(func() { So(r.Stop(), ShouldBeNil) })

		r.WaitRegistration()
		c, err := r.Credentials()
		So(err, ShouldBeNil)
		So(c.User, ShouldEqual, host.User)
		So(c.Password, ShouldEqual, host.Password)

		service.User = host2.User
		service.Password = host2.Password
		r2, err := NewRegistration(host2, service)
		So(err, ShouldBeNil)
		Reset(func() { So(r2.Stop(), ShouldBeNil) })

		Convey("It should return the new credentials", func() {
			time.Sleep(500 * time.Millisecond)

			c, err := r.Credentials()
			So(err, ShouldBeNil)
			So(c.User, ShouldEqual, host2.User)
			So(c.Password, ShouldEqual, host2.Password)
		})
	})
}
