module github.com/Scalingo/etcd-discovery/v7

go 1.20

require (
	github.com/gofrs/uuid v4.4.0+incompatible
	github.com/golang/mock v1.6.0
	github.com/smartystreets/goconvey v1.8.1
	github.com/stretchr/testify v1.8.4
	go.etcd.io/etcd/client/pkg/v3 v3.5.9
	// The latest versions of etcd have been migrated to go modules.
	// Since this change the version of the etcd client we are currently
	// using in this package has been named v2.
	// This does not mean that it does not work with the etcd server version 3.
	//
	// The package "go.etcd.io/etcd/client/v3" is a complete refactoring
	// of the client and uses grpc instead of http, that we currently use.
	go.etcd.io/etcd/client/v2 v2.305.9
	gopkg.in/errgo.v1 v1.0.1
)

require (
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/gopherjs/gopherjs v1.18.0-beta3 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/modern-go/concurrent v0.0.0-20180228061459-e0a39a4cb421 // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/smarty/assertions v1.15.0 // indirect
	go.etcd.io/etcd/api/v3 v3.5.9 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.24.0 // indirect
	golang.org/x/sys v0.9.0 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
