package main

import (
	_ "embed"
	"fmt"
	"net/http"
)

//go:embed sdk.js
var sdk string

func (p *PayPal) script(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	fmt.Fprint(w, sdk)
}
