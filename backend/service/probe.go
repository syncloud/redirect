package service

import (
	"crypto/tls"
	"fmt"
	"github.com/syncloud/redirect/model"
	"io/ioutil"
	"net/http"
	"time"
)

type ProbeResponse struct {
	Message    string `json:"message"`
	DeviceIp   string `json:"device_ip"`
	StatusCode int    `json:"-"`
}

type ProbeDb interface {
	GetDomainByToken(token string) (*model.Domain, error)
	GetUser(id int64) (*model.User, error)
}

type PortProbe struct {
	db ProbeDb
}

func NewPortProbe(db ProbeDb) *PortProbe {
	return &PortProbe{db: db}
}

func (p PortProbe) Probe(token string, port int, protocol string, ip string) (*ProbeResponse, error) {

	domain, err := p.db.GetDomainByToken(token)
	if err != nil {
		return nil, fmt.Errorf("unknown domain update token")
	}

	user, err := p.db.GetUser(domain.UserId)
	if err != nil {
		return nil, fmt.Errorf("unknown user for domain update token: %s", token)
	}

	if domain == nil || user == nil || !user.Active {
		return nil, fmt.Errorf("unknown user for domain update token: %s", token)
	}

	url := fmt.Sprintf("%s://%s:%d/ping", protocol, ip, port)
	client := &http.Client{
		Timeout: time.Second * 1,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	result := &ProbeResponse{DeviceIp: ip, Message: "Port is not reachable"}
	resp, err := client.Get(url)
	if err != nil {
		result.StatusCode = 500
		return result, nil
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		result.StatusCode = 500
		return result, nil
	}
	result.StatusCode = resp.StatusCode
	if resp.StatusCode == 200 {
		result.Message = string(body)
	}
	return result, nil
}
