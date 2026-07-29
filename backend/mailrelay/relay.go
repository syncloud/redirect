package mailrelay

import (
	"fmt"
	"strings"

	"github.com/syncloud/redirect/model"
	"go.uber.org/zap"
)

type Domains interface {
	GetDomain(token string) (*model.Domain, error)
}

type Plans interface {
	MessageLimit(userId int64) int64
}

type Usage interface {
	Sent(domain string) (int64, error)
	Increment(domain string, count int64) error
}

type Blocklist interface {
	Blocked(domain string) (bool, error)
}

var (
	ErrUnknownToken = fmt.Errorf("unknown login or password")
	ErrNotOwned     = fmt.Errorf("unknown login or password")
	ErrBlocked      = fmt.Errorf("sending is blocked after spam complaints, contact support")
	ErrNotAllowed   = fmt.Errorf("mail relay requires an active subscription")
	ErrOverLimit    = fmt.Errorf("monthly mail relay limit exceeded")
)

type Relay struct {
	domains   Domains
	plans     Plans
	usage     Usage
	blocklist Blocklist
	logger    *zap.Logger
}

func New(domains Domains, plans Plans, usage Usage, blocklist Blocklist, logger *zap.Logger) *Relay {
	return &Relay{domains: domains, plans: plans, usage: usage, blocklist: blocklist, logger: logger}
}

// Authorize checks the credentials a device presents over SMTP AUTH. The login
// is the device domain and the password is its domain update token, the same
// pair the frp traffic relay authenticates with, so nothing extra is issued.
func (r *Relay) Authorize(login string, password string) (*model.Domain, error) {
	domain, err := r.domains.GetDomain(password)
	if err != nil || domain == nil {
		return nil, ErrUnknownToken
	}
	if !strings.EqualFold(login, domain.Name) {
		return nil, ErrNotOwned
	}
	blocked, err := r.blocklist.Blocked(domain.Name)
	if err != nil {
		return nil, err
	}
	if blocked {
		return nil, ErrBlocked
	}
	limit := r.plans.MessageLimit(domain.UserId)
	if limit <= 0 {
		return nil, ErrNotAllowed
	}
	sent, err := r.usage.Sent(domain.Name)
	if err != nil {
		return nil, err
	}
	if sent >= limit {
		return nil, ErrOverLimit
	}
	return domain, nil
}

// Allowed reports whether the sender address belongs to the authenticated
// domain, so a device can only send as its own users.
func (r *Relay) Allowed(domain *model.Domain, from string) bool {
	at := strings.LastIndex(from, "@")
	if at < 0 {
		return false
	}
	return strings.EqualFold(from[at+1:], domain.Name)
}

func (r *Relay) Sent(domain *model.Domain, recipients int) error {
	return r.usage.Increment(domain.Name, int64(recipients))
}
