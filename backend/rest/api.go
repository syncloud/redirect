package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/kr/pretty"
	"github.com/syncloud/redirect/metrics"
	"github.com/syncloud/redirect/model"
	"github.com/syncloud/redirect/probe"
	"golang.org/x/net/netutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
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
	Probe(token string, port int, ip string) (*string, error)
}

type ApiCertbot interface {
	Present(token string, fqdn string, values []string) error
	CleanUp(token, fqdn string) error
}

type Api struct {
	statsdClient metrics.StatsdClient
	domains      ApiDomains
	users        ApiUsers
	mail         ApiMail
	probe        ApiPortProbe
	certbot      ApiCertbot
	domain       string
	count404     int64
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
 r.HandleFunc("/user/log_v2", Handle(a.UserLogV2)).Methods("POST")
	r.HandleFunc("/probe/port_v2", a.PortProbeV2).Methods("GET")
	r.HandleFunc("/probe/port_v3", Handle(a.PortProbeV3)).Methods("POST")
	r.NotFoundHandler = http.HandlerFunc(a.notFoundHandler)

	r.Use(headers)

	var listener net.Listener
	if strings.HasPrefix(socket, "tcp://") {
		address := strings.TrimPrefix(socket, "tcp://")
		tcpListener, err := net.Listen("tcp", address)
		if err != nil {
			panic(err)
		}
		listener = tcpListener
	} else {
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
		listener = unixListener
	}

	srv := &http.Server{
		Handler:      r,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
	l := netutil.LimitListener(listener, 1000)
	log.Printf("Started backend (%s)\n", socket)
	if err := srv.Serve(l); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
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
	request := model.DomainUpdateRequest{MapLocalAddress: false, Ipv4Enabled: true, Ipv6Enabled: true}
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		log.Println("unable to parse domain update request", err)
		return nil, errors.New("invalid request")
	}
	ip, err := a.requestIp(req)
	if err != nil {
		return nil, err
	}

	log.Printf("/domain/update, token: %# v, ipv4 enabled: %v, ip: %# v, ipv6 enabled: %v, ipv6: %# v\n", pretty.Formatter(request.Token), request.Ipv4Enabled, pretty.Formatter(request.Ip), request.Ipv6Enabled, pretty.Formatter(request.Ipv6))
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
	err = a.certbot.Present(request.Token, request.Fqdn, request.Values)
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
	err = a.certbot.CleanUp(request.Token, request.Fqdn)
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

func (a *Api) UserLogV2(_ http.ResponseWriter, req *http.Request) (interface{}, error) {
	a.statsdClient.Incr("rest.user_v2.log", 1)

request := model.SendLogsRequest{}
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		log.Println("unable to parse send logs v2 request", err)
		return nil, errors.New("invalid request")
	}

	
	user, err := a.users.GetUserByUpdateToken(request.Token)
	if err != nil {
		return nil, fmt.Errorf("user token: %s, error: %s", request.Token, err)
	}
	if user == nil {
		return nil, fmt.Errorf("wrong user token: %s", request.Token)
	}
	err = a.mail.SendLogs(user.Email, request.Data, request.IncludeSupport)
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

	requestIp, err := a.requestIp(req)
	if err != nil {
		fail(w, err)
		return
	}
	ip := *requestIp
	if ipParam, ok := req.URL.Query()["ip"]; ok {
		ip = ipParam[0]
	}

	result := &probe.Response{DeviceIp: ip}
	message, err := a.probe.Probe(tokenParam[0], port, ip)
	if err != nil {
		result.Message = err.Error()
		result.StatusCode = 500
	} else {
		result.Message = *message
		result.StatusCode = 200
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

func (a *Api) PortProbeV3(_ http.ResponseWriter, req *http.Request) (interface{}, error) {
	a.statsdClient.Incr("rest.probe.port_v3", 1)
	request := model.PortProbeRequest{}
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		log.Println("unable to parse probe v3 request", err)
		return nil, errors.New("invalid request")
	}

	ip := request.Ip
	if ip == nil {
		ip, err = a.requestIp(req)
		if err != nil {
			return nil, err
		}
	}

	message, err := a.probe.Probe(request.Token, request.Port, *ip)
	if err != nil {
		resultAddr := net.ParseIP(*ip)
		ipType := "4"
		resultIp := ""
		if resultAddr.To4() == nil {
			ipType = "6"
			resultIp = resultAddr.To16().String()
		} else {
			resultIp = resultAddr.To4().String()
		}

		return nil, model.NewServiceErrorWithCode(fmt.Sprintf("using device public IP: '%v' which is IPv%v. Details: %s", resultIp, ipType, err.Error()), 200)
	}

	return *message, nil
}

func (a *Api) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	a.count404++
	if a.count404%100 == 0 {
		log.Printf("404 counter: %v\n", a.count404)
	}
	http.NotFound(w, r)
}
