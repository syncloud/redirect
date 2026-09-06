package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	address := flag.String("address", ":4582", "listen address")
	self := flag.String("url", "", "how a browser reaches this faker")
	flag.Parse()

	if *self == "" {
		log.Fatalln("--url is required, a browser has to reach the checkout page")
	}

	log.Printf("stripe faker on %s, reachable at %s", *address, *self)
	log.Fatal(http.ListenAndServe(*address, NewStripe(NewSessions(), *self).Handler()))
}
