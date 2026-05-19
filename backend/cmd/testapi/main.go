package main

import (
	"github.com/syncloud/redirect/log"
	"github.com/syncloud/redirect/metrics"
	"github.com/syncloud/redirect/rest"
	"os"
)

func main() {
	domain := os.Getenv("DOMAIN")
	api := rest.NewApi(
		&TestDomains{},
		&TestUsers{},
		&TestMail{},
		&TestPortProbe{},
		&TestCertbot{},
		metrics.New(),
		domain,
		os.Getenv("SOCKET"),
		log.Default(),
	)
	api.Start()

}
