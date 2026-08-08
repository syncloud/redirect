package inbound

import (
	"fmt"
	"strings"
)

var (
	ErrNoSuchDomain = fmt.Errorf("no such domain here")
	ErrNotAccepted  = fmt.Errorf("mail is not accepted for this domain")
)

type Router struct {
	store DomainStore
}

func NewRouter(store DomainStore) *Router {
	return &Router{store: store}
}

func (r *Router) Route(recipient string) (string, error) {
	name, err := recipientDomain(recipient)
	if err != nil {
		return "", err
	}
	domain, err := r.store.GetDomainByName(name)
	if err != nil {
		return "", err
	}
	if domain == nil {
		return "", ErrNoSuchDomain
	}
	if !domain.Relay {
		return "", ErrNotAccepted
	}
	return domain.Name, nil
}

func recipientDomain(recipient string) (string, error) {
	at := strings.LastIndex(recipient, "@")
	if at < 0 || at == len(recipient)-1 {
		return "", ErrNoSuchDomain
	}
	return strings.ToLower(recipient[at+1:]), nil
}
