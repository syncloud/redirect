package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/syncloud/redirect/metrics"
	"github.com/syncloud/redirect/model"
	"github.com/syncloud/redirect/service"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type ApiDomains interface {
	DomainAcquire(request model.DomainAcquireRequest, domainField string) (*model.Domain, error)
	Availability(request model.DomainAvailabilityRequest) (*model.Domain, error)
	Update(request model.DomainUpdateRequest, requestIp *string) (*model.Domain, error)
	GetDomain(token string) (*model.Domain, error)
	GetDomains(user *model.User) ([]*model.Domain, error)
}

type ApiUsers interface {
	CreateNewUser(request model.UserCreateRequest) (*model.User, error)
	Authenticate(email *string, password *string) (*model.User, error)
	GetUserByUpdateToken(updateToken string) (*model.User, error)
}

type ApiMail interface {
	SendLogs(to string, data string, includeSupport bool) error
}

type ApiPortProbe interface {
	Probe(token string, port int, protocol string, ip string) (*service.ProbeResponse, error)
}

type ApiCertbot interface {
	Present(token, fqdn, value string) error
	CleanUp(token, fqdn, value string) error
}

type Api struct {
	statsdClient metrics.StatsdClient
	domains      ApiDomains
	users        ApiUsers
	mail         ApiMail
	probe        ApiPortProbe
	certbot      ApiCertbot
	domain       string
}

func NewApi(statsdClient metrics.StatsdClient, service ApiDomains, users ApiUsers, mail ApiMail, probe ApiPortProbe, certbot ApiCertbot, domain string) *Api {
	return &Api{statsdClient: statsdClient, domains: service, users: users, mail: mail, probe: probe, certbot: certbot, domain: domain}
}

func (a *Api) StartApi(socket string) {
	r := mux.NewRouter()
	r.HandleFunc("/status", Handle(a.Status)).Methods("GET")
	r.HandleFunc("/certbot/present", Handle(a.CertbotPresent)).Methods("POST")
	r.HandleFunc("/certbot/cleanup", Handle(a.CertbotCleanUp)).Methods("POST")
	r.HandleFunc("/domain/update", Handle(a.DomainUpdate)).Methods("POST")
	r.HandleFunc("/domain/get", Handle(a.DomainGet)).Methods("GET")
	r.HandleFunc("/domain/acquire", a.DomainAcquireV1).Methods("POST")
	r.HandleFunc("/domain/acquire_v2", Handle(a.DomainAcquireV2)).Methods("POST")
	r.HandleFunc("/domain/availability", Handle(a.DomainAvailability)).Methods("POST")
	r.HandleFunc("/user/create", Handle(a.UserCreate)).Methods("POST")
	r.HandleFunc("/user/create_v2", Handle(a.UserCreateV2)).Methods("POST")
	r.HandleFunc("/user/get", Handle(a.UserGet)).Methods("GET") //deprecated
	r.HandleFunc("/user", Handle(a.User)).Methods("POST")
	r.HandleFunc("/user/log", Handle(a.UserLog)).Methods("POST")
	r.HandleFunc("/probe/port_v2", a.PortProbeV2).Methods("GET")
	r.NotFoundHandler = http.HandlerFunc(notFoundHandler)

	r.Use(headers)

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
	log.Printf("Started backend (%s)\n", socket)
	_ = http.Serve(unixListener, r)

}

