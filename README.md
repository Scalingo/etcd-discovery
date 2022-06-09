# Etcd Discovery v7.1.1

This is a golang package for managing services over the decentralized key-value store etcd.

> https://github.com/etcd-io/etcd

To install it:

`go get github.com/Scalingo/etcd-discovery/service/v7`

## API

### Register a service

```go
/*
 * First argument is the name of the service
 * Then a struct containing the service informations
 * The registeration will stop when the context will be canceled.
 * It will return the service uuid and a channel which will send back any modifications made to the service by the other host of the same service. This is usefull for credential synchronisation.
 */
ctx, cancel := context.WithCancel(context.Background())
registration := service.Register(
  ctx,
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
  })
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

# Generate the mocks

Generate the mocks with:

```shell
for interface in $(grep --extended-regexp --no-message --no-filename "type .* interface" ./service/* | cut -d " " -f 2)
do
  mockgen -destination service/servicemock/gomock_$(echo $interface | tr '[:upper:]' '[:lower:]').go -package servicemock github.com/Scalingo/etcd-discovery/v7/service $interface
done
```

## Release a New Version

Bump new version number in `CHANGELOG.md` and `README.md`.

Commit, tag and create a new release:

```shell
git add CHANGELOG.md README.md
git commit -m "Bump v7.1.1"
git tag v7.1.1
git push origin master
git push --tags
hub release create v7.1.1
```

The title of the release should be the version number and the text of the release is the same as the changelog.
