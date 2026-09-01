module github.com/Scalingo/etcd-discovery/v8

go 1.26

require (
	github.com/Scalingo/go-utils/errors/v3 v3.2.1
	github.com/Scalingo/go-utils/logger v1.12.2
	github.com/gofrs/uuid/v5 v5.5.1
	github.com/sirupsen/logrus v1.10.2
	github.com/stretchr/testify v1.12.1
	go.etcd.io/etcd/client/pkg/v3 v3.7.1
	// The latest versions of etcd have been migrated to go modules.
	// Since this change the version of the etcd client we are currently
	// using in this package has been named v2.
	// This does not mean that it does not work with the etcd server version 3.
	//
	// The package "go.etcd.io/etcd/client/v3" is a complete refactoring
	// of the client and uses grpc instead of http, that we currently use.
	go.etcd.io/etcd/client/v2 v2.305.33
	go.uber.org/mock v0.6.0
)

require (
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	go.etcd.io/etcd/api/v3 v3.7.1 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.28.0 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/sys v0.47.0 // indirect
	gopkg.in/errgo.v1 v1.0.1 // indirect
	gopkg.in/mgo.v2 v2.0.0-20190816093944-a6b53ec6cb22 // indirect
)
