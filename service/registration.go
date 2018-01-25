package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	etcd "github.com/coreos/etcd/client"
)

// RegistrationWrapper wreap the uuid and the credential channel to provide a more user friendly API for the Register Method
type RegistrationWrapper interface {
	Ready() bool                       // Ready return strue if the service registred yet? This method should no be blocking
	WaitRegistration()                 // WaitRegistration wait for the first registration
	Credentials() (Credentials, error) // Credentials return the current credentials or an error if the service is not registred yet
	UUID() string                      // UUID return the host UUID
	Stop()                             // Stop the registration and free the resources (stop goroutines/close channels)
}

// Registration is the RegistrationWrapper implementation used by the Register method
type Registration struct {
	// Lifecycle management
	wg          *sync.WaitGroup
	cancel      context.CancelFunc
	shutdownErr error

	credsWatcher       chan Credentials
	clientCredsChan    chan Credentials
	ready              chan struct{}
	mutex              *sync.Mutex
	currentCredentials *Credentials

	// Data
	host        Host
	hostJSON    string
	service     Service
	serviceJSON string
}

// NewRegistration initialize the Registration struct
func NewRegistration(host Host, service Service) (*Registration, error) {
	hostBytes, err := json.Marshal(&host)
	if err != nil {
		return nil, fmt.Errorf("fail to serialize host to JSON: %v", err)
	}
	hostJSON := string(hostBytes)

	serviceBytes, err := json.Marshal(&service)
	if err != nil {
		return nil, fmt.Errorf("fail to serialize services to JSON: %v", err)
	}
	serviceJSON := string(serviceBytes)

	ctx, cancel := context.WithCancel(context.Background())

	r := &Registration{
		wg:     &sync.WaitGroup{},
		cancel: cancel,
		// Use a buffered channel as intial registratino should be storing the
		// credentials before the update routine is started.
		credsWatcher:       make(chan Credentials, 1),
		clientCredsChan:    make(chan Credentials),
		ready:              make(chan struct{}),
		mutex:              &sync.Mutex{},
		currentCredentials: nil,
		host:               host,
		hostJSON:           hostJSON,
		service:            service,
		serviceJSON:        serviceJSON,
	}

	r.wg.Add(2)
	go func() {
		defer r.wg.Done()
		r.currentCredentialsUpdater(ctx)
	}()

	go func() {
		defer r.wg.Done()
		initialID := r.initialRegistration(ctx)

		if host.Public {
			r.wg.Add(1)
			go func() {
				defer r.wg.Done()
				r.credentialChangeWatcher(ctx, initialID)
			}()
		}
		r.updateRoutine(ctx)
	}()

	return r, nil
}

func (r *Registration) Stop() error {
	r.cancel()
	r.wg.Wait()
	return r.shutdownErr
}

// WaitRegistration wait for the first registration to happen, meaning that the service is succesfulled registred to the etcd service
func (r *Registration) WaitRegistration() {
	<-r.ready
}

// Ready is a non blocking method that return true if the service is registred to the etcd service false otherwise
func (r *Registration) Ready() bool {
	select {
	case <-r.ready:
		return true
	default:
		return false
	}
}

// UUID of the current host
func (r *Registration) UUID() string {
	return r.host.UUID
}

// Credentials return the service credentials or an error if the service is not registred yet
func (r *Registration) Credentials() (Credentials, error) {
	r.mutex.Lock()
	cred := r.currentCredentials
	r.mutex.Unlock()
	if cred == nil {
		return Credentials{}, errors.New("Not ready")
	}
	return *cred, nil
}

func (r *Registration) currentCredentialsUpdater(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case newCredentials := <-r.clientCredsChan:
			r.mutex.Lock()
			r.currentCredentials = &newCredentials
			r.mutex.Unlock()

			if !r.Ready() {
				close(r.ready)
			}
		}
	}
}

