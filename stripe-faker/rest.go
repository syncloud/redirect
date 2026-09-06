package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type Stripe struct {
	sessions *Sessions
	self     string
}

func NewStripe(sessions *Sessions, self string) *Stripe {
	return &Stripe{sessions: sessions, self: self}
}

func (s *Stripe) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/checkout/sessions", s.create)
	mux.HandleFunc("/v1/checkout/sessions/", s.retrieve)
	mux.HandleFunc("/pay/", s.pay)
	return logging(mux)
}

func (s *Stripe) create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "unreadable form", http.StatusBadRequest)
		return
	}
	amount, err := strconv.Atoi(r.PostForm.Get("line_items[0][price_data][unit_amount]"))
	if err != nil {
		http.Error(w, "unit_amount is not a number of minor units", http.StatusBadRequest)
		return
	}
	currency := strings.ToLower(r.PostForm.Get("line_items[0][price_data][currency]"))
	success := r.PostForm.Get("success_url")

	session := s.sessions.Add(currency, amount)
	write(w, map[string]any{
		"id":             session.Id,
		"object":         "checkout.session",
		"url":            fmt.Sprintf("%s/pay/%s?success=%s", s.self, session.Id, success),
		"payment_status": "unpaid",
		"amount_total":   0,
		"currency":       session.Currency,
	})
}

func (s *Stripe) retrieve(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/v1/checkout/sessions/")
	session := s.sessions.Get(id)
	if session == nil {
		http.Error(w, "no such session", http.StatusNotFound)
		return
	}
	status := "unpaid"
	total := 0
	if session.Paid {
		status = "paid"
		total = session.Amount
	}
	write(w, map[string]any{
		"id":             session.Id,
		"object":         "checkout.session",
		"payment_status": status,
		"amount_total":   total,
		"currency":       session.Currency,
	})
}

func (s *Stripe) pay(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/pay/")
	session := s.sessions.Get(id)
	if session == nil {
		http.Error(w, "no such session", http.StatusNotFound)
		return
	}
	success := r.URL.Query().Get("success")
	if r.Method == http.MethodPost {
		s.sessions.Pay(id)
		http.Redirect(w, r, strings.ReplaceAll(success, "{CHECKOUT_SESSION_ID}", id), http.StatusFound)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `<!doctype html><title>Stripe faker</title>
<h1>Stripe faker</h1>
<p data-testid="faker-amount">%s %s</p>
<form method="post">
  <button type="submit" data-testid="faker-pay">Pay</button>
</form>`, strings.ToUpper(session.Currency), money(session.Amount))
}

func money(minor int) string {
	return fmt.Sprintf("%d.%02d", minor/100, minor%100)
}

func write(w http.ResponseWriter, body any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Println("cannot write response:", err)
	}
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("stripe %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
