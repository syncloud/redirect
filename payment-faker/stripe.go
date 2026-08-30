package main

import (
	"fmt"
	"net/http"
	"strings"
)

type Stripe struct {
	orders *Orders
	self   string
}

func NewStripe(orders *Orders, self string) *Stripe {
	return &Stripe{orders: orders, self: self}
}

func (s *Stripe) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/checkout/sessions", s.sessions)
	mux.HandleFunc("/v1/checkout/sessions/", s.session)
	mux.HandleFunc("/pay/", s.pay)
	return logging("stripe", mux)
}

func (s *Stripe) sessions(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "unreadable form", http.StatusBadRequest)
		return
	}
	amount := r.PostForm.Get("line_items[0][price_data][unit_amount]")
	currency := strings.ToUpper(r.PostForm.Get("line_items[0][price_data][currency]"))
	success := r.PostForm.Get("success_url")

	order := s.orders.Add("cs_faker_", currency, amount)
	write(w, map[string]any{
		"id":             order.Id,
		"object":         "checkout.session",
		"url":            fmt.Sprintf("%s/pay/%s?success=%s", s.self, order.Id, success),
		"payment_status": "unpaid",
		"amount_total":   0,
		"currency":       strings.ToLower(currency),
	})
}

func (s *Stripe) session(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/v1/checkout/sessions/")
	order := s.orders.Get(id)
	if order == nil {
		http.Error(w, "no such session", http.StatusNotFound)
		return
	}
	status := "unpaid"
	total := 0
	if order.Captured {
		status = "paid"
		total = atoi(order.Value)
	}
	write(w, map[string]any{
		"id":             order.Id,
		"object":         "checkout.session",
		"payment_status": status,
		"amount_total":   total,
		"currency":       strings.ToLower(order.Currency),
	})
}

func (s *Stripe) pay(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/pay/")
	order := s.orders.Get(id)
	if order == nil {
		http.Error(w, "no such session", http.StatusNotFound)
		return
	}
	success := r.URL.Query().Get("success")
	if r.Method == http.MethodPost {
		s.orders.Capture(id)
		http.Redirect(w, r, strings.ReplaceAll(success, "{CHECKOUT_SESSION_ID}", id), http.StatusFound)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `<!doctype html><title>Stripe faker</title>
<h1>Stripe faker</h1>
<p data-testid="faker-amount">%s %s</p>
<form method="post">
  <button type="submit" data-testid="faker-pay">Pay</button>
</form>`, order.Currency, order.Value)
}

func atoi(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int(c-'0')
	}
	return n
}