func (r *Registration) initialRegistration(ctx context.Context) uint64 {
	var (
		id  uint64
		err error
	)
	// First register on ETCD the service and the host, retry until done
	// this operation can't fail.
	//
	// id is the current modification index of the service key.
	// this is used for the watcher.
	for {
		id, err = r.registerService()
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}

	for err != nil {
		err = r.registerHost()
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}

	r.credsWatcher <- Credentials{
		User:     r.service.User,
		Password: r.service.Password,
	}

	return id
}

func (r *Registration) updateRoutine(ctx context.Context) {
	ticker := time.NewTicker((HEARTBEAT_DURATION - 1) * time.Second)
	for {
		select {
		case <-ctx.Done():
			_, err := KAPI().Delete(context.Background(), r.hostKey(), &etcd.DeleteOptions{Recursive: false})
			if err != nil {
				logger.Println("fail to remove key", r.hostKey())
			}
			ticker.Stop()
			return
		case credentials := <-r.credsWatcher: // If the credentials have been changed
			// We update our cache
			r.host.User = credentials.User
			r.host.Password = credentials.Password
			r.service.User = credentials.User
			r.service.Password = credentials.Password

			// Re-marshal the host
			hostBytes, err := json.Marshal(&r.host)
			if err != nil {
				logger.Printf("fail to marshal JSON host %v: %v", r.host, err)
			}
			r.hostJSON = string(hostBytes)

			// synchro the host informations
			err = r.registerHost()
			if err != nil {
				logger.Printf("fail to register host %v: %v", r.host, err)
			}

			// and transmit them to the client
			select {
			case r.clientCredsChan <- credentials:
			default:
				// default case if the context is canceled, no one is listening to the
				// channel
			}
		case <-ticker.C:
			err := r.registerHost()
			// If for any random reason, there is an error,
			// we retry every second until it's ok.
			for err != nil {
				logger.Printf("lost registration of '%v': %v (%v)", r.service, err, Client().Endpoints())
				time.Sleep(1 * time.Second)

				err = r.registerHost()
				if err == nil {
					logger.Printf("recover registration of '%v'", r.service)
				}
			}
		}
	}
}

func (r *Registration) credentialChangeWatcher(ctx context.Context, initialID uint64) {
	var (
		service Service
	)

	watcher := KAPI().Watcher(r.serviceKey(), &etcd.WatcherOptions{
		AfterIndex: initialID,
	})
	for {
		resp, err := watcher.Next(ctx)
		if err == context.Canceled {
			return
		}

		if err != nil {
			// We've lost the connexion to etcd. Sleep 1s and retry
			logger.Printf("lost watcher of '%v': '%v' (%v)", r.serviceKey(), err, Client().Endpoints())
			time.Sleep(1 * time.Second)
			continue
		}

		err = json.Unmarshal([]byte(resp.Node.Value), &service)
		if err != nil {
			logger.Printf("invalid JSON for service '%v': '%v' (%v)", r.serviceKey(), err, Client().Endpoints())
			continue
		}

		// We've got the modification, send it to the register agent
		r.credsWatcher <- Credentials{
			User:     service.User,
			Password: service.Password,
		}
	}
}

func (r *Registration) registerHost() error {
	_, err := KAPI().Set(context.Background(), r.hostKey(), r.hostJSON, &etcd.SetOptions{TTL: HEARTBEAT_DURATION * time.Second})
	if err != nil {
		return fmt.Errorf("unable to register host: %v", err)
	}
	return nil
}

func (r *Registration) registerService() (uint64, error) {
	key, err := KAPI().Set(context.Background(), r.serviceKey(), r.serviceJSON, nil)
	if err != nil {
		return 0, fmt.Errorf("unable to register service: %v", err)
	}

	return key.Node.ModifiedIndex, nil
}

func (r *Registration) hostKey() string {
	return fmt.Sprintf("/services/%s/%s", r.service.Name, r.host.UUID)
}

func (r *Registration) serviceKey() string {
	return fmt.Sprintf("/services_infos/%s", r.service.Name)
}
