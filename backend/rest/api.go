package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/smira/go-statsd"
	"github.com/syncloud/redirect/model"
	"github.com/syncloud/redirect/service"
	"log"
	"net"
	"net/http"
	"os"
)

type Api struct {
	statsdClient *statsd.Client
	service      *service.Service
}

func NewApi(statsdClient *statsd.Client, service *service.Service) *Api {
	return &Api{statsdClient, service}
}
func (a *Api) Start(socket string) {
	http.HandleFunc("/status", Handle(a.Status))
	http.HandleFunc("/domain/update", Handle(a.DomainUpdate))
	http.HandleFunc("/domain/get", Handle(a.DomainGet))
	server := http.Server{}
	if _, err := os.Stat(socket); err == nil {
		err := os.Remove(socket)
		if err != nil {
			panic(err)
		}
	}
	unixListener, err := net.Listen("unix", socket)
	if err != nil {
		panic(err)
	}
	if err := os.Chmod(socket, 0777); err != nil {
		log.Fatal(err)
	}
	log.Println("Started backend")
	_ = server.Serve(unixListener)

}

type Response struct {
	Success            bool                       `json:"success"`
	Message            string                     `json:"message,omitempty"`
	Data               *interface{}               `json:"data,omitempty"`
	ParametersMessages *[]model.ParameterMessages `json:"parameters_messages,omitempty"`
}

func fail(w http.ResponseWriter, serviceError *model.Error) {
	statusCode := serviceError.StatusCode
	response := Response{Success: false, Message: "Unknown Error"}
	if serviceError.Error != nil {
		response.Message = serviceError.Error.Error()
	} else if serviceError.ParameterErrors != nil {
		response.ParametersMessages = serviceError.ParameterErrors
		response.Message = "There's an error in parameters"
	}

	responseJson, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), statusCode)
	} else {
		http.Error(w, string(responseJson), statusCode)
	}
}

func success(w http.ResponseWriter, message string, data interface{}) {
	response := Response{
		Success: true,
		Message: message,
		Data:    &data,
	}
	responseJson, err := json.Marshal(response)
	if err != nil {
		fail(w, model.UnknownError(err))
	} else {
		_, _ = fmt.Fprintf(w, string(responseJson))
	}
}

func Handle(f func(w http.ResponseWriter, req *http.Request) (string, interface{}, *model.Error)) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		message, data, err := f(w, req)
		if err != nil {
			fail(w, err)
		} else {
			success(w, message, data)
		}
	}
}

func (a *Api) Status(_ http.ResponseWriter, _ *http.Request) (string, interface{}, *model.Error) {
	return "Up and running", "OK", nil
}

func (a *Api) DomainUpdate(w http.ResponseWriter, req *http.Request) (string, interface{}, *model.Error) {
	a.statsdClient.Incr("rest.domain.update", 1)
	if err := req.ParseForm(); err != nil {
		return "", nil, model.UnknownError(errors.New("cannot parse post form"))
	}

	request := model.DomainUpdateRequest{MapLocalAddress: false}
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		return "", nil, model.UnknownError(err)
	}
	domain, paramError := a.service.Update(request, a.requestIp(req))
	if paramError != nil {
		return "", nil, paramError
	}
	return "Domain was updated", domain, nil
}

func (a *Api) DomainGet(w http.ResponseWriter, req *http.Request) (string, interface{}, *model.Error) {
	a.statsdClient.Incr("rest.domain.get", 1)
	keys, ok := req.URL.Query()["token"]
	if !ok {
		return "", nil, model.UnknownError(errors.New("no token"))
	}

	domain, paramError := a.service.GetDomain(keys[0])
	if paramError != nil {
		return "", nil, paramError
	}
	return "Domain retrieved", domain, nil
}

func (a *Api) requestIp(req *http.Request) *string {
	requestIp := req.Header.Get("X-FORWARDED-FOR")
	if requestIp != "" {
		return &requestIp
	}

	requestAddr := req.RemoteAddr
	ip, _, err := net.SplitHostPort(requestAddr)
	if err != nil {
		log.Println("cannot parse request addr", err)
		return nil
	}
	return &ip
}
