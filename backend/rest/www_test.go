package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/syncloud/redirect/log"
	"github.com/syncloud/redirect/metrics"
	"github.com/syncloud/redirect/model"
	"github.com/syncloud/redirect/product"
	"net/http"
	"net/http/httptest"
	"testing"
)

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

type WwwNsCheckerStub struct {
}

func (w WwwNsCheckerStub) Check(_ int64, _ string) (*model.NameServerCheckResult, error) {
	panic("implement me")
}

type WwwUsersStub struct {
	authenticated bool
}

func (w WwwUsersStub) GetUser(id int64) (*model.User, error) {
	return &model.User{Id: id, Email: "test@example.com"}, nil
}

func (w WwwUsersStub) GetUserByEmail(_ string) (*model.User, error) {
	panic("implement me")
}

func (w WwwUsersStub) CreateNewUser(_ model.UserCreateRequest) (*model.User, error) {
	panic("implement me")
}

func (w WwwUsersStub) Authenticate(email *string, _ *string) (*model.User, error) {
	if w.authenticated {
		return &model.User{Id: 1, Email: *email}, nil
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

func (w WwwUsersStub) Subscribe(_ *model.User, _ string, _ int, _ string) error {
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

type WwwStripeStub struct {
}

func (w WwwStripeStub) CreateCheckout(_ string) (string, error) {
	panic("implement me")
}

func (w WwwStripeStub) GetCheckoutSubscription(_ string) (string, string, error) {
	panic("implement me")
}

func (w WwwStripeStub) MaxEnabled() bool {
	return false
}

type WwwRelayStub struct {
}

func (w WwwRelayStub) UsedBytes(_ int64) (int64, error) {
	return 0, nil
}

func (w WwwRelayStub) LimitBytes(_ int64) int64 {
	return 0
}

func (w WwwRelayStub) Enabled(_ int64) (bool, error) {
	return false, nil
}

type WwwPayPalStub struct {
}

func (w WwwPayPalStub) PlanId(_ string) (string, error) {
	return "", nil
}

func (w WwwPayPalStub) Tier(_ string) string {
	return model.PlanPro
}

func (w WwwPayPalStub) Plans() model.PlanResponse {
	return model.PlanResponse{}
}

func TestLogin_CreateSession(t *testing.T) {

	www := NewWww(
		&WwwDomainsStub{},
		&WwwNsCheckerStub{},
		&WwwUsersStub{authenticated: true},
		&WwwActionsStub{},
		&WwwMailStub{},
		&WwwStripeStub{},
		&WwwOrdersStub{},
		&WwwRelayStub{},
		&WwwMailRelayStub{},
		&WwwPayPalStub{},
		metrics.New(),
		"example.com",
		[]byte("secret_key"),
		"",
		1000,
		10000,
		log.Default(),
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
		&WwwDomainsStub{},
		&WwwNsCheckerStub{},
		&WwwUsersStub{authenticated: true},
		&WwwActionsStub{},
		&WwwMailStub{},
		&WwwStripeStub{},
		&WwwOrdersStub{},
		&WwwRelayStub{},
		&WwwMailRelayStub{},
		&WwwPayPalStub{},
		metrics.New(),
		"example.com",
		[]byte("secret_key"),
		"",
		1000,
		10000,
		log.Default(),
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
		&WwwDomainsStub{},
		&WwwNsCheckerStub{},
		&WwwUsersStub{authenticated: true},
		&WwwActionsStub{},
		&WwwMailStub{},
		&WwwStripeStub{},
		&WwwOrdersStub{},
		&WwwRelayStub{},
		&WwwMailRelayStub{},
		&WwwPayPalStub{},
		metrics.New(),
		"example.com",
		[]byte("secret_key"),
		"",
		1000,
		10000,
		log.Default(),
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
		&WwwDomainsStub{},
		&WwwNsCheckerStub{},
		&WwwUsersStub{authenticated: true},
		&WwwActionsStub{},
		&WwwMailStub{},
		&WwwStripeStub{},
		&WwwOrdersStub{},
		&WwwRelayStub{},
		&WwwMailRelayStub{},
		&WwwPayPalStub{},
		metrics.New(),
		"example.com",
		[]byte("secret_key"),
		"",
		1000,
		10000,
		log.Default(),
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
	_, err = www.UserLogout(rr, req, model.User{})
	if err != nil {
		t.Fatal(err)
	}
	assert.Contains(t, rr.Header().Get("Set-Cookie"), "session=;")

}

type WwwMailRelayStub struct{}

func (s *WwwMailRelayStub) UsedMessages(_ int64) (int64, error) { return 0, nil }
func (s *WwwMailRelayStub) LimitMessages(_ int64) int64         { return 0 }
func (s *WwwMailRelayStub) Enabled(_ int64) (bool, error)       { return false, nil }

func sessionWww() *Www {
	return NewWww(
		&WwwDomainsStub{},
		&WwwNsCheckerStub{},
		&WwwUsersStub{authenticated: true},
		&WwwActionsStub{},
		&WwwMailStub{},
		&WwwStripeStub{},
		&WwwOrdersStub{},
		&WwwRelayStub{},
		&WwwMailRelayStub{},
		&WwwPayPalStub{},
		metrics.New(),
		"example.com",
		[]byte("secret_key"),
		"",
		1000,
		10000,
		log.Default(),
	)
}

func TestSessionCarryingOnlyAnEmailIsRejected(t *testing.T) {
	www := sessionWww()
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp := httptest.NewRecorder()
	session, err := www.getSession(req)
	if err != nil {
		t.Fatal(err)
	}
	session.Values["email"] = "test@example.com"
	if err := session.Save(req, resp); err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Cookie", resp.Header().Get("Set-Cookie"))

	_, err = www.getSessionUser(req)
	assert.Error(t, err)
}

func TestSessionCarryingAUserIdIsAccepted(t *testing.T) {
	www := sessionWww()
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp := httptest.NewRecorder()
	if err := www.setSessionUser(resp, req, int64(42)); err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Cookie", resp.Header().Get("Set-Cookie"))

	user, err := www.getSessionUser(req)
	assert.NoError(t, err)
	assert.Equal(t, int64(42), user.Id)
}

type WwwOrdersStub struct{}

func (s *WwwOrdersStub) Catalog() []product.Device { return nil }

func (s *WwwOrdersStub) Shipping() int { return 1500 }

func (s *WwwOrdersStub) Start(_ *product.Order, _ string) (string, error) { return "", nil }

func (s *WwwOrdersStub) Complete(_ int64, _ string) error { return nil }
