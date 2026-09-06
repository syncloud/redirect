package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type PayPal struct {
	orders *Orders
}

func NewPayPal(orders *Orders) *PayPal {
	return &PayPal{orders: orders}
}

func (p *PayPal) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/sdk/js", p.script)
	mux.HandleFunc("/v1/oauth2/token", p.token)
	mux.HandleFunc("/v2/checkout/orders", p.create)
	mux.HandleFunc("/v2/checkout/orders/", p.order)
	mux.HandleFunc("/v1/billing/subscriptions/", p.subscription)
	return logging(mux)
}

func (p *PayPal) token(w http.ResponseWriter, _ *http.Request) {
	write(w, map[string]any{
		"access_token": "faker-token",
		"token_type":   "Bearer",
		"expires_in":   32400,
		"app_id":       "faker",
		"nonce":        "faker",
	})
}

func (p *PayPal) create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		PurchaseUnits []struct {
			Amount struct {
				CurrencyCode string `json:"currency_code"`
				Value        string `json:"value"`
			} `json:"amount"`
		} `json:"purchase_units"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.PurchaseUnits) == 0 {
		http.Error(w, "no purchase unit", http.StatusBadRequest)
		return
	}
	unit := body.PurchaseUnits[0].Amount
	order := p.orders.Add(unit.CurrencyCode, unit.Value)
	write(w, map[string]any{
		"id":     order.Id,
		"status": "CREATED",
		"links": []map[string]string{
			{"rel": "approve", "href": "http://paypal-faker/approve/" + order.Id, "method": "GET"},
		},
	})
}

func (p *PayPal) order(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v2/checkout/orders/")
	id, action, _ := strings.Cut(path, "/")
	order := p.orders.Get(id)
	if order == nil {
		http.Error(w, "no such order", http.StatusNotFound)
		return
	}
	if action == "capture" {
		p.orders.Capture(id)
	}
	write(w, map[string]any{
		"id":     order.Id,
		"status": "COMPLETED",
		"purchase_units": []map[string]any{{
			"payments": map[string]any{
				"captures": []map[string]any{{
					"id":     "CAPTURE" + order.Id,
					"status": "COMPLETED",
					"amount": map[string]string{
						"currency_code": order.Currency,
						"value":         order.Value,
					},
				}},
			},
		}},
	})
}

func (p *PayPal) subscription(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/billing/subscriptions/")
	id, action, _ := strings.Cut(path, "/")
	if action == "cancel" {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	write(w, map[string]any{
		"id":          id,
		"status":      "ACTIVE",
		"plan_id":     "P-FAKERPLAN",
		"create_time": "2026-01-01T00:00:00Z",
	})
}

func write(w http.ResponseWriter, body any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Println("cannot write response:", err)
	}
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("paypal %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
