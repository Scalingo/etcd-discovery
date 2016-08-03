package service

import (
	etcd "github.com/coreos/etcd/client"
)

func IsKeyAlreadyExistError(err error) bool {
	etcdErr, ok := err.(*etcd.Error)
	return ok && etcdErr.Code == 105
}
