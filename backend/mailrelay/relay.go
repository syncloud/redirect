package mailrelay

import (
	"fmt"
	"strings"
	"sync"
	"time"

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
	SentByUser(userId int64) (int64, error)
	Increment(domain string, count int64) error
}

type Blocklist interface {
	Blocked(domain string) (bool, error)
}

type Warner interface {
	Warn(userId int64, used int64, limit int64) error
}

var (
	ErrUnknownToken = fmt.Errorf("unknown login or password")
	ErrNotOwned     = fmt.Errorf("unknown login or password")
	ErrBlocked      = fmt.Errorf("sending is blocked after spam complaints, contact support")
	ErrNotAllowed   = fmt.Errorf("mail relay is not available on this plan")
	ErrOverLimit    = fmt.Errorf("monthly mail relay limit exceeded, upgrade for more")
)

type Relay struct {
	domains   Domains
	plans     Plans
	usage     Usage
	blocklist Blocklist
	warner    Warner
	logger    *zap.Logger

	mutex  sync.Mutex
	month  string
	warned map[int64]bool
}

func New(domains Domains, plans Plans, usage Usage, blocklist Blocklist, warner Warner, logger *zap.Logger) *Relay {
	return &Relay{
		domains: domains, plans: plans, usage: usage, blocklist: blocklist,
		warner: warner, logger: logger, warned: map[int64]bool{},
	}
}

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
	sent, err := r.usage.SentByUser(domain.UserId)
	if err != nil {
		return nil, err
	}
	if sent >= limit {
		return nil, ErrOverLimit
	}
	return domain, nil
}

func (r *Relay) Allowed(domain *model.Domain, from string) bool {
	at := strings.LastIndex(from, "@")
	if at < 0 {
		return false
	}
	return strings.EqualFold(from[at+1:], domain.Name)
}

func (r *Relay) Sent(domain *model.Domain, recipients int) error {
	if err := r.usage.Increment(domain.Name, int64(recipients)); err != nil {
		return err
	}
	return r.warn(domain)
}

func (r *Relay) warn(domain *model.Domain) error {
	if r.warner == nil {
		return nil
	}
	limit := r.plans.MessageLimit(domain.UserId)
	if limit <= 0 {
		return nil
	}
	sent, err := r.usage.SentByUser(domain.UserId)
	if err != nil {
		return err
	}
	if sent*100 < limit*80 {
		return nil
	}
	if !r.shouldWarn(domain.UserId) {
		return nil
	}
	return r.warner.Warn(domain.UserId, sent, limit)
}

func (r *Relay) shouldWarn(userId int64) bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	current := time.Now().UTC().Format("2006-01")
	if current != r.month {
		r.month = current
		r.warned = map[int64]bool{}
	}
	if r.warned[userId] {
		return false
	}
	r.warned[userId] = true
	return true
}
