# Changelog

## To Be Released

* Bump github.com/gofrs/uuid from 3.3.0+incompatible to 3.4.0+incompatible #45

## v7.0.2

* Use github.com/gofrs/uuid for UUID generation
* Transform into a go module
* Use go.etcd.io/etcd instead of github.com/coreos/etcd

## v7.0.1

* Fix panic when loosing registration

## v7.0.0

* Use context.Context to handle lifecycle of the registration

## v6.0.2

* Fix: Auto-fill wont work if the hostname exists
* Reset the AfterIndex if we've lost the etcd-watcher

## v6.0.1

* Try to auto-fill the private_hostname and private_ports if the service is'nt public

## v6.0

* Add the Registration struct to help Register provide a more user firendly API

## v5.0.1

* Use the host private hostname if the hostname is not set for the uuid generation

## v5.0

* Add the notion of Public URL
* Synchronize host password
* Remove the update subscriber
* Improve the Get() API

## v4.1

* Add `name` field in service_info struct

## v4.0

* Add notion of service info, to define for instance if a service is critical or not

## v3.3

* Ability to get TLS certs/key/ca from environment directly without being stored in files
  `ETCD_TLS_INMEMORY=true`

## v3.2

* String method on array of Host

## v3.1

* Fix TLS client from environment variables

## v3.0

* Compatibility API v3
* Use officail go client

## v2.0

* Change way to represent interfaces :
  As a service can bind different interfaces (not only http),
  it has now to specify them.

## v1.0

* Simple implementation of Register/SubscribeNew/SubscribeDown
* Test for these features