func (a *Api) Status(_ http.ResponseWriter, req *http.Request) (interface{}, error) {
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
	request := model.DomainAcquireRequest{}
	if userDomain := req.PostForm.Get("user_domain"); userDomain != "" {
		request.DeprecatedUserDomain = &userDomain
		request.ForwardCompatibleDomain(a.domain)
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
	domain, err := a.domains.DomainAcquire(request, "user_domain")
	if err != nil {
		fail(w, err)
		return
	}
	response := DomainAcquireResponse{
		Success:              true,
		UpdateToken:          *domain.UpdateToken,
		DeprecatedUserDomain: domain.DeprecatedUserDomain,
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

func (a *Api) UserCreate(_ http.ResponseWriter, req *http.Request) (interface{}, error) {
	a.statsdClient.Incr("rest.user.create", 1)
	err := req.ParseForm()
	if err != nil {
		log.Println("unable to parse form", err)
		return nil, errors.New("invalid request")
	}
	request := model.UserCreateRequest{}
	if email := req.PostForm.Get("email"); email != "" {
		request.Email = &email
	}
	if password := req.PostForm.Get("password"); password != "" {
		request.Password = &password
	}

	return a.users.CreateNewUser(request)
}

func (a *Api) UserCreateV2(_ http.ResponseWriter, req *http.Request) (interface{}, error) {
	a.statsdClient.Incr("rest.user.create", 1)
	request := model.UserCreateRequest{}
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		log.Println("unable to parse user create request", err)
		return nil, errors.New("invalid request")
	}

	return a.users.CreateNewUser(request)
}

func (a *Api) DomainAcquireV2(_ http.ResponseWriter, req *http.Request) (interface{}, error) {
	a.statsdClient.Incr("rest.domain.acquire_v2", 1)
	request := model.DomainAcquireRequest{}
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		log.Println("unable to parse domain acquire request", err)
		return nil, errors.New("invalid request")
	}
	domain, err := a.domains.DomainAcquire(request, "domain")
	if err != nil {
		return nil, err
	}
	return domain, nil
}

func (a *Api) DomainAvailability(_ http.ResponseWriter, req *http.Request) (interface{}, error) {
	a.statsdClient.Incr("rest.domain.availability", 1)
	request := model.DomainAvailabilityRequest{}
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		log.Println("unable to parse request", err)
		return nil, errors.New("invalid request")
	}
	domain, err := a.domains.Availability(request)
	if err != nil {
		return nil, err
	}
	return domain, nil
}

func (a *Api) DomainUpdate(_ http.ResponseWriter, req *http.Request) (interface{}, error) {
	a.statsdClient.Incr("rest.domain.update", 1)
	request := model.DomainUpdateRequest{MapLocalAddress: false}
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		log.Println("unable to parse domain update request", err)
		return nil, errors.New("invalid request")
	}
	ip, err := a.requestIp(req)
	if err != nil {
		return nil, err
	}

	domain, err := a.domains.Update(request, ip)
	if err != nil {
		return nil, err
	}
	return domain, nil
}

func (a *Api) CertbotPresent(_ http.ResponseWriter, req *http.Request) (interface{}, error) {
	a.statsdClient.Incr("rest.certbot.present", 1)
	request := model.CertbotPresentRequest{}
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		log.Println("unable to parse request", err)
		return nil, errors.New("invalid request")
	}
	err = a.certbot.Present(request.Token, request.Fqdn, request.Value)
	return nil, err
}

func (a *Api) CertbotCleanUp(_ http.ResponseWriter, req *http.Request) (interface{}, error) {
	a.statsdClient.Incr("rest.certbot.cleanup", 1)
	request := model.CertbotCleanUpRequest{}
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		log.Println("unable to parse request", err)
		return nil, errors.New("invalid request")
	}
	err = a.certbot.CleanUp(request.Token, request.Fqdn, request.Value)
	return nil, err
}

func (a *Api) DomainGet(_ http.ResponseWriter, req *http.Request) (interface{}, error) {
	a.statsdClient.Incr("rest.domain.get", 1)
	keys, ok := req.URL.Query()["token"]
	if !ok {
		return nil, errors.New("no token")
	}

	domain, err := a.domains.GetDomain(keys[0])
	if err != nil {
		return nil, err
	}
	return domain, nil
}

func (a *Api) requestIp(req *http.Request) (*string, error) {
	requestIp := req.Header.Get("X-FORWARDED-FOR")
	if requestIp != "" {
		return &requestIp, nil
	}

	requestAddr := req.RemoteAddr
	ip, _, err := net.SplitHostPort(requestAddr)
	if err != nil {
		log.Println("cannot parse request addr", err)
		return nil, err
	}
	return &ip, nil
}

