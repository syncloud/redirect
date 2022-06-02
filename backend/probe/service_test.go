package probe

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/syncloud/redirect/model"
	"io/ioutil"
	"net/http"
	"testing"
)

type DbStub struct {
	domainExists bool
	domainError  bool
	userExists   bool
	userError    bool
}

func (d DbStub) GetDomainByToken(_ string) (*model.Domain, error) {
	if d.domainError {
		return nil, fmt.Errorf("error")
	}
	if d.domainExists {
		return &model.Domain{}, nil
	}
	return nil, nil
}

func (d DbStub) GetUser(_ int64) (*model.User, error) {
	if d.userError {
		return nil, fmt.Errorf("error")
	}
	if d.userExists {
		return &model.User{Active: true}, nil
	}
	return nil, nil
}

type ClientStub struct {
	response string
	status   int
}

func (c ClientStub) Get(_ string) (resp *http.Response, err error) {
	if c.status == 500 {
		return nil, fmt.Errorf("error code: %v", c.status)
	}

	r := ioutil.NopCloser(bytes.NewReader([]byte(c.response)))
	return &http.Response{
		StatusCode: c.status,
		Body:       r,
	}, nil
}

func TestProbe_Success(t *testing.T) {
	service := New(&DbStub{domainExists: true, userExists: true}, &ClientStub{"OK", 200})
	_, err := service.Probe("existing", 1, "1.1.1.1")
	assert.Nil(t, err)
}

func TestProbe_UnknownDomain_Fail(t *testing.T) {
	service := New(&DbStub{domainExists: false}, &ClientStub{})
	_, err := service.Probe("existing", 1, "1.1.1.1")
	assert.NotNil(t, err)
}

func TestProbe_ErrorDomain_Fail(t *testing.T) {
	service := New(&DbStub{domainError: true}, &ClientStub{})
	_, err := service.Probe("existing", 1, "1.1.1.1")
	assert.NotNil(t, err)
}

func TestProbe_UnknownUser_Fail(t *testing.T) {
	service := New(&DbStub{domainExists: true, userExists: false}, &ClientStub{})
	_, err := service.Probe("existing", 1, "1.1.1.1")
	assert.NotNil(t, err)
}

func TestProbe_ErrorUser_Fail(t *testing.T) {
	service := New(&DbStub{domainExists: true, userError: true}, &ClientStub{})
	_, err := service.Probe("existing", 1, "1.1.1.1")
	assert.NotNil(t, err)
}

func TestProbe_HttpError_Fail(t *testing.T) {
	service := New(&DbStub{domainExists: true, userExists: true}, &ClientStub{status: 500})
	_, err := service.Probe("existing", 1, "1.1.1.1")
	assert.NotNil(t, err)
}

func TestProbe_HttpStatusNonOK_Fail(t *testing.T) {
	service := New(&DbStub{domainExists: true, userExists: true}, &ClientStub{status: 502})
	_, err := service.Probe("existing", 1, "1.1.1.1")
	assert.NotNil(t, err)
}
