package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/smira/go-statsd"
	"github.com/stretchr/testify/assert"
	"github.com/syncloud/redirect/model"
	"net/http"
	"net/http/httptest"
	"testing"
)

type StatsdClientStub struct {
}

func (n StatsdClientStub) Incr(stat string, count int64, tags ...statsd.Tag) {
}

type WwwDomainsStub struct {
}

func (w WwwDomainsStub) DeleteDomain(userId int64, domainName string) error {
	panic("implement me")
}

func (w WwwDomainsStub) GetDomains(user *model.User) ([]*model.Domain, error) {
	panic("implement me")
}

func (w WwwDomainsStub) DeleteAllDomains(userId int64) error {
	panic("implement me")
}

type WwwUsersStub struct {
	authenticated bool
}

func (w WwwUsersStub) GetUserByEmail(userEmail string) (*model.User, error) {
	panic("implement me")
}

func (w WwwUsersStub) CreateNewUser(request model.UserCreateRequest) (*model.User, error) {
	panic("implement me")
}

func (w WwwUsersStub) Authenticate(email *string, password *string) (*model.User, error) {
	if w.authenticated {
		return &model.User{Email: *email}, nil
	} else {
		return nil, fmt.Errorf("not authenticated")
	}
}

func (w WwwUsersStub) UserSetPassword(request *model.UserPasswordSetRequest) error {
	panic("implement me")
}

func (w WwwUsersStub) Save(user *model.User) error {
	panic("implement me")
}

func (w WwwUsersStub) PlanSubscribe(user *model.User, subscriptionId string) error {
	panic("implement me")
}

func (w WwwUsersStub) Activate(token string) error {
	panic("implement me")
}

func (w WwwUsersStub) Delete(userId int64) error {
	panic("implement me")
}

type WwwActionsStub struct {
}

func (w WwwActionsStub) DeleteActions(userId int64) error {
	panic("implement me")
}

func (w WwwActionsStub) UpsertPasswordAction(userId int64) (*model.Action, error) {
	panic("implement me")
}

type WwwMailStub struct {
}

func (w WwwMailStub) SendResetPassword(to string, token string) error {
	panic("implement me")
}

func TestLogin_CreateSession(t *testing.T) {

	www := NewWww(
		&StatsdClientStub{}, &WwwDomainsStub{}, &WwwUsersStub{authenticated: true}, &WwwActionsStub{}, &WwwMailStub{},
		"example.com", "paypal_plan_id", "paypal_client_id", []byte("secret_key"))
	email := "test@example.com"
	password := "password"
	user := &model.UserAuthenticateRequest{Email: &email, Password: &password}
	userJson, err := json.Marshal(user)
	if err != nil {
		t.Fatal(err)
	}
	body := bytes.NewBuffer(userJson)
	req, err := http.NewRequest("GET", "/", body)
	if err != nil {
		t.Fatal(err)
	}
	//req.AddCookie(&http.Cookie{Name: "session", Value: "123"})
	rr := httptest.NewRecorder()
	_, err = www.UserLogin(rr, req)
	if err != nil {
		t.Fatal(err)
	}
	assert.Contains(t, rr.Header().Get("Set-Cookie"), "session=")

}

func TestLogout_ClearSession(t *testing.T) {

	www := NewWww(
		&StatsdClientStub{}, &WwwDomainsStub{}, &WwwUsersStub{authenticated: true}, &WwwActionsStub{}, &WwwMailStub{},
		"example.com", "paypal_plan_id", "paypal_client_id", []byte("secret_key"))
	email := "test@example.com"
	password := "password"
	user := &model.UserAuthenticateRequest{Email: &email, Password: &password}
	userJson, err := json.Marshal(user)
	if err != nil {
		t.Fatal(err)
	}
	body := bytes.NewBuffer(userJson)
	req, err := http.NewRequest("GET", "/", body)
	if err != nil {
		t.Fatal(err)
	}
	req.AddCookie(&http.Cookie{Name: "session", Value: "MTYyOTY3MDk0OHxEdi1CQkFFQ180SUFBUkFCRUFBQUJQLUNBQUE9fEZHUw9y4LnPQcECsWcJCSehnQXkmZM0nJrMDfjsaXsW"})
	rr := httptest.NewRecorder()
	_, err = www.UserLogout(rr, req)
	if err != nil {
		t.Fatal(err)
	}
	assert.Contains(t, rr.Header().Get("Set-Cookie"), "session=;")

}