func (a *Api) UserGet(_ http.ResponseWriter, req *http.Request) (interface{}, error) {

	a.statsdClient.Incr("rest.user.get", 1)
	emails, ok := req.URL.Query()["email"]
	if !ok {
		return nil, errors.New("no email")
	}
	passwords, ok := req.URL.Query()["password"]
	if !ok {
		return nil, errors.New("no password")
	}
	request := model.UserAuthenticateRequest{Email: &emails[0], Password: &passwords[0]}
	user, err := a.users.Authenticate(request.Email, request.Password)
	if err != nil {
		log.Println("unable to get a user", err)
		return nil, errors.New("invalid request")
	}
	domains, err := a.domains.GetDomains(user)
	if err != nil {
		log.Println("unable to get domains for a user", err)
		return nil, errors.New("invalid request")
	}
	if domains == nil {
		domains = make([]*model.Domain, 0)
	}
	return &model.UserResponse{
		Email:        user.Email,
		Active:       user.Active,
		UpdateToken:  user.UpdateToken,
		Unsubscribed: !user.NotificationEnabled,
		Timestamp:    user.Timestamp,
		Domains:      domains}, nil
}

func (a *Api) User(_ http.ResponseWriter, req *http.Request) (interface{}, error) {

	a.statsdClient.Incr("rest.user", 1)
	request := &model.UserAuthenticateRequest{}
	err := json.NewDecoder(req.Body).Decode(request)
	if err != nil {
		log.Println("unable to parse user request", err)
		return nil, errors.New("unable to parse user request")
	}
	user, err := a.users.Authenticate(request.Email, request.Password)
	if err != nil {
		log.Println("unable to get a user", err)
		return nil, errors.New("unable to get a user")
	}
	domains, err := a.domains.GetDomains(user)
	if err != nil {
		log.Println("unable to get domains for a user", err)
		return nil, errors.New("unable to get domains for a user")
	}
	if domains == nil {
		domains = make([]*model.Domain, 0)
	}
	return &model.UserResponse{
		Email:        user.Email,
		Active:       user.Active,
		UpdateToken:  user.UpdateToken,
		Unsubscribed: !user.NotificationEnabled,
		Timestamp:    user.Timestamp,
		Domains:      domains}, nil
}

func (a *Api) UserLog(_ http.ResponseWriter, req *http.Request) (interface{}, error) {
	a.statsdClient.Incr("rest.user.log", 1)

	err := req.ParseForm()
	if err != nil {
		return nil, errors.New("invalid request")
	}

	token := req.PostForm.Get("token")
	data := req.PostForm.Get("data")
	includeSupport := strings.ToLower(req.PostForm.Get("include_support")) == "true"
	user, err := a.users.GetUserByUpdateToken(token)
	if err != nil {
		return nil, fmt.Errorf("user token: %s, error: %s", token, err)
	}
	if user == nil {
		return nil, fmt.Errorf("wrong user token: %s", token)
	}
	err = a.mail.SendLogs(user.Email, data, includeSupport)
	return "Error report sent successfully", err
}

func (a *Api) PortProbeV2(w http.ResponseWriter, req *http.Request) {
	a.statsdClient.Incr("rest.probe.port_v2", 1)
	tokenParam, ok := req.URL.Query()["token"]
	if !ok {
		fail(w, model.SingleParameterError("token", "Missing"))
		return
	}
	portParam, ok := req.URL.Query()["port"]
	if !ok {
		fail(w, model.SingleParameterError("port", "Missing"))
		return
	}
	port, err := strconv.Atoi(portParam[0])
	if err != nil {
		fail(w, err)
		return
	}

	protocol := "https"
	if protocolParam, ok := req.URL.Query()["protocol"]; ok {
		protocol = protocolParam[0]
	}
	requestIp, err := a.requestIp(req)
	if err != nil {
		fail(w, err)
		return
	}
	ip := *requestIp
	if ipParam, ok := req.URL.Query()["ip"]; ok {
		ip = ipParam[0]
	}

	result, err := a.probe.Probe(tokenParam[0], port, protocol, ip)
	if err != nil {
		fail(w, err)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	responseJson, err := json.Marshal(result)
	if err != nil {
		fail(w, err)
		return
	}
	w.WriteHeader(result.StatusCode)
	_, _ = fmt.Fprintf(w, string(responseJson))
}
