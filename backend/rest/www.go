package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/smira/go-statsd"
	"github.com/syncloud/redirect/model"
	"github.com/syncloud/redirect/service"
	"log"
	"net"
	"net/http"
	"os"
)

type Www struct {
	statsdClient   *statsd.Client
	domains        *service.Domains
	users          *service.Users
	actions        *service.Actions
	mail           *service.Mail
	probe          *service.PortProbe
	domain         string
	payPalPlanId   string
	payPalClientId string
	store          *sessions.CookieStore
}

func NewWww(statsdClient *statsd.Client, domains *service.Domains, users *service.Users, actions *service.Actions,
	mail *service.Mail, probe *service.PortProbe, domain string, payPalPlanId string, payPalClientId string, authSecretSey []byte) *Www {

	return &Www{statsdClient: statsdClient, domains: domains, users: users, actions: actions,
		mail: mail, probe: probe, domain: domain, payPalPlanId: payPalPlanId, payPalClientId: payPalClientId,
		store: sessions.NewCookieStore(authSecretSey),
	}
}

func (www *Www) StartWww(socket string) {

	r := mux.NewRouter()
	r.HandleFunc("/user/reset_password", Handle(www.WebUserPasswordReset)).Methods("POST")
	r.HandleFunc("/user/set_password", Handle(www.UserSetPassword)).Methods("POST")
	r.HandleFunc("/user/activate", Handle(www.WebUserActivate)).Methods("POST")
	r.HandleFunc("/user/create", Handle(www.UserCreateV2)).Methods("POST")
	r.HandleFunc("/user/login", Handle(www.UserLogin)).Methods("POST")

	r.HandleFunc("/logout", www.Secured(Handle(www.UserLogout))).Methods("POST")
	r.HandleFunc("/notification/enable", www.Secured(Handle(www.WebNotificationEnable))).Methods("POST")
	r.HandleFunc("/notification/disable", www.Secured(Handle(www.WebNotificationDisable))).Methods("POST")
	r.HandleFunc("/user", www.Secured(Handle(www.WebUserDelete))).Methods("DELETE")
	r.HandleFunc("/user", www.Secured(Handle(www.WebUser))).Methods("GET")
	r.HandleFunc("/domains", www.Secured(Handle(www.WebDomains))).Methods("GET")
	r.HandleFunc("/plan", www.Secured(Handle(www.WebPlan))).Methods("GET")
	r.HandleFunc("/plan/subscribe", www.Secured(Handle(www.WebPlanSubscribe))).Methods("POST")
	r.HandleFunc("/domain", www.Secured(Handle(www.DomainDelete))).Methods("DELETE")
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

func (www *Www) getSession(r *http.Request) (*sessions.Session, error) {
	return www.store.Get(r, "session")
}

func (www *Www) setSessionEmail(w http.ResponseWriter, r *http.Request, email string) error {
	session, err := www.getSession(r)
	if err != nil {
		return err
	}
	session.Values["email"] = email
	return session.Save(r, w)
}

func (www *Www) clearSessionEmail(w http.ResponseWriter, r *http.Request) error {
	session, err := www.getSession(r)
	if err != nil {
		return err
	}
	delete(session.Values, "email")
	return session.Save(r, w)
}

func (www *Www) getSessionEmail(r *http.Request) (*string, error) {
	session, err := www.getSession(r)
	if err != nil {
		return nil, err
	}
	email, found := session.Values["email"]
	if !found {
		return nil, fmt.Errorf("no session found")
	}

	emailString := email.(string)
	return &emailString, nil
}

func (www *Www) getSessionUser(r *http.Request) (*model.User, error) {
	email, err := www.getSessionEmail(r)
	if err != nil {
		return nil, err
	}
	user, err := www.users.GetUserByEmail(*email)
	if err != nil {
		log.Println("unable to get a user", err)
		return nil, errors.New("invalid request")
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (www *Www) Secured(handle func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := www.getSessionEmail(r)
		if err != nil {
			log.Printf("error %v", err)
			fail(w, model.NewServiceErrorWithCode("Unauthorized", 401))
			return
		}
		handle(w, r)
	}
}

func (www *Www) WebNotificationEnable(_ http.ResponseWriter, req *http.Request) (interface{}, error) {
	www.statsdClient.Incr("www.notification.enable", 1)
	user, err := www.getSessionUser(req)
	if err != nil {
		return nil, err
	}
	user.NotificationEnabled = true
	return "OK", www.users.Save(user)
}

func (www *Www) WebNotificationDisable(_ http.ResponseWriter, req *http.Request) (interface{}, error) {
	www.statsdClient.Incr("www.notification.disable", 1)
	user, err := www.getSessionUser(req)
	if err != nil {
		return nil, err
	}
	user.NotificationEnabled = false
	return "OK", www.users.Save(user)
}

func (www *Www) WebUserDelete(_ http.ResponseWriter, req *http.Request) (interface{}, error) {
	www.statsdClient.Incr("www.user.delete", 1)
	user, err := www.getSessionUser(req)
	if err != nil {
		return nil, err
	}

	err = www.domains.DeleteAllDomains(user.Id)
	if err != nil {
		log.Println("unable to delete domains for a user", err)
		return nil, errors.New("invalid request")
	}
	err = www.actions.DeleteActions(user.Id)
	if err != nil {
		log.Println("unable to delete actions for a user", err)
		return nil, errors.New("invalid request")
	}

	err = www.users.Delete(user.Id)
	if err != nil {
		log.Println("unable to delete a user", err)
		return nil, errors.New("invalid request")
	}

	return "OK", nil
}

func (www *Www) WebUser(_ http.ResponseWriter, req *http.Request) (interface{}, error) {
	www.statsdClient.Incr("www.user.get", 1)
	user, err := www.getSessionUser(req)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (www *Www) WebDomains(_ http.ResponseWriter, req *http.Request) (interface{}, error) {
	www.statsdClient.Incr("www.domains", 1)
	user, err := www.getSessionUser(req)
	if err != nil {
		return nil, err
	}
	domains, err := www.domains.GetDomains(user)
	if err != nil {
		log.Println("unable to get domains for a user", err)
		return nil, errors.New("invalid request")
	}

	return domains, nil
}

func (www *Www) WebPlan(http.ResponseWriter, *http.Request) (interface{}, error) {
	www.statsdClient.Incr("www.plan.get", 1)
	return model.PlanResponse{PlanId: www.payPalPlanId, ClientId: www.payPalClientId}, nil
}

func (www *Www) WebPlanSubscribe(_ http.ResponseWriter, req *http.Request) (interface{}, error) {
	www.statsdClient.Incr("www.plan.subscribe", 1)
	user, err := www.getSessionUser(req)
	if err != nil {
		return nil, err
	}
	request := model.PlanSubscribeRequest{}
	err = json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		log.Println("unable to parse plan subscribe request", err)
		return nil, errors.New("invalid request")
	}

	err = www.users.PlanSubscribe(user, request.SubscriptionId)
	if err != nil {
		log.Println("unable to do a plan subscribe for a user", err)
		return nil, errors.New("invalid request")
	}

	return "OK", nil
}

func (www *Www) WebUserPasswordReset(_ http.ResponseWriter, req *http.Request) (interface{}, error) {
	www.statsdClient.Incr("www.user.reset_password", 1)
	request := model.UserPasswordResetRequest{}
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		log.Println("unable to parse domain acquire request", err)
		return nil, errors.New("invalid request")
	}
	user, err := www.users.GetUserByEmail(request.Email)
	if err != nil {
		log.Println("unable to get a user", err)
		return nil, errors.New("invalid request")
	}

	if user != nil && user.Active {
		action, err := www.actions.UpsertPasswordAction(user.Id)
		if err != nil {
			log.Println("unable to upsert action", err)
			return nil, errors.New("invalid request")
		}
		err = www.mail.SendResetPassword(user.Email, action.Token)
		if err != nil {
			log.Println("unable to send mail", err)
			return nil, errors.New("invalid request")
		}
	}

	return "Reset password requested", nil
}

func (www *Www) requestIp(req *http.Request) (*string, error) {
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

func (www *Www) WebUserActivate(_ http.ResponseWriter, req *http.Request) (interface{}, error) {
	www.statsdClient.Incr("web.user.activate", 1)
	request := model.UserActivateRequest{}
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		log.Println("unable to parse user activate request", err)
		return nil, errors.New("invalid request")
	}
	err = www.users.Activate(request.Token)
	if err != nil {
		log.Println("unable to activate user", err)
		return nil, errors.New("invalid request")
	}
	return "User was activated", nil
}

func (www *Www) UserCreateV2(_ http.ResponseWriter, req *http.Request) (interface{}, error) {
	www.statsdClient.Incr("www.user.create", 1)
	request := model.UserCreateRequest{}
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		log.Println("unable to parse user create request", err)
		return nil, errors.New("invalid request")
	}

	return www.users.CreateNewUser(request)
}

