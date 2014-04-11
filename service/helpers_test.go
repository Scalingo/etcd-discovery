package service

import (
	"time"
)

func genHost(name ...interface{}) *Host {
	var strName string
	if len(name) == 1 {
		strName = name[0].(string)
	}

	// Empty if no arg, custom name otherways
	return NewHost(strName, "10000", "user", "secret")
}

func waitRegistration() {
	time.Sleep(200 * time.Millisecond)
}
