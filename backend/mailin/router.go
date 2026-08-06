package mailin

import (
	"fmt"
	"strings"
)

var (
	ErrNoSuchDomain = fmt.Errorf("no such domain here")
	ErrNotAccepted  = fmt.Errorf("mail is not accepted for this domain")
)

// Router decides which device a recipient belongs to. Every device is reached
// through the same frps multiplexer port and told apart by name, so there is
// nothing to look up beyond whether the domain wants mail at all.
type Router struct {
	store DomainStore
	muxer string
}

func NewRouter(store DomainStore, muxer string) *Router {
	return &Router{store: store, muxer: muxer}
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
	return &Route{Domain: domain.Name, Muxer: r.muxer}, nil
}

func recipientDomain(recipient string) (string, error) {
	at := strings.LastIndex(recipient, "@")
	if at < 0 || at == len(recipient)-1 {
		return "", ErrNoSuchDomain
	}
	return strings.ToLower(recipient[at+1:]), nil
}
