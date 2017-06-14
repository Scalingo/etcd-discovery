package service

import (
	"errors"
	"sync"
)

type RegistrationWrapper interface {
	Ready() bool
	WaitRegistration()
	Credentials() (Credentials, error)
	UUID() string
}

type Registration struct {
	credChan       chan (Credentials)
	readyChan      chan bool
	ready          bool
	uuid           string
	mutex          sync.Mutex
	curCredentials *Credentials
}

func NewRegistration(uuid string, cred chan Credentials) *Registration {
	r := &Registration{
		credChan:       cred,
		readyChan:      make(chan bool, 1),
		ready:          false,
		uuid:           uuid,
		mutex:          sync.Mutex{},
		curCredentials: nil,
	}
	go r.worker()
	return r
}

func (w *Registration) WaitRegistration() {
	<-w.readyChan
}

func (w *Registration) Ready() bool {
	w.mutex.Lock()
	ready := w.ready
	w.mutex.Unlock()
	return ready
}

func (w *Registration) UUID() string {
	return w.uuid
}

func (w *Registration) Credentials() (Credentials, error) {
	w.mutex.Lock()
	cred := w.curCredentials
	w.mutex.Unlock()
	if cred == nil {
		return Credentials{}, errors.New("Not ready")
	}
	return *cred, nil
}

func (w *Registration) worker() {
	for {
		newCred := <-w.credChan
		w.mutex.Lock()
		w.curCredentials = &newCred
		if !w.ready {
			w.readyChan <- true
		}
		w.ready = true
		w.mutex.Unlock()
	}
}
