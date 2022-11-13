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
	api := rest.NewApi(
		TestStatsdClient{},
		&TestDomains{},
		&TestUsers{},
		&TestMail{},
		&TestPortProbe{},
		&TestCertbot{},
		domain,
	)
	api.StartApi(os.Getenv("SOCKET"))

}
