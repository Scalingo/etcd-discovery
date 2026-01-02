package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistrationReady(t *testing.T) {
	t.Run("When a new registration is created", func(t *testing.T) {
		r := NewRegistration(t.Context(), "1234", make(chan Credentials))

		t.Run("it must send false", func(t *testing.T) {
			require.False(t, r.Ready())
		})
	})

	t.Run("After a service registration", func(t *testing.T) {
		cred := make(chan Credentials)
		r := NewRegistration(t.Context(), "1234", cred)
		cred <- Credentials{
			User:     "Moi",
			Password: "Lui",
		}

		t.Run("it must send true", func(t *testing.T) {
			require.True(t, r.Ready())
		})
	})
}

func TestWaitRegistration(t *testing.T) {
	t.Run("It must wait for a service registration", func(t *testing.T) {
		order := make([]bool, 2)
		cred := make(chan Credentials)
		r := NewRegistration(t.Context(), "1234", cred)
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

		assert.True(t, order[0])
		assert.False(t, order[1])
	})
}

func TestUUID(t *testing.T) {
	t.Run("It must send the original UUID", func(t *testing.T) {
		r := NewRegistration(t.Context(), "test-test-123", make(chan Credentials))
		assert.Equal(t, "test-test-123", r.UUID())
	})
}

func TestCredentials(t *testing.T) {
	t.Run("Before a service registration", func(t *testing.T) {
		r := NewRegistration(t.Context(), "test", make(chan Credentials))
		t.Run("It should return an error", func(t *testing.T) {
			_, err := r.Credentials()
			require.EqualError(t, err, "not ready")
		})
	})

	t.Run("After a service registration", func(t *testing.T) {
		cred := make(chan Credentials)
		r := NewRegistration(t.Context(), "test", cred)
		cred <- Credentials{
			User:     "1",
			Password: "2",
		}

		t.Run("It should return the new credentials", func(t *testing.T) {
			c, err := r.Credentials()
			require.NoError(t, err)
			assert.Equal(t, "1", c.User)
			assert.Equal(t, "2", c.Password)
		})
	})

	t.Run("After a credential update", func(t *testing.T) {
		cred := make(chan Credentials)
		r := NewRegistration(t.Context(), "test", cred)
		cred <- Credentials{
			User:     "1",
			Password: "2",
		}
		cred <- Credentials{
			User:     "3",
			Password: "4",
		}

		t.Run("It should return the new credentials", func(t *testing.T) {
			time.Sleep(1 * time.Second)
			c, err := r.Credentials()
			require.NoError(t, err)
			assert.Equal(t, "3", c.User)
			assert.Equal(t, "4", c.Password)
		})
	})
}
