package service

import (
	"context"
	"errors"
	"sync"
)

// RegistrationWrapper wreap the uuid and the credential channel to provide a more user friendly API for the Register Method
type RegistrationWrapper interface {
	Ready() bool                                // Ready return strue if the service registred yet? This method should no be blocking
	WaitRegistration(ctx context.Context) error // WaitRegistration wait for the first registration
	Credentials() (Credentials, error)          // Credentials return the current credentials or an error if the service is not registred yet
	UUID() string                               // UUID return the host UUID
}

// Registration is the RegistrationWrapper implementation used by the Register method
type Registration struct {
	credChan        chan Credentials
	readyChan       chan struct{}
	ready           bool
	waitErr         error
	uuid            string
	mutex           sync.Mutex
	signalReadyOnce sync.Once
	curCredentials  *Credentials
}

// NewRegistration initialize the Registration struct
func NewRegistration(ctx context.Context, uuid string, cred chan Credentials) *Registration {
	r := &Registration{
		credChan:        cred,
		readyChan:       make(chan struct{}),
		ready:           false,
		waitErr:         nil,
		uuid:            uuid,
		mutex:           sync.Mutex{},
		signalReadyOnce: sync.Once{},
		curCredentials:  nil,
	}
	go r.worker(ctx)
	return r
}

// WaitRegistration wait for the first registration to happen, meaning that the service is successfully registered on the etcd server.
// It also returns if the caller context is canceled before the first credentials arrive.
func (w *Registration) WaitRegistration(ctx context.Context) error {
	if w.Ready() {
		return nil
	}

	select {
	case <-ctx.Done():
		// This context only controls the current wait call. The background worker
		// keeps listening for the registration until its own parent context ends.
		return ctx.Err()
	case <-w.readyChan:
		w.mutex.Lock()
		err := w.waitErr
		w.mutex.Unlock()
		return err
	}
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
			// Unblock WaitRegistration callers even if the registration never reached etcd.
			w.signalReady(ctx.Err())
			return
		case newCred := <-w.credChan:
			w.mutex.Lock()
			w.curCredentials = &newCred
			if !w.ready {
				// The first credentials mark the registration as usable. Later updates only refresh the cache.
				w.ready = true
			}
			w.mutex.Unlock()
			w.signalReady(nil)
		}
	}
}

func (w *Registration) signalReady(err error) {
	w.signalReadyOnce.Do(func() {
		// Registration can become "ready" either because the first credentials arrived
		// or because the parent context ended first. Close the wait channel only once
		// so concurrent callers observe a single outcome.
		w.mutex.Lock()
		w.waitErr = err
		w.mutex.Unlock()
		close(w.readyChan)
	})
}
