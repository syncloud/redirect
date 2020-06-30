package rest

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
)

type Api struct{}

func NewApi() *Api {
	return &Api{}
}
func (api *Api) Start(socket string) {
	http.HandleFunc("/status", Handle(api.Status))
	server := http.Server{}

	unixListener, err := net.Listen("unix", socket)
	if err != nil {
		panic(err)
	}
	log.Println("Started backend")
	_ = server.Serve(unixListener)

}

type Response struct {
	Success bool         `json:"success"`
	Message *string      `json:"message,omitempty"`
	Data    *interface{} `json:"data,omitempty"`
}

func fail(w http.ResponseWriter, err error) {
	appError := err.Error()
	response := Response{
		Success: false,
		Message: &appError,
	}
	responseJson, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		_, _ = fmt.Fprintf(w, string(responseJson))
	}
}

func success(w http.ResponseWriter, data interface{}) {
	response := Response{
		Success: true,
		Data:    &data,
	}
	responseJson, err := json.Marshal(response)
	if err != nil {
		fail(w, err)
	} else {
		_, _ = fmt.Fprintf(w, string(responseJson))
	}
}

func Handle(f func(w http.ResponseWriter, req *http.Request) (interface{}, error)) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		log.Printf("request: %s\n", req.URL.Path)
		w.Header().Add("Content-Type", "application/json")
		data, err := f(w, req)
		if err != nil {
			fail(w, err)
		} else {
			success(w, data)
		}
	}
}

func (backend *Api) Status(_ http.ResponseWriter, _ *http.Request) (interface{}, error) {
	return "OK", nil
}