func (www *Www) DomainDelete(_ http.ResponseWriter, req *http.Request) (interface{}, error) {
	www.statsdClient.Incr("www.domain.delete", 1)
	domain := req.URL.Query().Get("domain")
	if domain == "" {
		return nil, errors.New("missing domain")
	}
	user, err := www.getSessionUser(req)
	if err != nil {
		return nil, err
	}
	err = www.domains.DeleteDomain(user.Id, domain)
	return "Domain deleted", err
}

func (www *Www) UserSetPassword(_ http.ResponseWriter, req *http.Request) (interface{}, error) {
	www.statsdClient.Incr("www.user.set_password", 1)
	request := &model.UserPasswordSetRequest{}
	err := json.NewDecoder(req.Body).Decode(request)
	if err != nil {
		log.Println("unable to parse user set password request", err)
		return nil, errors.New("invalid request")
	}
	err = www.users.UserSetPassword(request)
	return "Password was set successfully", err
}

func (www *Www) UserLogin(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	www.statsdClient.Incr("www.user.login", 1)
	request := &model.UserAuthenticateRequest{}
	err := json.NewDecoder(r.Body).Decode(request)
	if err != nil {
		log.Println("unable to parse user set password request", err)
		return nil, errors.New("invalid request")
	}
	_, err = www.users.Authenticate(request.Email, request.Password)
	if err != nil {
		return nil, err
	}
	err = www.clearSessionEmail(w, r)
	err = www.setSessionEmail(w, r, *request.Email)
	return "User logged in", err
}

func (www *Www) UserLogout(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	www.statsdClient.Incr("www.user.logout", 1)
	err := www.clearSessionEmail(w, r)
	return "User logged out", err
}
