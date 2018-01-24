package service

import (
	"context"
	"errors"
	"sync"
)

// RegistrationWrapper wreap the uuid and the credential channel to provide a more user friendly API for the Register Method
type RegistrationWrapper interface {
	Ready() bool                       // Ready return strue if the service registred yet? This method should no be blocking
	WaitRegistration()                 // WaitRegistration wait for the first registration
	Credentials() (Credentials, error) // Credentials return the current credentials or an error if the service is not registred yet
	UUID() string                      // UUID return the host UUID
}

// Registration is the RegistrationWrapper implementation used by the Register method
type Registration struct {
	credChan       chan (Credentials)
	readyChan      chan bool
	ready          bool
	uuid           string
	mutex          sync.Mutex
	curCredentials *Credentials
}

// NewRegistration initialize the Registration struct
func NewRegistration(ctx context.Context, uuid string, cred chan Credentials) *Registration {
	r := &Registration{
		credChan:       cred,
		readyChan:      make(chan bool, 1),
		ready:          false,
		uuid:           uuid,
		mutex:          sync.Mutex{},
		curCredentials: nil,
	}
	go r.worker(ctx)
	return r
}

// WaitRegistration wait for the first registration to happen, meaning that the service is succesfulled registred to the etcd service
func (w *Registration) WaitRegistration() {
	if w.Ready() {
		return
	}
	<-w.readyChan
}

// Ready is a non blocking method that return true if the service is registred to the etcd service false otherwise
func (w *Registration) Ready() bool {
	w.mutex.Lock()
	ready := w.ready
	w.mutex.Unlock()
	return ready
}

// UUID of the current host
func (w *Registration) UUID() string {
	return w.uuid
}

// Credentials return the service credentials or an error if the service is not registred yet
func (w *Registration) Credentials() (Credentials, error) {
	w.mutex.Lock()
	cred := w.curCredentials
	w.mutex.Unlock()
	if cred == nil {
		return Credentials{}, errors.New("Not ready")
	}
	return *cred, nil
}

func (w *Registration) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case newCred := <-w.credChan:
			w.mutex.Lock()
			w.curCredentials = &newCred
			if !w.ready {
				w.readyChan <- true
			}
			w.ready = true
			w.mutex.Unlock()
		}
	}
}
