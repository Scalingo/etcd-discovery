package service

func genHost(name string) Host {
	return Host{
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
		UUID:     "1234",
	}
}

func genService(name string) *Service {
	return &Service{
		Name:     name,
		Critical: true,
		Hostname: "public.dev",
		User:     "user",
		Password: "password",
		Ports: Ports{
			"http": "80",
		},
		Public: true,
	}
}
