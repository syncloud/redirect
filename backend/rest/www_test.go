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

func (n StatsdClientStub) Incr(_ string, _ int64, _ ...statsd.Tag) {
}

type WwwDomainsStub struct {
}

func (w WwwDomainsStub) DeleteDomain(_ int64, _ string) error {
	panic("implement me")
}

func (w WwwDomainsStub) GetDomains(_ *model.User) ([]*model.Domain, error) {
	panic("implement me")
}

func (w WwwDomainsStub) DeleteAllDomains(_ int64) error {
	panic("implement me")
}

type WwwUsersStub struct {
	authenticated bool
}

func (w WwwUsersStub) GetUserByEmail(_ string) (*model.User, error) {
	panic("implement me")
}

func (w WwwUsersStub) CreateNewUser(_ model.UserCreateRequest) (*model.User, error) {
	panic("implement me")
}

func (w WwwUsersStub) Authenticate(email *string, _ *string) (*model.User, error) {
	if w.authenticated {
		return &model.User{Email: *email}, nil
	} else {
		return nil, fmt.Errorf("not authenticated")
	}
}

func (w WwwUsersStub) UserSetPassword(_ *model.UserPasswordSetRequest) error {
	panic("implement me")
}

func (w WwwUsersStub) Save(_ *model.User) error {
	panic("implement me")
}

func (w WwwUsersStub) Subscribe(_ *model.User, _ string, _ int) error {
	panic("implement me")
}

func (w WwwUsersStub) Unsubscribe(_ *model.User) error {
	panic("implement me")
}

func (w WwwUsersStub) Activate(_ string) error {
	panic("implement me")
}

func (w WwwUsersStub) Delete(_ int64) error {
	panic("implement me")
}

type WwwActionsStub struct {
}

func (w WwwActionsStub) UpsertPasswordAction(_ int64) (*model.Action, error) {
	panic("implement me")
}

type WwwMailStub struct {
}

func (w WwwMailStub) SendResetPassword(_ string, _ string) error {
	panic("implement me")
}

func TestLogin_CreateSession(t *testing.T) {

	www := NewWww(
		&StatsdClientStub{},
		&WwwDomainsStub{},
		&WwwUsersStub{authenticated: true},
		&WwwActionsStub{},
		&WwwMailStub{},
		"example.com",
		"paypal_plan_monthly_id",
		"paypal_plan_annual_id",
		"paypal_client_id",
		[]byte("secret_key"),
		"",
	)
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

func TestLoginAgain_NotError(t *testing.T) {

	www := NewWww(
		&StatsdClientStub{},
		&WwwDomainsStub{},
		&WwwUsersStub{authenticated: true},
		&WwwActionsStub{},
		&WwwMailStub{},
		"example.com",
		"paypal_plan_monthly_id",
		"paypal_plan_annual_id",
		"paypal_client_id",
		[]byte("secret_key"),
		"",
	)
	email := "test@example.com"
	password := "password"
	user := &model.UserAuthenticateRequest{Email: &email, Password: &password}
	userJson, err := json.Marshal(user)
	if err != nil {
		t.Fatal(err)
	}
	body1 := bytes.NewBuffer(userJson)
	req1, err := http.NewRequest("GET", "/", body1)
	if err != nil {
		t.Fatal(err)
	}
	//req.AddCookie(&http.Cookie{Name: "session", Value: "123"})
	rr1 := httptest.NewRecorder()
	_, err = www.UserLogin(rr1, req1)
	if err != nil {
		t.Fatal(err)
	}
	session1 := rr1.Header().Get("Set-Cookie")
	assert.Contains(t, session1, "session=")

	body2 := bytes.NewBuffer(userJson)
	req2, err := http.NewRequest("GET", "/", body2)
	if err != nil {
		t.Fatal(err)
	}
	req2.AddCookie(&http.Cookie{Name: "session", Value: "MTYyOTY3MDk0OHxEdi1CQkFFQ180SUFBUkFCRUFBQUJQLUNBQUE9fEZHUw9y4LnPQcECsWcJCSehnQXkmZM0nJrMDfjsaXsW"})
	rr2 := httptest.NewRecorder()
	_, err = www.UserLogin(rr2, req2)
	if err != nil {
		t.Fatal(err)
	}
	session2 := rr2.Header().Get("Set-Cookie")
	assert.Contains(t, session2, "session=")

}

func TestLoginFresh_NotError(t *testing.T) {

	www := NewWww(
		&StatsdClientStub{},
		&WwwDomainsStub{},
		&WwwUsersStub{authenticated: true},
		&WwwActionsStub{},
		&WwwMailStub{},
		"example.com",
		"paypal_plan_monthly_id",
		"paypal_plan_annual_id",
		"paypal_client_id",
		[]byte("secret_key"),
		"",
	)
	email := "test@example.com"
	password := "password"
	user := &model.UserAuthenticateRequest{Email: &email, Password: &password}
	userJson, err := json.Marshal(user)
	if err != nil {
		t.Fatal(err)
	}
	body1 := bytes.NewBuffer(userJson)
	req1, err := http.NewRequest("GET", "/", body1)
	if err != nil {
		t.Fatal(err)
	}
	req1.AddCookie(&http.Cookie{Name: "session", Value: "eyJfZnJlc2giOmZhbHNlLCJfaWQiOnsiIGIiOiJZVE16WVdFNVlUVmhaVGcwTTJGbVpXUTNPV1JsWldZMlpXVmpNbVZqWmpNPSJ9LCJ1c2VyX2lkIjoicmliYWxraW5AZ21haWwuY29tIn0.YPnjUw.oTdMJAFq_zIxUuLmduu9McEbtVs"})
	rr1 := httptest.NewRecorder()
	_, err = www.UserLogin(rr1, req1)
	if err != nil {
		t.Fatal(err)
	}
	session1 := rr1.Header().Get("Set-Cookie")
	assert.Contains(t, session1, "session=")

}

func TestLogout_ClearSession(t *testing.T) {

	www := NewWww(
		&StatsdClientStub{},
		&WwwDomainsStub{},
		&WwwUsersStub{authenticated: true},
		&WwwActionsStub{},
		&WwwMailStub{},
		"example.com",
		"paypal_plan_monthly_id",
		"paypal_plan_annual_id",
		"paypal_client_id",
		[]byte("secret_key"),
		"",
	)
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
