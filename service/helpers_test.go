package service

func genHost() *Host {
		return &Host{
			User: "user",
			Password: "secret",
			Port: "10000",
		}
}
