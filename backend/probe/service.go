package probe

import (
	"fmt"
	"github.com/syncloud/redirect/model"
	"io/ioutil"
)

type Response struct {
	Message    string `json:"message"`
	DeviceIp   string `json:"device_ip"`
	StatusCode int    `json:"-"`
}

type Db interface {
	GetDomainByToken(token string) (*model.Domain, error)
	GetUser(id int64) (*model.User, error)
}

type Service struct {
	db     Db
	client HttpClient
}

func New(db Db, client HttpClient) *Service {
	return &Service{db: db, client: client}
}

func (p Service) Probe(token string, port int, ip string) (*Response, error) {

	domain, err := p.db.GetDomainByToken(token)
	if err != nil || domain == nil {
		return nil, fmt.Errorf("unknown domain update token")
	}

	user, err := p.db.GetUser(domain.UserId)
	if err != nil {
		return nil, fmt.Errorf("unknown user for domain update token: %s", token)
	}

	if user == nil || !user.Active {
		return nil, fmt.Errorf("unknown user for domain update token: %s", token)
	}

	url := fmt.Sprintf("https://%s:%d/ping", ip, port)
	result := &Response{DeviceIp: ip, Message: "Port is not reachable"}
	resp, err := p.client.Get(url)
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
