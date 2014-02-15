package service

import (
	"github.com/coreos/go-etcd/etcd"
)

func IsKeyAlreadyExistError(err error) bool {
	etcdErr, ok := err.(*etcd.EtcdError)
	return ok && etcdErr.ErrorCode == 105
}

func IsKeyNotFoundError(err error) bool {
	etcdErr, ok := err.(*etcd.EtcdError)
	return ok && etcdErr.ErrorCode == 100
}
