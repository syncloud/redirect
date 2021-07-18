package rest

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
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
}

func NewWww(statsdClient *statsd.Client, domains *service.Domains, users *service.Users, actions *service.Actions, mail *service.Mail, probe *service.PortProbe, domain string, payPalPlanId string, payPalClientId string) *Www {
	return &Www{statsdClient: statsdClient, domains: domains, users: users, actions: actions, mail: mail, probe: probe, domain: domain, payPalPlanId: payPalPlanId, payPalClientId: payPalClientId}
}

func (w *Www) StartWww(socket string) {
	r := mux.NewRouter()
	r.HandleFunc("/web/notification/enable", Handle(w.WebNotificationEnable)).Methods("POST")
	r.HandleFunc("/web/notification/disable", Handle(w.WebNotificationDisable)).Methods("POST")
	r.HandleFunc("/web/user", Handle(w.WebUserDelete)).Methods("DELETE")
	r.HandleFunc("/web/user", Handle(w.WebUser)).Methods("GET")
	r.HandleFunc("/web/domains", Handle(w.WebDomains)).Methods("GET")
	r.HandleFunc("/web/plan", Handle(w.WebPlan)).Methods("GET")
	r.HandleFunc("/web/plan/subscribe", Handle(w.WebPlanSubscribe)).Methods("POST")
	r.HandleFunc("/web/user/reset_password", Handle(w.WebUserPasswordReset)).Methods("POST")
	r.HandleFunc("/web/user/set_password", Handle(w.UserSetPassword)).Methods("POST")
	r.HandleFunc("/web/user/activate", Handle(w.WebUserActivate)).Methods("POST")
	r.HandleFunc("/web/user/create", Handle(w.UserCreateV2)).Methods("POST")
	r.HandleFunc("/web/user/login", Handle(w.UserLogin)).Methods("POST")
	r.HandleFunc("/web/domain", Handle(w.DomainDelete)).Methods("DELETE")

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
	log.Printf("Started backend (%s)\n", socket)
	_ = http.Serve(unixListener, r)

}

func (w *Www) WebNotificationEnable(req *http.Request) (interface{}, error) {
	w.statsdClient.Incr("www.notification.enable", 1)
	user, err := w.getAuthenticatedUser(req)
	if err != nil {
		return nil, err
	}
	user.NotificationEnabled = true
	return "OK", w.users.Save(user)
}

func (w *Www) WebNotificationDisable(req *http.Request) (interface{}, error) {
	w.statsdClient.Incr("www.notification.disable", 1)
	user, err := w.getAuthenticatedUser(req)
	if err != nil {
		return nil, err
	}
	user.NotificationEnabled = false
	return "OK", w.users.Save(user)
}

func (w *Www) WebUserDelete(req *http.Request) (interface{}, error) {
	w.statsdClient.Incr("www.user.delete", 1)
	user, err := w.getAuthenticatedUser(req)
	if err != nil {
		return nil, err
	}

	err = w.domains.DeleteAllDomains(user.Id)
	if err != nil {
		log.Println("unable to delete domains for a user", err)
		return nil, errors.New("invalid request")
	}
	err = w.actions.DeleteActions(user.Id)
	if err != nil {
		log.Println("unable to delete actions for a user", err)
		return nil, errors.New("invalid request")
	}

	err = w.users.Delete(user.Id)
	if err != nil {
		log.Println("unable to delete a user", err)
		return nil, errors.New("invalid request")
	}

	return "OK", nil
}

func (w *Www) WebUser(req *http.Request) (interface{}, error) {
	w.statsdClient.Incr("www.user.get", 1)
	user, err := w.getAuthenticatedUser(req)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (w *Www) WebDomains(req *http.Request) (interface{}, error) {
	w.statsdClient.Incr("www.domains", 1)
	user, err := w.getAuthenticatedUser(req)
	if err != nil {
		return nil, err
	}
	domains, err := w.domains.GetDomains(user)
	if err != nil {
		log.Println("unable to get domains for a user", err)
		return nil, errors.New("invalid request")
	}

	return domains, nil
}

func (w *Www) WebPlan(req *http.Request) (interface{}, error) {
	w.statsdClient.Incr("www.plan.get", 1)
	return model.PlanResponse{PlanId: w.payPalPlanId, ClientId: w.payPalClientId}, nil
}

func (w *Www) WebPlanSubscribe(req *http.Request) (interface{}, error) {
	w.statsdClient.Incr("www.plan.subscribe", 1)
	user, err := w.getAuthenticatedUser(req)
	if err != nil {
		return nil, err
	}
	request := model.PlanSubscribeRequest{}
	err = json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		log.Println("unable to parse plan subscribe request", err)
		return nil, errors.New("invalid request")
	}

	err = w.users.PlanSubscribe(user, request.SubscriptionId)
	if err != nil {
		log.Println("unable to do a plan subscribe for a user", err)
		return nil, errors.New("invalid request")
	}

	return "OK", nil
}

