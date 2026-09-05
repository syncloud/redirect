package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	address := flag.String("address", ":4581", "listen address")
	flag.Parse()

	log.Printf("paypal faker on %s", *address)
	log.Fatal(http.ListenAndServe(*address, NewPayPal(NewOrders()).Handler()))
}
