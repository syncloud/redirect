package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/smira/go-statsd"
	"github.com/syncloud/redirect/model"
	"github.com/syncloud/redirect/service"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
)

type Api struct {
	statsdClient *statsd.Client
	domains      *service.Domains
	users        *service.Users
	actions      *service.Actions
	mail         *service.Mail
	probe        *service.PortProbe
}

func NewApi(statsdClient *statsd.Client, service *service.Domains, users *service.Users, actions *service.Actions, mail *service.Mail, probe *service.PortProbe) *Api {
	return &Api{statsdClient: statsdClient, domains: service, users: users, actions: actions, mail: mail, probe: probe}
}

func (a *Api) Start(socket string) {
	r := mux.NewRouter()
	r.HandleFunc("/status", Handle(a.Status)).Methods("GET")
	r.HandleFunc("/domain/update", Handle(a.DomainUpdate)).Methods("POST")
	r.HandleFunc("/domain/get", Handle(a.DomainGet)).Methods("GET")
	r.HandleFunc("/domain/acquire", a.DomainAcquireV1).Methods("POST")
	r.HandleFunc("/domain/acquire_v2", Handle(a.DomainAcquireV2)).Methods("POST")
	r.HandleFunc("/domain/availability", Handle(a.DomainAvailability)).Methods("POST")
	r.HandleFunc("/user/create", Handle(a.UserCreate)).Methods("POST")
	r.HandleFunc("/user/create_v2", Handle(a.UserCreateV2)).Methods("POST")
	r.HandleFunc("/user/get", Handle(a.UserGet)).Methods("GET")
	r.HandleFunc("/user/log", Handle(a.UserLog)).Methods("POST")
	r.HandleFunc("/probe/port_v2", a.PortProbeV2).Methods("GET")

	r.HandleFunc("/web/notification/subscribe", Handle(a.WebNotificationSubscribe)).Methods("POST")
	r.HandleFunc("/web/notification/unsubscribe", Handle(a.WebNotificationUnsubscribe)).Methods("POST")
	r.HandleFunc("/web/user", Handle(a.WebUserDelete)).Methods("DELETE")
	r.HandleFunc("/web/user", Handle(a.WebUser)).Methods("GET")
	r.HandleFunc("/web/domains", Handle(a.WebDomains)).Methods("GET")
	r.HandleFunc("/web/premium/request", Handle(a.WebPremiumRequest)).Methods("POST")
	r.HandleFunc("/web/user/reset_password", Handle(a.WebUserPasswordReset)).Methods("POST")
	r.HandleFunc("/web/user/activate", Handle(a.WebUserActivate)).Methods("POST")
	r.HandleFunc("/web/user/create", Handle(a.UserCreateV2)).Methods("POST")

	r.Use(middleware)

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
	_ = http.Serve(unixListener, r)

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

func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func Handle(f func(req *http.Request) (interface{}, error)) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
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
	domain, err := a.domains.DomainAcquire(request)
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

func (a *Api) UserCreate(req *http.Request) (interface{}, error) {
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

func (a *Api) UserCreateV2(req *http.Request) (interface{}, error) {
	a.statsdClient.Incr("rest.user.create", 1)
	request := model.UserCreateRequest{}
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		log.Println("unable to parse user create request", err)
		return nil, errors.New("invalid request")
	}

	return a.users.CreateNewUser(request)
}

func (a *Api) DomainAcquireV2(req *http.Request) (interface{}, error) {
	a.statsdClient.Incr("rest.domain.acquire_v2", 1)
	request := model.DomainAcquireRequest{}
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		log.Println("unable to parse domain acquire request", err)
		return nil, errors.New("invalid request")
	}
	domain, err := a.domains.DomainAcquire(request)
	if err != nil {
		return nil, err
	}
	return domain, nil
}

