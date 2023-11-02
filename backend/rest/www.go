package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/syncloud/redirect/metrics"
	"github.com/syncloud/redirect/model"
	"golang.org/x/net/netutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

type WwwDomains interface {
	DeleteDomain(userId int64, domainName string) error
	GetDomains(user *model.User) ([]*model.Domain, error)
	DeleteAllDomains(userId int64) error
}

type WwwUsers interface {
	GetUserByEmail(userEmail string) (*model.User, error)
	CreateNewUser(request model.UserCreateRequest) (*model.User, error)
	Authenticate(email *string, password *string) (*model.User, error)
	UserSetPassword(request *model.UserPasswordSetRequest) error
	Save(user *model.User) error
	PlanSubscribe(user *model.User, subscriptionId string) error
	PlanUnSubscribe(user *model.User) error
	Activate(token string) error
	Delete(userId int64) error
}

type WwwActions interface {
	DeleteActions(userId int64) error
	UpsertPasswordAction(userId int64) (*model.Action, error)
}

type WwwMail interface {
	SendResetPassword(to string, token string) error
}

type Www struct {
	statsdClient        metrics.StatsdClient
	domains             WwwDomains
	users               WwwUsers
	actions             WwwActions
	mail                WwwMail
	domain              string
	payPalPlanMonthlyId string
	payPalPlanAnnualId  string
	payPalClientId      string
	store               *sessions.CookieStore
	count404            int64
	socket              string
}

func NewWww(
	statsdClient metrics.StatsdClient,
	domains WwwDomains,
	users WwwUsers,
	actions WwwActions,
	mail WwwMail,
	domain string,
	payPalPlanMonthlyId string,
	payPalPlanAnnualId string,
	payPalClientId string,
	authSecretSey []byte,
	socket string,
) *Www {
	return &Www{
		statsdClient:        statsdClient,
		domains:             domains,
		users:               users,
		actions:             actions,
		mail:                mail,
		domain:              domain,
		payPalPlanMonthlyId: payPalPlanMonthlyId,
		payPalPlanAnnualId:  payPalPlanAnnualId,
		payPalClientId:      payPalClientId,
		store:               sessions.NewCookieStore(authSecretSey),
		socket:              socket,
	}
}

