package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/syncloud/redirect/model"
	"github.com/syncloud/redirect/service"
	"net/http"
	"net/http/httptest"
	"testing"
)

type ApiDomainsStub struct {
}

func (w *ApiDomainsStub) GetDomains(user *model.User) ([]*model.Domain, error) {
	return []*model.Domain{}, nil
}

func (w *ApiDomainsStub) DomainAcquire(request model.DomainAcquireRequest, domainField string) (*model.Domain, error) {
	//TODO implement me
	panic("implement me")
}

func (w *ApiDomainsStub) Availability(request model.DomainAvailabilityRequest) (*model.Domain, error) {
	//TODO implement me
	panic("implement me")
}

func (w *ApiDomainsStub) Update(request model.DomainUpdateRequest, requestIp *string) (*model.Domain, error) {
	//TODO implement me
	panic("implement me")
}

func (w *ApiDomainsStub) GetDomain(token string) (*model.Domain, error) {
	//TODO implement me
	panic("implement me")
}

type ApiUsersStub struct {
	email    *string
	password *string
}

func (a *ApiUsersStub) CreateNewUser(request model.UserCreateRequest) (*model.User, error) {
	//TODO implement me
	panic("implement me")
}

func (a *ApiUsersStub) Authenticate(email *string, password *string) (*model.User, error) {
	a.email = email
	a.password = password
	return &model.User{Email: *email}, nil
}

func (a *ApiUsersStub) GetUserByUpdateToken(updateToken string) (*model.User, error) {
	//TODO implement me
	panic("implement me")
}

type ApiMailStub struct {
}

func (a *ApiMailStub) SendLogs(to string, data string, includeSupport bool) error {
	//TODO implement me
	panic("implement me")
}

type ApiPortProbeStub struct {
}

func (a *ApiPortProbeStub) Probe(token string, port int, protocol string, ip string) (*service.ProbeResponse, error) {
	//TODO implement me
	panic("implement me")
}

type ApiCertbotStub struct {
}

func (a ApiCertbotStub) Present(token string, fqdn string, value string) error {
	//TODO implement me
	panic("implement me")
}

func (a ApiCertbotStub) CleanUp(token string, fqdn string, value string) error {
	//TODO implement me
	panic("implement me")
}

func TestParameterErrorToError(t *testing.T) {
	err := &model.ParameterError{ParameterErrors: &[]model.ParameterMessages{{
		Parameter: "param", Messages: []string{"error"},
	}}}
	response, code := ErrorToResponse(err)
	assert.Equal(t, 400, code)
	assert.Equal(t, "param", (*response.ParametersMessages)[0].Parameter)
}

func TestServiceErrorToError(t *testing.T) {
	err := model.NewServiceError("error")
	response, code := ErrorToResponse(err)
	assert.Equal(t, 400, code)
	assert.Equal(t, "error", response.Message)
}

func TestErrorToError(t *testing.T) {
	err := fmt.Errorf("error")
	response, code := ErrorToResponse(err)
	assert.Equal(t, 500, code)
	assert.Equal(t, "error", response.Message)
}

func TestLogin_SpecialSymbol(t *testing.T) {

	users := &ApiUsersStub{}
	api := NewApi(
		&StatsdClientStub{}, &ApiDomainsStub{}, users, &ApiMailStub{},
		&ApiPortProbeStub{}, &ApiCertbotStub{}, "example.com")
	email := "test@example.com"
	password := "password;&\" "
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
	rr := httptest.NewRecorder()
	_, err = api.User(rr, req)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, email, *users.email)
	assert.Equal(t, password, *users.password)

}
