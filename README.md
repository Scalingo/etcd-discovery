Etcd-discovery
==============

This is a golang package for managing services over the decentralized key-value store etcd

> http://github.com/coreos/etcd

To install it:

`go get github.com/Scalingo/etcd-discovery/service`

API
---

### Register a service

```go
/*
 * First argument is the name of the service
 * Then a struct containing the service informations
 * The registeration will stop when the context will be canceled.
 * It will return the service uuid and a channel which will send back any modifications made to the service by the other host of the same service. This is usefull for credential synchronisation.
 */
registration, err := service.Register(
  "my-service",
  &service.Host{
    Hostname: "public-domain.dev",
    Ports: service.Ports{
      "http":  "80",
      "https": "443",
    },
    Critical:       true,
    User:           gopassword.Generate(10),
    Password:       gopassword.Generate(10),
    Public: true,
    PrivateHostname: "node-1.internal.dev",
    PrivatePorts: service.ports{
      "http":  "8080",
      "https": "80443",
    },
  },
)
if err != nil {
  // Register return an error if it fails to initialize
  // Then, it reconnects automatically to etcd etc. if required
}

// ...

// To release properly resources
err = registration.Stop()
if err != nil {
  // Do something with error
}
```

This will create two different etcd keys:

* `/services/name_of_service/you-service-uuid-node-1.internal.dev` containing:
```json
{
   "name": "public-domain.dev",
   "service_name": "my-service",
   "user": "user",
   "password": "secret",
   "public": true,
   "private_hostname": "node-1.internal.dev",
   "private_ports": {
      "http": "8080",
      "https": "80443"
   },
   "critcal": true,
   "uuid": "your-service-uuid-node-1.internal.dev",
   "ports":{
      "http":"80",
      "https":"443"
   }
}
```

* `/service_infos/name_of_service` containing:
```json
{
   "name": "my-service",
   "critical": true,
   "hostname": "public-domain.dev",
   "user": "user",
   "password": "password",
   "ports": {
      "http": "80",
      "https": "443"
   },
   "public": true
}
```

### Subscribe to new service

When a service is added from another host, if you want your application to
notice it and communicating with it, it is necessary to watch these
notifications.

```go
newHosts := service.SubscribeNew("name_of_service")
for host := range newHosts {
  fmt.Println(host.Name, "has registered")
}
```

### Watch down services

```go
deadHosts := service.SubscribeDown("name_of_service")
for hostname := range deadHosts {
  fmt.Println(hostname, "is dead, RIP")
}
```
