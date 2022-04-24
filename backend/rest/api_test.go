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
	domainUpdateRequest model.DomainUpdateRequest
}

func (d *ApiDomainsStub) GetDomains(_ *model.User) ([]*model.Domain, error) {
	return []*model.Domain{}, nil
}

func (d *ApiDomainsStub) DomainAcquire(_ model.DomainAcquireRequest, _ string) (*model.Domain, error) {
	//TODO implement me
	panic("implement me")
}

func (d *ApiDomainsStub) Availability(_ model.DomainAvailabilityRequest) (*model.Domain, error) {
	//TODO implement me
	panic("implement me")
}

func (d *ApiDomainsStub) Update(request model.DomainUpdateRequest, requestIp *string) (*model.Domain, error) {
	d.domainUpdateRequest = request
	return &model.Domain{}, nil
}

func (d *ApiDomainsStub) GetDomain(_ string) (*model.Domain, error) {
	//TODO implement me
	panic("implement me")
}

type ApiUsersStub struct {
	email    *string
	password *string
}

func (a *ApiUsersStub) CreateNewUser(_ model.UserCreateRequest) (*model.User, error) {
	//TODO implement me
	panic("implement me")
}

func (a *ApiUsersStub) Authenticate(email *string, password *string) (*model.User, error) {
	a.email = email
	a.password = password
	return &model.User{Email: *email}, nil
}

func (a *ApiUsersStub) GetUserByUpdateToken(_ string) (*model.User, error) {
	//TODO implement me
	panic("implement me")
}

type ApiMailStub struct {
}

func (a *ApiMailStub) SendLogs(_ string, _ string, _ bool) error {
	//TODO implement me
	panic("implement me")
}

type ApiPortProbeStub struct {
}

func (a *ApiPortProbeStub) Probe(_ string, _ int, _ string) (*service.ProbeResponse, error) {
	//TODO implement me
	panic("implement me")
}

type ApiCertbotStub struct {
}

func (a ApiCertbotStub) Present(_ string, _ string, _ []string) error {
	//TODO implement me
	panic("implement me")
}

func (a ApiCertbotStub) CleanUp(_ string, _ string) error {
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

func TestDomainUpdate_Ipv4Enabled(t *testing.T) {

	domains := &ApiDomainsStub{}
	api := NewApi(
		&StatsdClientStub{}, domains, &ApiUsersStub{}, &ApiMailStub{},
		&ApiPortProbeStub{}, &ApiCertbotStub{}, "example.com")
	request := `
{ "port": 1 }
`
	body := bytes.NewBufferString(request)
	req, err := http.NewRequest("GET", "/", body)
	if err != nil {
		t.Fatal(err)
	}
	req.RemoteAddr = "127.0.0.1:0"
	rr := httptest.NewRecorder()
	_, err = api.DomainUpdate(rr, req)
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, domains.domainUpdateRequest.Ipv4Enabled)

}

func TestDomainUpdate_Ipv4Disabled(t *testing.T) {

	domains := &ApiDomainsStub{}
	api := NewApi(
		&StatsdClientStub{}, domains, &ApiUsersStub{}, &ApiMailStub{},
		&ApiPortProbeStub{}, &ApiCertbotStub{}, "example.com")
	request := `
{ "ipv4_enabled": false}
`
	body := bytes.NewBufferString(request)
	req, err := http.NewRequest("GET", "/", body)
	if err != nil {
		t.Fatal(err)
	}
	req.RemoteAddr = "127.0.0.1:0"
	rr := httptest.NewRecorder()
	_, err = api.DomainUpdate(rr, req)
	if err != nil {
		t.Fatal(err)
	}
	assert.False(t, domains.domainUpdateRequest.Ipv4Enabled)

}
