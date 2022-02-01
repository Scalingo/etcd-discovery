module github.com/Scalingo/etcd-discovery/v7

go 1.16

require (
	github.com/gofrs/uuid v4.2.0+incompatible
	github.com/golang/mock v1.6.0
	github.com/smartystreets/goconvey v1.7.2
	go.etcd.io/etcd/client/pkg/v3 v3.5.2
	// The latest versions of etcd have been migrated to go modules.
	// Since this change the version of the etcd client we are currently
	// using in this package has been named v2.
	// This does not mean that it does not work with the etcd server version 3.
	//
	// The package "go.etcd.io/etcd/client/v3" is a complete refactoring
	// of the client and uses grpc instead of http, that we currently use.
	go.etcd.io/etcd/client/v2 v2.305.2
	gopkg.in/errgo.v1 v1.0.1
)
