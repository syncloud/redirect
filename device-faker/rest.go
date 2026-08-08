package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type TunnelRequest struct {
	Domain string `json:"domain"`
	Token  string `json:"token"`
}

type Rest struct {
	address string
	mailbox *Mailbox
	tunnels *Tunnels
}

func NewRest(address string, mailbox *Mailbox, tunnels *Tunnels) *Rest {
	return &Rest{address: address, mailbox: mailbox, tunnels: tunnels}
}

func (r *Rest) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/faker/messages", r.messages)
	mux.HandleFunc("/faker/behaviour", r.behaviour)
	mux.HandleFunc("/faker/tunnel", r.tunnel)
	mux.HandleFunc("/faker/reset", r.reset)
	go func() {
		if err := http.ListenAndServe(r.address, mux); err != nil {
			log.Printf("device api stopped: %v", err)
		}
	}()
	return nil
}

func (r *Rest) messages(w http.ResponseWriter, _ *http.Request) {
	respond(w, r.mailbox.Messages())
}

func (r *Rest) behaviour(w http.ResponseWriter, request *http.Request) {
	var behaviour Behaviour
	if err := json.NewDecoder(request.Body).Decode(&behaviour); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	r.mailbox.SetBehaviour(behaviour)
	respond(w, r.mailbox.Behaviour())
}

func (r *Rest) tunnel(w http.ResponseWriter, request *http.Request) {
	var tunnel TunnelRequest
	if err := json.NewDecoder(request.Body).Decode(&tunnel); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := r.tunnels.Start(tunnel.Domain, tunnel.Token); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	respond(w, tunnel)
}

func (r *Rest) reset(w http.ResponseWriter, _ *http.Request) {
	r.tunnels.StopAll()
	r.mailbox.Reset()
	respond(w, r.mailbox.Messages())
}

func respond(w http.ResponseWriter, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("device api failed to respond: %v", err)
	}
}
