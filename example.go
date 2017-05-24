package main

import (
	"log"

	"github.com/Scalingo/etcd-discovery/service"
	"github.com/Scalingo/gopassword"
)

func main() {

	stopper := make(chan struct{})
	changes := service.Register(
		"mon-service",
		&service.Host{
			Name: "172.17.0.1",
			Ports: service.Ports{
				"http":  "8080",
				"https": "80443",
			},
		},
		&service.Infos{
			Critical:       true,
			User:           gopassword.Generate(10),
			Password:       gopassword.Generate(10),
			PublicHostname: "scalingo.dev",
		}, stopper)
	for {
		c := <-changes
		log.Println("---- CHANGEMENT DE MOT DE PASSE -----")
		log.Println("User : \t" + c.User)
		log.Println("Password: \t" + c.Password)
	}
}
