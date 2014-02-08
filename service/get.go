package service

func Get(service string) ([]*Host, error) {
	res, err := client.Get("/services/" + service, false, true)
	if err != nil {
		return nil, err
	}

	return buildHostsFromNodes(res.Node.Nodes), nil
}
