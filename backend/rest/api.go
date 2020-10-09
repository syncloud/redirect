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
	service      *service.Domains
}

func NewApi(statsdClient *statsd.Client, service *service.Domains) *Api {
	return &Api{statsdClient, service}
}
func (a *Api) Start(socket string) {
	http.HandleFunc("/status", Handle("GET", a.Status))
	http.HandleFunc("/domain/update", Handle("POST", a.DomainUpdate))
	http.HandleFunc("/domain/get", Handle("GET", a.DomainGet))
	http.HandleFunc("/domain/acquire", a.DomainAcquireV1)
	http.HandleFunc("/domain/acquire_v2", Handle("POST", a.DomainAcquireV2))
	http.HandleFunc("/domain/acquire_custom", Handle("POST", a.CustomDomainAcquire))
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

type DomainAcquireResponse struct {
	Success     bool   `json:"success"`
	UserDomain  string `json:"user_domain,omitempty"`
	UpdateToken string `json:"update_token,omitempty"`
}

type Response struct {
	Success            bool                       `json:"success"`
	Message            string                     `json:"message,omitempty"`
	Data               *interface{}               `json:"data,omitempty"`
	ParametersMessages *[]model.ParameterMessages `json:"parameters_messages,omitempty"`
}

func fail(w http.ResponseWriter, err error) {
	response, statusCode := ErrorToResponse(err)
	responseJson, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), statusCode)
	} else {
		http.Error(w, string(responseJson), statusCode)
	}
}

func ErrorToResponse(err error) (Response, int) {
	response := Response{Success: false, Message: "Unknown Error"}
	statusCode := 500
	switch v := err.(type) {
	case *model.ParameterError:
		response.ParametersMessages = v.ParameterErrors
		statusCode = 400
	case *model.ServiceError:
		statusCode = 400
	}
	response.Message = err.Error()
	return response, statusCode
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

func Handle(method string, f func(req *http.Request) (interface{}, error)) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		if req.Method != method {
			fail(w, errors.New(fmt.Sprintf("wrong method %s, should be POST", req.Method)))
		}
		data, err := f(req)
		if err != nil {
			fail(w, err)
		} else {
			success(w, data)
		}
	}
}

func (a *Api) Status(req *http.Request) (interface{}, error) {
	if req.Method != "GET" {
		return nil, errors.New(fmt.Sprintf("wrong method %s, should be GET", req.Method))
	}
	return "OK", nil
}

func (a *Api) DomainAcquireV1(w http.ResponseWriter, req *http.Request) {
	a.statsdClient.Incr("rest.domain.acquire", 1)
	err := req.ParseForm()
	if err != nil {
		log.Println("unable to parse form", err)
		fail(w, errors.New("invalid request"))
		return
	}
	if req.Method != "POST" {
		fail(w, errors.New(fmt.Sprintf("wrong method %s, should be POST", req.Method)))
		return
	}
	request := model.DomainAcquireRequest{}
	if userDomain := req.PostForm.Get("user_domain"); userDomain != "" {
		request.UserDomain = &userDomain
	}
	if password := req.PostForm.Get("password"); password != "" {
		request.Password = &password
	}
	if email := req.PostForm.Get("email"); email != "" {
		request.Email = &email
	}
	if deviceMacAddress := req.PostForm.Get("device_mac_address"); deviceMacAddress != "" {
		request.DeviceMacAddress = &deviceMacAddress
	}
	if deviceName := req.PostForm.Get("device_name"); deviceName != "" {
		request.DeviceName = &deviceName
	}
	if deviceTitle := req.PostForm.Get("device_title"); deviceTitle != "" {
		request.DeviceTitle = &deviceTitle
	}
	domain, err := a.service.DomainAcquire(request)
	if err != nil {
		fail(w, err)
		return
	}
	response := DomainAcquireResponse{
		Success:     true,
		UpdateToken: *domain.UpdateToken,
		UserDomain:  domain.UserDomain,
	}
	w.Header().Add("Content-Type", "application/json")
	responseJson, err := json.Marshal(response)
	if err != nil {
		fail(w, err)
		return
	} else {
		_, _ = fmt.Fprintf(w, string(responseJson))
	}
}

func (a *Api) DomainAcquireV2(req *http.Request) (interface{}, error) {
	a.statsdClient.Incr("rest.domain.acquire", 1)
	request := model.DomainAcquireRequest{}
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		log.Println("unable to parse domain acquire request", err)
		return nil, errors.New("invalid request")
	}
	domain, err := a.service.DomainAcquire(request)
	if err != nil {
		return nil, err
	}
	return domain, nil
}

func (a *Api) CustomDomainAcquire(req *http.Request) (interface{}, error) {
	a.statsdClient.Incr("rest.custom.domain.acquire", 1)
	request := model.CustomDomainAcquireRequest{}
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		log.Println("unable to parse custom domain acquire request", err)
		return nil, errors.New("invalid request")
	}
	domain, err := a.service.CustomDomainAcquire(request)
	if err != nil {
		return nil, err
	}
	return domain, nil
}

func (a *Api) DomainUpdate(req *http.Request) (interface{}, error) {
	a.statsdClient.Incr("rest.domain.update", 1)
	request := model.DomainUpdateRequest{MapLocalAddress: false}
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		log.Println("unable to parse domain update request", err)
		return nil, errors.New("invalid request")
	}
	domain, err := a.service.Update(request, a.requestIp(req))
	if err != nil {
		return nil, err
	}
	return domain, nil
}

func (a *Api) DomainGet(req *http.Request) (interface{}, error) {
	a.statsdClient.Incr("rest.domain.get", 1)
	keys, ok := req.URL.Query()["token"]
	if !ok {
		return nil, errors.New("no token")
	}

	domain, err := a.service.GetDomain(keys[0])
	if err != nil {
		return nil, err
	}
	return domain, nil
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