func (a *Api) DomainAvailability(req *http.Request) (interface{}, error) {
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

func (a *Api) WebNotificationSubscribe(req *http.Request) (interface{}, error) {
	a.statsdClient.Incr("www.user.subscribe", 1)
	user, err := a.getAuthenticatedUser(req)
	if err != nil {
		return nil, err
	}
	user.Unsubscribed = false
	return "OK", a.users.Save(user)
}

func (a *Api) WebNotificationUnsubscribe(req *http.Request) (interface{}, error) {
	a.statsdClient.Incr("www.user.unsubscribe", 1)
	user, err := a.getAuthenticatedUser(req)
	if err != nil {
		return nil, err
	}
	user.Unsubscribed = true
	return "OK", a.users.Save(user)
}

func (a *Api) WebUserDelete(req *http.Request) (interface{}, error) {
	a.statsdClient.Incr("www.user.delete", 1)
	user, err := a.getAuthenticatedUser(req)
	if err != nil {
		return nil, err
	}

	err = a.domains.DeleteAllDomains(user.Id)
	if err != nil {
		log.Println("unable to delete domains for a user", err)
		return nil, errors.New("invalid request")
	}
	err = a.actions.DeleteActions(user.Id)
	if err != nil {
		log.Println("unable to delete actions for a user", err)
		return nil, errors.New("invalid request")
	}

	err = a.users.Delete(user.Id)
	if err != nil {
		log.Println("unable to delete a user", err)
		return nil, errors.New("invalid request")
	}

	return "OK", nil
}

func (a *Api) WebUser(req *http.Request) (interface{}, error) {
	a.statsdClient.Incr("www.user.get", 1)
	user, err := a.getAuthenticatedUser(req)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (a *Api) WebDomains(req *http.Request) (interface{}, error) {
	a.statsdClient.Incr("www.domains", 1)
	user, err := a.getAuthenticatedUser(req)
	if err != nil {
		return nil, err
	}
	domains, err := a.domains.GetDomains(user)
	if err != nil {
		log.Println("unable to get domains for a user", err)
		return nil, errors.New("invalid request")
	}

	return domains, nil
}

func (a *Api) WebPremiumRequest(req *http.Request) (interface{}, error) {
	a.statsdClient.Incr("www.premium.request", 1)
	user, err := a.getAuthenticatedUser(req)
	if err != nil {
		return nil, err
	}

	err = a.users.RequestPremiumAccount(user)
	if err != nil {
		log.Println("unable to request premium account for a user", err)
		return nil, errors.New("invalid request")
	}

	return "OK", nil
}

func (a *Api) WebUserPasswordReset(req *http.Request) (interface{}, error) {
	a.statsdClient.Incr("www.user.reset_password", 1)
	request := model.UserPasswordResetRequest{}
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		log.Println("unable to parse domain acquire request", err)
		return nil, errors.New("invalid request")
	}
	user, err := a.users.GetUserByEmail(request.Email)
	if err != nil {
		log.Println("unable to get a user", err)
		return nil, errors.New("invalid request")
	}

	if user != nil && user.Active {
		action, err := a.actions.UpsertPasswordAction(user.Id)
		if err != nil {
			log.Println("unable to upsert action", err)
			return nil, errors.New("invalid request")
		}
		err = a.mail.SendResetPassword(user.Email, action.Token)
		if err != nil {
			log.Println("unable to send mail", err)
			return nil, errors.New("invalid request")
		}
	}

	return "Reset password requested", nil
}

func (a *Api) WebUserActivate(req *http.Request) (interface{}, error) {
	a.statsdClient.Incr("rest.user.activate", 1)
	request := model.UserActivateRequest{}
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		log.Println("unable to parse user activate request", err)
		return nil, errors.New("invalid request")
	}
	err = a.users.Activate(request.Token)
	if err != nil {
		log.Println("unable to activate user", err)
		return nil, errors.New("invalid request")
	}
	return "User was activated", nil
}

func (a *Api) DomainUpdate(req *http.Request) (interface{}, error) {
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

func (a *Api) DomainGet(req *http.Request) (interface{}, error) {
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

func (a *Api) getAuthenticatedUser(req *http.Request) (*model.User, error) {
	userEmail := req.Header.Get("RedirectUserEmail")
	if userEmail == "" {
		log.Println("no user session")
		return nil, errors.New("invalid request")
	}
	user, err := a.users.GetUserByEmail(userEmail)
	if err != nil {
		log.Println("unable to get a user", err)
		return nil, errors.New("invalid request")
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (a *Api) UserGet(req *http.Request) (interface{}, error) {

	a.statsdClient.Incr("rest.user.get", 1)
	emails, ok := req.URL.Query()["email"]
	if !ok {
		return nil, errors.New("no email")
	}
	passwords, ok := req.URL.Query()["password"]
	if !ok {
		return nil, errors.New("no password")
	}
	request := model.UserGetRequest{Email: &emails[0], Password: &passwords[0]}
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

	return &model.UserResponse{
		Email:           user.Email,
		Active:          user.Active,
		UpdateToken:     user.UpdateToken,
		Unsubscribed:    user.Unsubscribed,
		PremiumStatusId: user.PremiumStatusId,
		Timestamp:       user.Timestamp,
		Domains:         domains}, nil
}

func (a *Api) UserLog(req *http.Request) (interface{}, error) {
	a.statsdClient.Incr("rest.user.log", 1)

	err := req.ParseForm()
	if err != nil {
		return nil, errors.New("invalid request")
	}

	token := req.PostForm.Get("token")
	data := req.PostForm.Get("data")
	includeSupport := req.PostForm.Get("include_support") == "true"
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
