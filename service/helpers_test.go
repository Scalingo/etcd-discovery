package service

func genHost(name string) *Host {
	// Empty if no arg, custom name otherways
	host, _ := NewHost(name, Ports{"http": "10000"}, "user", "secret")
	return host
}
