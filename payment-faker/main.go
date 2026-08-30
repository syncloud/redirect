package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	paypalAddress := flag.String("paypal", ":4581", "paypal faker address")
	stripeAddress := flag.String("stripe", ":4582", "stripe faker address")
	stripeSelf := flag.String("stripe-url", "", "how a browser reaches the stripe faker")
	flag.Parse()

	if *stripeSelf == "" {
		log.Fatalln("--stripe-url is required, a browser has to be able to reach the checkout page")
	}

	orders := NewOrders()

	go func() {
		log.Printf("paypal faker on %s", *paypalAddress)
		log.Fatal(http.ListenAndServe(*paypalAddress, NewPayPal(orders).Handler()))
	}()

	log.Printf("stripe faker on %s, reachable at %s", *stripeAddress, *stripeSelf)
	log.Fatal(http.ListenAndServe(*stripeAddress, NewStripe(orders, *stripeSelf).Handler()))
}
