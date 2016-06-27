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
 * Then a struct containing some information
    * The Name attribute is optional, the variable is taken from the environment variable HOSTNAME or from os.Hostname()
 * The stop channel exists if you want to be able to stop the registeration
 */
stop := make(chan struct{})
service.Register(
  "name_of_service",
  &service.Host{
    Name: "hostname_"
    User: "user",
    Password: "secret",
    Ports: map[string]string{
      "http":"1234",
    },
  },
  stop,
)
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
