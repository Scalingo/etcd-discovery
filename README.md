# Etcd Discovery v7.1.5

This is a Go package for managing services over the decentralized key-value store [etcd](https://github.com/etcd-io/etcd).

To install it:

```sh
go get github.com/Scalingo/etcd-discovery/v8/service
```

Registering a service consists of providing a public hostname or/and a private hostname:
* A public hostname resolves to an IP address usable from everywhere on the internet.
* A private hostname resolves to an IP address usable from the private network of the company.

## API

### Register a Service

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
    User:           gopassword.Generate(32),
    Password:       gopassword.Generate(32),
    Public: true,
    PrivateHostname: "node-1.internal.dev",
    PrivatePorts: service.ports{
      "http":  "8080",
      "https": "80443",
    },
  })
```

The `Public` attribute specifies whether a service has a public hostname or not. If set to false, setting the `Hostname` is equivalent to setting the `PrivateHostname`.

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

### Subscribe to New Service

When a service is added from another host, if you want your application to
notice it and communicating with it, it is necessary to watch these
notifications.

```go
newHosts := service.SubscribeNew("name_of_service")
for host := range newHosts {
  fmt.Println(host.Name, "has registered")
}
```

### Watch Down Services

```go
deadHosts := service.SubscribeDown("name_of_service")
for hostname := range deadHosts {
  fmt.Println(hostname, "is dead, RIP")
}
```

# Generate the Mocks

Generate the mocks with:

```shell
for interface in $(grep --extended-regexp --no-message --no-filename "type .* interface" ./service/* | cut -d " " -f 2)
do
  mockgen -destination service/servicemock/gomock_$(echo $interface | tr '[:upper:]' '[:lower:]').go -package servicemock github.com/Scalingo/etcd-discovery/v8/service $interface
done
```

# Release a New Version

Bump new version number in `CHANGELOG.md` and `README.md`.

Commit, tag and create a new release:

```sh
version="7.1.5"

git switch --create release/${version}
git add CHANGELOG.md README.md
git commit --message="Bump v${version}"
git push --set-upstream origin release/${version}
gh pr create --reviewer=Scalingo/team-ist --fill-first --base master
```

Once the pull request merged, you can tag the new release.

```sh
git tag v${version}
git push origin master v${version}
gh release create v${version}
```

The title of the release should be the version number and the text of the release is the same as the changelog.
