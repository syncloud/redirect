package rest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/syncloud/redirect/event"
	"github.com/syncloud/redirect/grant"
	"github.com/syncloud/redirect/log"
	"github.com/syncloud/redirect/metrics"
	"github.com/syncloud/redirect/model"
	"github.com/syncloud/redirect/product"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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

func (w WwwStripeStub) PlanInfo(_ string) (string, string, error) {
	return model.PeriodMonth, model.PlanPro, nil
}

func (w WwwStripeStub) Switch(_ string) (string, error) {
	return "", nil
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

func (w WwwPayPalStub) Plans() model.PlanResponse {
	return model.PlanResponse{}
}

func (w WwwPayPalStub) PlanInfo(_ string) (string, string, error) {
	return model.PeriodMonth, model.PlanPro, nil
}

func (w WwwPayPalStub) SwitchToAnnual(_ string) (string, error) {
	return "", nil
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
		&WwwGrantsStub{},
		&WwwRelayStub{},
		&WwwMailRelayStub{},
		&WwwPayPalStub{},
		metrics.New(),
		"example.com",
		[]byte("secret_key"),
		"",
		1000,
		10000,
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
		&WwwGrantsStub{},
		&WwwRelayStub{},
		&WwwMailRelayStub{},
		&WwwPayPalStub{},
		metrics.New(),
		"example.com",
		[]byte("secret_key"),
		"",
		1000,
		10000,
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
		&WwwGrantsStub{},
		&WwwRelayStub{},
		&WwwMailRelayStub{},
		&WwwPayPalStub{},
		metrics.New(),
		"example.com",
		[]byte("secret_key"),
		"",
		1000,
		10000,
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
		&WwwGrantsStub{},
		&WwwRelayStub{},
		&WwwMailRelayStub{},
		&WwwPayPalStub{},
		metrics.New(),
		"example.com",
		[]byte("secret_key"),
		"",
		1000,
		10000,
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

func orderWww(orders *WwwOrdersStub) *Www {
	return NewWww(
		&WwwDomainsStub{},
		&WwwNsCheckerStub{},
		&WwwUsersStub{authenticated: true},
		&WwwActionsStub{},
		&WwwMailStub{},
		&WwwStripeStub{},
		orders,
		&WwwGrantsStub{},
		&WwwRelayStub{},
		&WwwMailRelayStub{},
		&WwwPayPalStub{},
		metrics.New(),
		"example.com",
		[]byte("secret_key"),
		"",
		1000,
		10000,
		1000,
		10000,
		log.Default(),
	)
}

func grantWww(grants *WwwGrantsStub) *Www {
	return NewWww(
		&WwwDomainsStub{},
		&WwwNsCheckerStub{},
		&WwwUsersStub{authenticated: true},
		&WwwActionsStub{},
		&WwwMailStub{},
		&WwwStripeStub{},
		&WwwOrdersStub{},
		grants,
		&WwwRelayStub{},
		&WwwMailRelayStub{},
		&WwwPayPalStub{},
		metrics.New(),
		"example.com",
		[]byte("secret_key"),
		"",
		1000,
		10000,
		1000,
		10000,
		log.Default(),
	)
}

func redeemRequest(t *testing.T, code string) *http.Request {
	body, err := json.Marshal(model.PlanRedeemRequest{Code: code})
	if err != nil {
		t.Fatal(err)
	}
	return httptest.NewRequest("POST", "/plan/redeem", strings.NewReader(string(body)))
}

func sessionWww() *Www {
	return NewWww(
		&WwwDomainsStub{},
		&WwwNsCheckerStub{},
		&WwwUsersStub{authenticated: true},
		&WwwActionsStub{},
		&WwwMailStub{},
		&WwwStripeStub{},
		&WwwOrdersStub{},
		&WwwGrantsStub{},
		&WwwRelayStub{},
		&WwwMailRelayStub{},
		&WwwPayPalStub{},
		metrics.New(),
		"example.com",
		[]byte("secret_key"),
		"",
		1000,
		10000,
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

type WwwOrdersStub struct {
	started *product.Order
}

func (s *WwwOrdersStub) Catalog() []product.Device { return nil }

func (s *WwwOrdersStub) Shipping() int { return 1500 }

func (s *WwwOrdersStub) Redact(_ int64) error { return nil }

func (s *WwwOrdersStub) Start(order *product.Order, _ string) (string, error) {
	s.started = order
	return "REF", nil
}

func (s *WwwOrdersStub) Complete(_ int64, _ string) (*product.Order, error) {
	return &product.Order{Id: 1, Email: "buyer@example.com"}, nil
}

func (s *WwwOrdersStub) Mine(_ int64) ([]*product.Order, error) { return nil, nil }

func (s *WwwOrdersStub) All() ([]*product.Order, error) { return nil, nil }

func (s *WwwOrdersStub) SetStatus(_ int64, _ string, _ string) error { return nil }

func (s *WwwOrdersStub) Detail(_ int64, _ int64, _ bool) (*product.Order, []*product.OrderEvent, error) {
	return &product.Order{}, nil, nil
}

func (s *WwwOrdersStub) Unfinished(_ int64) ([]*product.Order, error) { return nil, nil }

func (s *WwwOrdersStub) Retry(_ int64, _ int64, _ string) (*product.Order, error) {
	return &product.Order{}, nil
}

func (s *WwwOrdersStub) Describe(device, option string) (string, string, error) {
	return device, option, nil
}

func shopWww() *Www {
	return NewWww(
		&WwwDomainsStub{},
		&WwwNsCheckerStub{},
		&WwwUsersStub{},
		&WwwActionsStub{},
		&WwwMailStub{},
		&WwwStripeStub{},
		&WwwOrdersStub{},
		&WwwGrantsStub{},
		&WwwRelayStub{},
		&WwwMailRelayStub{},
		&WwwPayPalStub{},
		metrics.New(),
		"example.com",
		[]byte("secret_key"),
		"",
		1000,
		10000,
		1000,
		10000,
		log.Default(),
	)
}

func postEvent(www *Www, body string) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	www.SiteEvent(rr, httptest.NewRequest("POST", "/event", bytes.NewReader([]byte(body))))
	return rr
}

func TestSiteEvent_CountsAKnownStep(t *testing.T) {
	www := shopWww()
	for _, name := range event.Shop() {
		rr := postEvent(www, fmt.Sprintf(`{"event":"%s","language":"de"}`, name))
		assert.Equal(t, http.StatusNoContent, rr.Code, name)
	}
}

func TestSiteEvent_RefusesAnythingNotOnTheList(t *testing.T) {
	www := shopWww()
	for _, name := range []string{"", "shop.anything", "user.deleted", "shop.view "} {
		rr := postEvent(www, fmt.Sprintf(`{"event":"%s"}`, name))
		assert.Equal(t, http.StatusNotFound, rr.Code, name)
	}
}

func TestSiteEvent_RefusesRubbish(t *testing.T) {
	www := shopWww()
	assert.Equal(t, http.StatusBadRequest, postEvent(www, "not json").Code)
}

func TestSiteEvent_AcceptsAnUnknownLanguageWithoutMakingALabelForIt(t *testing.T) {
	www := shopWww()
	rr := postEvent(www, fmt.Sprintf(`{"event":"%s","language":"klingon"}`, event.ShopView))
	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func orderRequest(t *testing.T, email string) *http.Request {
	body := `{"device":"x1","option":"1t","provider":"stripe","email":"` + email + `",` +
		`"name":"Ada","address":"1 Road","city":"London","postcode":"E1","country":"United Kingdom"}`
	req, err := http.NewRequest("POST", "/device/order", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	return req
}

func TestDeviceOrderTakesTheEmailTheGuestTyped(t *testing.T) {
	orders := &WwwOrdersStub{}
	www := orderWww(orders)

	_, err := www.DeviceOrder(httptest.NewRecorder(), orderRequest(t, "Guest@Example.COM"), model.User{})
	if err != nil {
		t.Fatal(err)
	}
	if orders.started == nil {
		t.Fatal("an order with no account was not started")
	}
	if orders.started.UserId != 0 {
		t.Fatalf("a guest order was attached to account %d", orders.started.UserId)
	}
	if orders.started.Email != "guest@example.com" {
		t.Fatalf("the guest email was stored as %q", orders.started.Email)
	}
}

func TestDeviceOrderIgnoresAnEmailSentByAnAccountHolder(t *testing.T) {
	orders := &WwwOrdersStub{}
	www := orderWww(orders)

	_, err := www.DeviceOrder(httptest.NewRecorder(), orderRequest(t, "someone@else.com"),
		model.User{Id: 42, Email: "owner@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if orders.started.Email != "owner@example.com" {
		t.Fatalf("an account order was confirmed to %q", orders.started.Email)
	}
	if orders.started.UserId != 42 {
		t.Fatalf("the order lost its account: %d", orders.started.UserId)
	}
}

func TestDeviceOrderRefusesAGuestEmailItCannotDeliverTo(t *testing.T) {
	for _, typed := range []string{"", "guest", "guest@", "guest@example", "a b@example.com"} {
		orders := &WwwOrdersStub{}
		www := orderWww(orders)

		_, err := www.DeviceOrder(httptest.NewRecorder(), orderRequest(t, typed), model.User{})
		if err == nil {
			t.Fatalf("%q was accepted as an email", typed)
		}
		if orders.started != nil {
			t.Fatalf("%q started an order anyway", typed)
		}
	}
}

func TestPlanRedeemTakesTheCodeFromTheBuyerAndReportsThePeriod(t *testing.T) {
	endsAt := time.Date(2027, 11, 15, 0, 0, 0, 0, time.UTC)
	grants := &WwwGrantsStub{redemption: &grant.Redemption{Id: 3, Plan: "pro", EndsAt: endsAt}}
	www := grantWww(grants)

	response, err := www.PlanRedeem(httptest.NewRecorder(), redeemRequest(t, "aaaa-bbbb-cccc"),
		model.User{Id: 42})
	if err != nil {
		t.Fatal(err)
	}

	if grants.redeemed != "aaaa-bbbb-cccc" {
		t.Fatalf("redeemed %q", grants.redeemed)
	}
	if grants.forUser != 42 {
		t.Fatalf("redeemed for account %d", grants.forUser)
	}
	redeemed, ok := response.(model.PlanRedeemResponse)
	if !ok {
		t.Fatalf("response %T", response)
	}
	if redeemed.Plan != "pro" || !redeemed.EndsAt.Equal(endsAt) {
		t.Fatalf("response %+v", redeemed)
	}
}

func TestPlanRedeemPassesOnWhyACodeWasRefused(t *testing.T) {
	grants := &WwwGrantsStub{err: grant.ErrUsedUp}
	www := grantWww(grants)

	_, err := www.PlanRedeem(httptest.NewRecorder(), redeemRequest(t, "aaaa-bbbb-cccc"),
		model.User{Id: 42})

	refused := &model.ServiceError{}
	if !errors.As(err, &refused) {
		t.Fatalf("error %T %v", err, err)
	}
	if refused.StatusCode != http.StatusBadRequest {
		t.Fatalf("a mistyped code answered %d", refused.StatusCode)
	}
	if refused.Error() != grant.ErrUsedUp.Error() {
		t.Fatalf("message %q", refused.Error())
	}
}
