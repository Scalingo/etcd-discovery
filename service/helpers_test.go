package service

func genHost(name string) *Host {
	// Empty if no arg, custom name otherways
	return &Host{
		Name:     "test-service",
		Hostname: "public.dev",
		Ports: Ports{
			"http": "10000",
		},
		User:            "user",
		Password:        "password",
		Public:          true,
		PrivateHostname: name + "-private.dev",
		PrivatePorts: Ports{
			"http": "20000",
		},
		Critical: true,
		Uuid:     "1234",
	}
}
