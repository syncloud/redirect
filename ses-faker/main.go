package main

import (
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
)

type Message struct {
	Source       string   `json:"source"`
	Destinations []string `json:"destinations"`
	Body         string   `json:"body"`
}

type Behaviour struct {
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Store struct {
	mutex     sync.Mutex
	messages  []Message
	behaviour Behaviour
}

func NewStore() *Store {
	return &Store{behaviour: Behaviour{Status: 200}}
}

func (s *Store) Add(message Message) Behaviour {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.behaviour.Status != 200 {
		return s.behaviour
	}
	s.messages = append(s.messages, message)
	return s.behaviour
}

func (s *Store) Messages() []Message {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return append([]Message{}, s.messages...)
}

func (s *Store) SetBehaviour(behaviour Behaviour) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if behaviour.Status == 0 {
		behaviour.Status = 200
	}
	s.behaviour = behaviour
}

func (s *Store) Reset() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.messages = nil
	s.behaviour = Behaviour{Status: 200}
}

type sendRawEmailResponse struct {
	XMLName xml.Name `xml:"SendRawEmailResponse"`
	Result  struct {
		MessageId string `xml:"MessageId"`
	} `xml:"SendRawEmailResult"`
}

type errorResponse struct {
	XMLName xml.Name `xml:"ErrorResponse"`
	Error   struct {
		Type    string `xml:"Type"`
		Code    string `xml:"Code"`
		Message string `xml:"Message"`
	} `xml:"Error"`
}

type API struct {
	store *Store
}

func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/faker/messages":
		a.messages(w, r)
	case "/faker/behaviour":
		a.behaviour(w, r)
	case "/faker/reset":
		a.store.Reset()
		w.WriteHeader(http.StatusOK)
	default:
		a.ses(w, r)
	}
}

func (a *API) messages(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(a.store.Messages())
}

func (a *API) behaviour(w http.ResponseWriter, r *http.Request) {
	var behaviour Behaviour
	if err := json.NewDecoder(r.Body).Decode(&behaviour); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	a.store.SetBehaviour(behaviour)
	w.WriteHeader(http.StatusOK)
}

func (a *API) ses(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if action := r.PostFormValue("Action"); action != "SendRawEmail" {
		a.fail(w, http.StatusBadRequest, "InvalidAction", fmt.Sprintf("unsupported action %q", action))
		return
	}

	message := Message{Source: r.PostFormValue("Source")}
	for key, values := range r.PostForm {
		if len(key) > len("Destinations.member.") && key[:len("Destinations.member.")] == "Destinations.member." {
			message.Destinations = append(message.Destinations, values[0])
		}
	}
	if raw := r.PostFormValue("RawMessage.Data"); raw != "" {
		if decoded, err := base64.StdEncoding.DecodeString(raw); err == nil {
			message.Body = string(decoded)
		} else {
			message.Body = raw
		}
	}

	behaviour := a.store.Add(message)
	if behaviour.Status != 200 {
		a.fail(w, behaviour.Status, behaviour.Code, behaviour.Message)
		return
	}

	response := sendRawEmailResponse{}
	response.Result.MessageId = fmt.Sprintf("faker-%d", len(a.store.Messages()))
	w.Header().Set("Content-Type", "text/xml")
	_ = xml.NewEncoder(w).Encode(response)
}

func (a *API) fail(w http.ResponseWriter, status int, code string, message string) {
	if code == "" {
		code = "InternalFailure"
	}
	response := errorResponse{}
	response.Error.Type = "Sender"
	response.Error.Code = code
	response.Error.Message = message
	w.Header().Set("Content-Type", "text/xml")
	w.WriteHeader(status)
	_ = xml.NewEncoder(w).Encode(response)
}

func env(name string, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func main() {
	address := env("SESSIM_ADDR", ":4579")
	log.Printf("ses faker listening on %s", address)
	if err := http.ListenAndServe(address, &API{store: NewStore()}); err != nil {
		log.Fatal(err)
	}
}
