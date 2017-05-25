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
    * The Name attribute is optional, the variable is taken from the environment variable HOSTNAME or from os.Hostname()
 * Then a struct containint the service informations.
    * The PublicHostname, User and Password will be fetched from the host informations if empty
 * The stop channel exists if you want to be able to stop the registeration
 *
 * It will return a channel which will send back any modifications made to the service by the other host of the same service. This is usefull for credential synchronisation.
 */
stopper := make(chan struct{})
changes := service.Register(
  "mon-service",
  &service.Host{
    Name: "172.17.0.1",
    Ports: service.Ports{
      "http":  "8080",
      "https": "80443",
    },
  },
  &service.Infos{
    Critical:       true,
    User:           gopassword.Generate(10),
    Password:       gopassword.Generate(10),
    PublicHostname: "scalingo.dev",
  }, stopper)
```

This will create two different etcd keys:

* `/services/name_of_service/hostname_` containing:
```json
{
   "name": "hostname_",
   "user": "user",
   "password": "secret",
   "ports":{
      "http":"1234"
   }
}
```

* `/service_infos/name_of_service` containing:
```json
{
   "name": "name_of_service",
   "critical": false
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
