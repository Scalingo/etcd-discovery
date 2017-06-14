package service

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRegistrationReady(t *testing.T) {
	Convey("When a new registration is created", t, func() {
		r := NewRegistration("1234", make(chan Credentials))
		Convey("it must send false", func() {
			So(r.Ready(), ShouldBeFalse)
		})
	})

	Convey("After a service registration", t, func() {
		cred := make(chan Credentials)
		r := NewRegistration("1234", cred)
		cred <- Credentials{
			User:     "Moi",
			Password: "Lui",
		}

		Convey("it must send true", func() {
			So(r.Ready(), ShouldBeTrue)
		})
	})
}

func TestWaitRegistration(t *testing.T) {
	Convey("It must wait for a service registration", t, func() {
		order := make([]bool, 2)
		cred := make(chan Credentials)
		r := NewRegistration("1234", cred)
		registrationChan := make(chan bool)
		go func() {
			for {
				r.WaitRegistration()
				registrationChan <- true
			}
		}()
		timer := time.NewTimer(1 * time.Second)
		for i := 0; i < 2; i++ {
			select {
			case <-timer.C:
				order[i] = true
			case <-registrationChan:
				order[i] = false
			}

			timer.Reset(1 * time.Second)
			cred <- Credentials{
				User:     "Salut",
				Password: "Toi",
			}
		}

		So(order[0], ShouldBeTrue)
		So(order[1], ShouldBeFalse)
	})
}

func TestUUID(t *testing.T) {
	Convey("It must send the original UUID", t, func() {
		r := NewRegistration("test-test-123", make(chan Credentials))
		So(r.UUID(), ShouldEqual, "test-test-123")
	})
}

func TestCredentials(t *testing.T) {
	Convey("Before a service registration", t, func() {
		r := NewRegistration("test", make(chan Credentials))
		Convey("It should return an error", func() {
			_, err := r.Credentials()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "Not ready")
		})
	})

	Convey("After a service registration", t, func() {
		cred := make(chan Credentials)
		r := NewRegistration("test", cred)
		cred <- Credentials{
			User:     "1",
			Password: "2",
		}

		Convey("It should return the new credentials", func() {
			c, err := r.Credentials()
			So(err, ShouldBeNil)
			So(c.User, ShouldEqual, "1")
			So(c.Password, ShouldEqual, "2")
		})
	})

	Convey("After a credential update", t, func() {
		cred := make(chan Credentials)
		r := NewRegistration("test", cred)
		cred <- Credentials{
			User:     "1",
			Password: "2",
		}
		cred <- Credentials{
			User:     "3",
			Password: "4",
		}

		Convey("It should return the new credentials", func() {
			time.Sleep(1 * time.Second)
			c, err := r.Credentials()
			So(err, ShouldBeNil)
			So(c.User, ShouldEqual, "3")
			So(c.Password, ShouldEqual, "4")
		})
	})
}
