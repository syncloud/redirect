package relay

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/syncloud/redirect/model"
	"go.uber.org/zap"
)

type fakeDomains struct {
	byToken map[string]*model.Domain
}

func (f *fakeDomains) GetDomain(token string) (*model.Domain, error) {
	domain, ok := f.byToken[token]
	if !ok {
		return nil, fmt.Errorf("unknown domain update token")
	}
	return domain, nil
}

type fakeLimiter struct {
	over map[string]bool
}

func (f *fakeLimiter) OverLimit(name string) bool {
	return f.over[name]
}

func newServer(byToken map[string]*model.Domain) *AuthServer {
	return NewAuthServer("127.0.0.1:0", &fakeDomains{byToken: byToken}, &fakeLimiter{over: map[string]bool{}}, "syncloud.it", zap.NewNop())
}

func domainNamed(name string) *model.Domain {
	return &model.Domain{Name: name}
}

func TestAuthorize_ValidTokenOwnsCustomDomain(t *testing.T) {
	s := newServer(map[string]*model.Domain{"good": domainNamed("alice.syncloud.it")})
	resp := s.Authorize(newProxyContent{
		User:          pluginUser{Metas: map[string]string{"token": "good"}},
		CustomDomains: []string{"alice.syncloud.it"},
	})
	assert.False(t, resp.Reject, resp.RejectReason)
}

func TestAuthorize_ValidTokenOwnsSubdomain(t *testing.T) {
	s := newServer(map[string]*model.Domain{"good": domainNamed("alice.syncloud.it")})
	resp := s.Authorize(newProxyContent{
		User:      pluginUser{Metas: map[string]string{"token": "good"}},
		Subdomain: "alice",
	})
	assert.False(t, resp.Reject, resp.RejectReason)
}

func TestAuthorize_MissingToken(t *testing.T) {
	s := newServer(map[string]*model.Domain{"good": domainNamed("alice.syncloud.it")})
	resp := s.Authorize(newProxyContent{CustomDomains: []string{"alice.syncloud.it"}})
	assert.True(t, resp.Reject)
}

func TestAuthorize_UnknownToken(t *testing.T) {
	s := newServer(map[string]*model.Domain{"good": domainNamed("alice.syncloud.it")})
	resp := s.Authorize(newProxyContent{
		User:          pluginUser{Metas: map[string]string{"token": "bad"}},
		CustomDomains: []string{"alice.syncloud.it"},
	})
	assert.True(t, resp.Reject)
}

func TestAuthorize_TokenDoesNotOwnRequestedDomain(t *testing.T) {
	s := newServer(map[string]*model.Domain{"good": domainNamed("alice.syncloud.it")})
	resp := s.Authorize(newProxyContent{
		User:          pluginUser{Metas: map[string]string{"token": "good"}},
		CustomDomains: []string{"bob.syncloud.it"},
	})
	assert.True(t, resp.Reject)
}

func TestEnforce_UnderLimitAllows(t *testing.T) {
	s := NewAuthServer("127.0.0.1:0", &fakeDomains{}, &fakeLimiter{over: map[string]bool{}}, "syncloud.it", zap.NewNop())
	assert.False(t, s.Enforce(newUserConnContent{ProxyName: "alice.syncloud.it"}).Reject)
}

func TestEnforce_OverLimitRejects(t *testing.T) {
	s := NewAuthServer("127.0.0.1:0", &fakeDomains{}, &fakeLimiter{over: map[string]bool{"alice.syncloud.it": true}}, "syncloud.it", zap.NewNop())
	assert.True(t, s.Enforce(newUserConnContent{ProxyName: "alice.syncloud.it"}).Reject)
}

func mailRelayDomain(name string, port int) *model.Domain {
	return &model.Domain{Name: name, MailRelay: true, SmtpPort: &port}
}

func TestAuthorize_SmtpAssignedPort(t *testing.T) {
	s := newServer(map[string]*model.Domain{"good": mailRelayDomain("alice.syncloud.it", 20000)})
	resp := s.Authorize(newProxyContent{
		User:       pluginUser{Metas: map[string]string{"token": "good"}},
		ProxyType:  proxyTypeTcp,
		RemotePort: 20000,
	})
	assert.False(t, resp.Reject, resp.RejectReason)
}

func TestAuthorize_SmtpOtherDevicePort(t *testing.T) {
	s := newServer(map[string]*model.Domain{"good": mailRelayDomain("alice.syncloud.it", 20000)})
	resp := s.Authorize(newProxyContent{
		User:       pluginUser{Metas: map[string]string{"token": "good"}},
		ProxyType:  proxyTypeTcp,
		RemotePort: 20001,
	})
	assert.True(t, resp.Reject)
}

func TestAuthorize_SmtpNoPortAssigned(t *testing.T) {
	s := newServer(map[string]*model.Domain{"good": &model.Domain{Name: "alice.syncloud.it", MailRelay: true}})
	resp := s.Authorize(newProxyContent{
		User:       pluginUser{Metas: map[string]string{"token": "good"}},
		ProxyType:  proxyTypeTcp,
		RemotePort: 20000,
	})
	assert.True(t, resp.Reject)
}

func TestAuthorize_SmtpMailRelayOff(t *testing.T) {
	port := 20000
	s := newServer(map[string]*model.Domain{
		"good": {Name: "alice.syncloud.it", MailRelay: false, SmtpPort: &port}})
	resp := s.Authorize(newProxyContent{
		User:       pluginUser{Metas: map[string]string{"token": "good"}},
		ProxyType:  proxyTypeTcp,
		RemotePort: 20000,
	})
	assert.True(t, resp.Reject)
}

func TestAuthorize_SmtpUnknownToken(t *testing.T) {
	s := newServer(map[string]*model.Domain{"good": mailRelayDomain("alice.syncloud.it", 20000)})
	resp := s.Authorize(newProxyContent{
		User:       pluginUser{Metas: map[string]string{"token": "bad"}},
		ProxyType:  proxyTypeTcp,
		RemotePort: 20000,
	})
	assert.True(t, resp.Reject)
}
