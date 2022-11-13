package main

import (
	"fmt"
	"github.com/syncloud/redirect/rest"
	"log"
	"os"
)

type logWriter struct {
}

func (writer logWriter) Write(bytes []byte) (int, error) {
	return fmt.Print(string(bytes))
}

func main() {
	log.SetFlags(0)
	log.SetOutput(new(logWriter))

	domain := os.Getenv("DOMAIN")
	statsdClient := TestStatsdClient{}
	mail := &TestMail{}
	users := &TestUsers{}
	domains := &TestDomains{}
	prober := &TestPortProbe{}
	certbot := &TestCertbot{}
	api := rest.NewApi(statsdClient, domains, users, mail, prober, certbot, domain)
	api.StartApi(os.Getenv("SOCKET"))

}
