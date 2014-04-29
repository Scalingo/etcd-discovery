package service

func Get(service string) ([]*Host, error) {
	res, err := Client().Get("/services/"+service, false, true)
	if err != nil {
		if IsKeyNotFoundError(err) {
			return []*Host{}, nil
		}
		return nil, err
	}

	return buildHostsFromNodes(res.Node.Nodes), nil
}
