package main

import (
	"github.com/syncloud/redirect/log"
	"github.com/syncloud/redirect/metrics"
	"github.com/syncloud/redirect/rest"
	"net/http"
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
		&TestComplaints{},
		domain,
		os.Getenv("SOCKET"),
		log.Default(),
	)
	api.Start()

}

type TestComplaints struct{}

func (t *TestComplaints) Handle(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