func (w *Www) WebUserPasswordReset(req *http.Request) (interface{}, error) {
	w.statsdClient.Incr("www.user.reset_password", 1)
	request := model.UserPasswordResetRequest{}
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		log.Println("unable to parse domain acquire request", err)
		return nil, errors.New("invalid request")
	}
	user, err := w.users.GetUserByEmail(request.Email)
	if err != nil {
		log.Println("unable to get a user", err)
		return nil, errors.New("invalid request")
	}

	if user != nil && user.Active {
		action, err := w.actions.UpsertPasswordAction(user.Id)
		if err != nil {
			log.Println("unable to upsert action", err)
			return nil, errors.New("invalid request")
		}
		err = w.mail.SendResetPassword(user.Email, action.Token)
		if err != nil {
			log.Println("unable to send mail", err)
			return nil, errors.New("invalid request")
		}
	}

	return "Reset password requested", nil
}

func (w *Www) requestIp(req *http.Request) (*string, error) {
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

func (w *Www) getAuthenticatedUser(req *http.Request) (*model.User, error) {
	userEmail := req.Header.Get("RedirectUserEmail")
	if userEmail == "" {
		log.Println("no user session")
		return nil, errors.New("invalid request")
	}
	user, err := w.users.GetUserByEmail(userEmail)
	if err != nil {
		log.Println("unable to get a user", err)
		return nil, errors.New("invalid request")
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (w *Www) WebUserActivate(req *http.Request) (interface{}, error) {
	w.statsdClient.Incr("web.user.activate", 1)
	request := model.UserActivateRequest{}
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		log.Println("unable to parse user activate request", err)
		return nil, errors.New("invalid request")
	}
	err = w.users.Activate(request.Token)
	if err != nil {
		log.Println("unable to activate user", err)
		return nil, errors.New("invalid request")
	}
	return "User was activated", nil
}

func (w *Www) UserCreateV2(req *http.Request) (interface{}, error) {
	w.statsdClient.Incr("www.user.create", 1)
	request := model.UserCreateRequest{}
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		log.Println("unable to parse user create request", err)
		return nil, errors.New("invalid request")
	}

	return w.users.CreateNewUser(request)
}

func (w *Www) DomainDelete(req *http.Request) (interface{}, error) {
	w.statsdClient.Incr("www.domain.delete", 1)
	domain := req.URL.Query().Get("domain")
	if domain == "" {
		return nil, errors.New("missing domain")
	}
	user, err := w.getAuthenticatedUser(req)
	if err != nil {
		return nil, err
	}
	err = w.domains.DeleteDomain(user.Id, domain)
	return "Domain deleted", err
}

func (w *Www) UserSetPassword(req *http.Request) (interface{}, error) {
	w.statsdClient.Incr("www.user.set_password", 1)
	request := &model.UserPasswordSetRequest{}
	err := json.NewDecoder(req.Body).Decode(request)
	if err != nil {
		log.Println("unable to parse user set password request", err)
		return nil, errors.New("invalid request")
	}
	err = w.users.UserSetPassword(request)
	return "Password was set successfully", err
}

func (w *Www) UserLogin(req *http.Request) (interface{}, error) {
	w.statsdClient.Incr("www.user.login", 1)
	request := &model.UserAuthenticateRequest{}
	err := json.NewDecoder(req.Body).Decode(request)
	if err != nil {
		log.Println("unable to parse user set password request", err)
		return nil, errors.New("invalid request")
	}
	_, err = w.users.Authenticate(request.Email, request.Password)
	if err != nil {
		return nil, err
	}
	return "User logged in", nil
}
