package service

func genHost(name ...interface{}) *Host {
	var strName string
	if len(name) == 1 {
		strName = name[0].(string)
	}

	// Empty if no arg, custom name otherways
	return &Host{
		Name:     strName,
		User:     "user",
		Password: "secret",
		Port:     "10000",
	}
}
