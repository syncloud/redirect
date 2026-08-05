package mailin

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/syncloud/redirect/model"
)

const DeviceHost = "127.0.0.1"

var (
	ErrNoSuchDomain  = fmt.Errorf("no such domain here")
	ErrNotAccepted   = fmt.Errorf("mail is not accepted for this domain")
	ErrNoDeviceRoute = fmt.Errorf("device has no inbound route")
)

type DomainStore interface {
	GetDomainByName(name string) (*model.Domain, error)
}

type Router struct {
	store DomainStore
	host  string
}

func NewRouter(store DomainStore, host string) *Router {
	return &Router{store: store, host: host}
}

type Route struct {
	Domain  string
	Address string
}

func (r *Router) Route(recipient string) (*Route, error) {
	name, err := recipientDomain(recipient)
	if err != nil {
		return nil, err
	}
	domain, err := r.store.GetDomainByName(name)
	if err != nil {
		return nil, err
	}
	if domain == nil {
		return nil, ErrNoSuchDomain
	}
	if !domain.MailRelay {
		return nil, ErrNotAccepted
	}
	if domain.SmtpPort == nil {
		return nil, ErrNoDeviceRoute
	}
	return &Route{
		Domain:  domain.Name,
		Address: net.JoinHostPort(r.host, strconv.Itoa(*domain.SmtpPort)),
	}, nil
}

func recipientDomain(recipient string) (string, error) {
	at := strings.LastIndex(recipient, "@")
	if at < 0 || at == len(recipient)-1 {
		return "", ErrNoSuchDomain
	}
	return strings.ToLower(recipient[at+1:]), nil
}