func (w *Www) Start() error {

	r := mux.NewRouter()
	r.HandleFunc("/user/reset_password", Handle(w.WebUserPasswordReset)).Methods("POST")
	r.HandleFunc("/user/set_password", Handle(w.UserSetPassword)).Methods("POST")
	r.HandleFunc("/user/activate", Handle(w.WebUserActivate)).Methods("POST")
	r.HandleFunc("/user/create", Handle(w.UserCreateV2)).Methods("POST")
	r.HandleFunc("/user/login", Handle(w.UserLogin)).Methods("POST")

	r.HandleFunc("/logout", w.Secured(Handle(w.UserLogout))).Methods("POST")
	r.HandleFunc("/notification/enable", w.Secured(Handle(w.WebNotificationEnable))).Methods("POST")
	r.HandleFunc("/notification/disable", w.Secured(Handle(w.WebNotificationDisable))).Methods("POST")
	r.HandleFunc("/user", w.Secured(Handle(w.WebUserDelete))).Methods("DELETE")
	r.HandleFunc("/user", w.Secured(Handle(w.WebUser))).Methods("GET")
	r.HandleFunc("/domains", w.Secured(Handle(w.WebDomains))).Methods("GET")
	r.HandleFunc("/plan", w.Secured(Handle(w.WebPlan))).Methods("GET")
	r.HandleFunc("/plan", w.Secured(Handle(w.WebPlanUnsubscribe))).Methods("DELETE")
	r.HandleFunc("/plan/subscribe", w.Secured(Handle(w.WebPlanSubscribe))).Methods("POST")
	r.HandleFunc("/domain", w.Secured(Handle(w.DomainDelete))).Methods("DELETE")
	r.NotFoundHandler = http.HandlerFunc(w.notFoundHandler)

	r.Use(headers)

	var listener net.Listener
	if strings.HasPrefix(w.socket, "tcp://") {
		address := strings.TrimPrefix(w.socket, "tcp://")
		tcpListener, err := net.Listen("tcp", address)
		if err != nil {
			return err
		}
		listener = tcpListener
	} else {
		if _, err := os.Stat(w.socket); err == nil {
			err := os.Remove(w.socket)
			if err != nil {
				return err
			}
		}
		unixListener, err := net.Listen("unix", w.socket)
		if err != nil {
			return err
		}
		if err := os.Chmod(w.socket, 0777); err != nil {
			return err
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
	log.Printf("Started backend (%s)\n", w.socket)
	return srv.Serve(l)
}

func (w *Www) getSession(r *http.Request) (*sessions.Session, error) {
	return w.store.Get(r, "session")
}

func (w *Www) setSessionEmail(resp http.ResponseWriter, r *http.Request, email string) error {
	session, err := w.getSession(r)
	if err != nil {
		return err
	}
	session.Values["email"] = email
	return session.Save(r, resp)
}

func (w *Www) clearSessionEmail(resp http.ResponseWriter, r *http.Request) error {
	r.Header.Del("Cookie")
	session, err := w.getSession(r)
	if err != nil {
		return err
	}
	delete(session.Values, "email")
	return session.Save(r, resp)
}

func (w *Www) getSessionEmail(r *http.Request) (*string, error) {
	session, err := w.getSession(r)
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

func (w *Www) getSessionUser(r *http.Request) (*model.User, error) {
	email, err := w.getSessionEmail(r)
	if err != nil {
		return nil, err
	}
	user, err := w.users.GetUserByEmail(*email)
	if err != nil {
		log.Println("unable to get a user", err)
		return nil, errors.New("invalid request")
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (w *Www) Secured(handle func(_ http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(resp http.ResponseWriter, r *http.Request) {
		_, err := w.getSessionEmail(r)
		if err != nil {
			log.Printf("error %v", err)
			fail(resp, model.NewServiceErrorWithCode("Unauthorized", 401))
			return
		}
		handle(resp, r)
	}
}

func (w *Www) WebNotificationEnable(_ http.ResponseWriter, req *http.Request) (interface{}, error) {
	w.statsdClient.Incr("www.notification.enable", 1)
	user, err := w.getSessionUser(req)
	if err != nil {
		return nil, err
	}
	user.NotificationEnabled = true
	return "OK", w.users.Save(user)
}

func (w *Www) WebNotificationDisable(_ http.ResponseWriter, req *http.Request) (interface{}, error) {
	w.statsdClient.Incr("www.notification.disable", 1)
	user, err := w.getSessionUser(req)
	if err != nil {
		return nil, err
	}
	user.NotificationEnabled = false
	return "OK", w.users.Save(user)
}

func (w *Www) WebUserDelete(_ http.ResponseWriter, req *http.Request) (interface{}, error) {
	w.statsdClient.Incr("www.user.delete", 1)
	user, err := w.getSessionUser(req)
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

func (w *Www) WebUser(_ http.ResponseWriter, req *http.Request) (interface{}, error) {
	w.statsdClient.Incr("www.user.get", 1)
	user, err := w.getSessionUser(req)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (w *Www) WebDomains(_ http.ResponseWriter, req *http.Request) (interface{}, error) {
	w.statsdClient.Incr("www.domains", 1)
	user, err := w.getSessionUser(req)
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

func (w *Www) WebPlan(http.ResponseWriter, *http.Request) (interface{}, error) {
	w.statsdClient.Incr("www.plan.get", 1)
	return model.PlanResponse{
		PlanMonthlyId: w.payPalPlanMonthlyId,
		PlanAnnualId:  w.payPalPlanAnnualId,
		ClientId:      w.payPalClientId,
	}, nil
}

func (w *Www) WebPlanUnsubscribe(_ http.ResponseWriter, req *http.Request) (interface{}, error) {
	w.statsdClient.Incr("www.plan.unsubscribe", 1)
	user, err := w.getSessionUser(req)
	if err != nil {
		return nil, err
	}

	err = w.users.PlanUnSubscribe(user)
	if err != nil {
		log.Println("unable to unsubscribe a plan subscribe for a user", err)
		return nil, errors.New("invalid request")
	}

	return "OK", nil
}

func (w *Www) WebPlanSubscribe(_ http.ResponseWriter, req *http.Request) (interface{}, error) {
	w.statsdClient.Incr("www.plan.subscribe", 1)
	user, err := w.getSessionUser(req)
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

func (w *Www) WebUserPasswordReset(_ http.ResponseWriter, req *http.Request) (interface{}, error) {
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

func (w *Www) WebUserActivate(_ http.ResponseWriter, req *http.Request) (interface{}, error) {
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

func (w *Www) UserCreateV2(_ http.ResponseWriter, req *http.Request) (interface{}, error) {
	w.statsdClient.Incr("www.user.create", 1)
	request := model.UserCreateRequest{}
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		log.Println("unable to parse user create request", err)
		return nil, errors.New("invalid request")
	}

	return w.users.CreateNewUser(request)
}

func (w *Www) DomainDelete(_ http.ResponseWriter, req *http.Request) (interface{}, error) {
	w.statsdClient.Incr("www.domain.delete", 1)
	domain := req.URL.Query().Get("domain")
	if domain == "" {
		return nil, errors.New("missing domain")
	}
	user, err := w.getSessionUser(req)
	if err != nil {
		return nil, err
	}
	err = w.domains.DeleteDomain(user.Id, domain)
	return "Domain deleted", err
}

func (w *Www) UserSetPassword(_ http.ResponseWriter, req *http.Request) (interface{}, error) {
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

func (w *Www) UserLogin(resp http.ResponseWriter, r *http.Request) (interface{}, error) {
	w.statsdClient.Incr("www.user.login", 1)
	request := &model.UserAuthenticateRequest{}
	err := json.NewDecoder(r.Body).Decode(request)
	if err != nil {
		log.Println("unable to parse user login request", err)
		return nil, errors.New("invalid request")
	}
	_, err = w.users.Authenticate(request.Email, request.Password)
	if err != nil {
		return nil, err
	}
	err = w.clearSessionEmail(resp, r)
	err = w.setSessionEmail(resp, r, *request.Email)
	return "User logged in", err
}

func (w *Www) UserLogout(resp http.ResponseWriter, r *http.Request) (interface{}, error) {
	w.statsdClient.Incr("www.user.logout", 1)
	http.SetCookie(resp, &http.Cookie{Name: "session", Value: "", MaxAge: -1})
	err := w.clearSessionEmail(resp, r)
	return "User logged out", err
}

func (w *Www) notFoundHandler(resp http.ResponseWriter, r *http.Request) {
	w.count404++
	if w.count404%100 == 0 {
		log.Printf("404 counter: %v\n", w.count404)
	}
	http.NotFound(resp, r)
}
