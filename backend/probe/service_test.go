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
	userExists   bool
}

func (d DbStub) GetDomainByToken(_ string) (*model.Domain, error) {
	if d.domainExists {
		return &model.Domain{}, nil
	}
	return nil, nil
}

func (d DbStub) GetUser(_ int64) (*model.User, error) {
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
	if c.status != 200 {
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
	probe, err := service.Probe("existing", 1, "1.1.1.1")
	assert.Nil(t, err)
	assert.Equal(t, 200, probe.StatusCode)
}

func TestProbe_WrongToken_Fail(t *testing.T) {
	service := New(&DbStub{}, &ClientStub{})
	_, err := service.Probe("existing", 1, "1.1.1.1")
	assert.NotNil(t, err)
	//assert.Equal(t, 200, probe.StatusCode)
}
