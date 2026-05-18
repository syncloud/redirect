package main

import (
	"github.com/syncloud/redirect/log"
	"github.com/syncloud/redirect/rest"
	"os"
)

func main() {
	domain := os.Getenv("DOMAIN")
	api := rest.NewApi(
		TestStatsdClient{},
		&TestDomains{},
		&TestUsers{},
		&TestMail{},
		&TestPortProbe{},
		&TestCertbot{},
		domain,
		os.Getenv("SOCKET"),
		log.Default(),
	)
	api.Start()

}
