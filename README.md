# Etcd Discovery v7.1.5

This is a Go package for managing services over the decentralized key-value store [etcd](https://github.com/etcd-io/etcd).

To install it:

```sh
go get github.com/Scalingo/etcd-discovery/v7/service
```

Registering a service consists of providing a public hostname or/and a private hostname:
* A public hostname resolves to an IP address usable from everywhere on the internet.
* A private hostname resolves to an IP address usable from the private network of the company.

## API

### Register a Service

```go
/*
 * First argument is the name of the service.
 * Then a struct containing the service information.
 * The registration will stop when the context will be canceled.
 * It will return the service uuid and a channel which will send back any modifications made to the
   service by the other host of the same service. This is useful for credential synchronization.
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
    Shard:          "shard-0",
    User:           gopassword.Generate(32),
    Password:       gopassword.Generate(32),
    Public: true,
    PrivateHostname: "node-1.internal.dev",
    PrivatePorts: service.Ports{
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
   "critical": true,
   "shard": "shard-0",
   "uuid": "your-service-uuid-node-1.internal.dev",
   "ports":{
      "http":"80",
      "https":"443"
   }
}
```

* `/services_infos/name_of_service` containing:

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

Shard information is stored per host under `/services/<name>/<uuid>`. It is intentionally not stored in
`/services_infos/<name>`, because different instances of the same service may register on different shards.

### Query a Service

Use `Get` to query all hosts for a service:

```go
hosts, err := service.Get("my-service").All()
url, err := service.Get("my-service").URL("http", "/health")
```

Use `GetForShard` to restrict host lookups to a specific shard:

```go
hosts, err := service.GetForShard("my-service", "shard-0").All()
url, err := service.GetForShard("my-service", "shard-0").URL("http", "/health")
```

`Service.All`, `Service.First`, `Service.One`, and `Service.URL` now take a `service.QueryOptions`
argument. Pass `service.QueryOptions{}` when no filtering is needed.

```go
s, err := service.Get("my-service").Service()
if err != nil {
  return err
}

host, err := s.First(service.QueryOptions{Shard: "shard-0"})
url, err := s.URL("http", "/health", service.QueryOptions{})
```

### Subscribe to New Service

When a service is added from another host, if you want your application to
notice it and communicating with it, it is necessary to watch these
notifications.

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

newHosts, errs := service.SubscribeNew(ctx, "name_of_service")
for host := range newHosts {
  fmt.Println(host.Name, "has registered")
}

if err := <-errs; err != nil {
  fmt.Println("watch failed:", err)
}
```

### Watch Down Services

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

deadHosts, errs := service.SubscribeDown(ctx, "name_of_service")
for hostname := range deadHosts {
  fmt.Println(hostname, "is dead, RIP")
}

if err := <-errs; err != nil {
  fmt.Println("watch failed:", err)
}
```

# Generate the Mocks

Generate the mocks with:

```shell
gomock_generator
```

> [!NOTE]
> `gomock_generator` binary should be installed: https://github.com/Scalingo/go-utils/tree/master/gomock_generator

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
