package service

type Infos struct {
	Name           string `json:"name"`
	Critical       bool   `json:"critical"`                  // Is the service critical to the infrastructure health?
	PublicHostname string `json:"public_hostname,omitempty"` // The service public hostname
	User           string `json:"user,omitempty"`            // The service username
	Password       string `json:"password,omitempty"`        // The service password
}

type Credentials struct {
	User     string
	Password string
}
