package caddysmtp

import (
	"crypto/tls"
	"fmt"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddytls"
	"github.com/mholt/caddy-l4/layer4"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(&Handler{})
}

type Handler struct {
	Hostname           string                      `json:"hostname,omitempty"`
	Upstream           string                      `json:"upstream,omitempty"`
	DefaultSni         string                      `json:"default_sni,omitempty"`
	ConnectionPolicies caddytls.ConnectionPolicies `json:"connection_policies,omitempty"`

	conversation *Conversation
	ctx          caddy.Context
	logger       *zap.Logger
}

func (*Handler) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "layer4.handlers.smtp_starttls",
		New: func() caddy.Module { return new(Handler) },
	}
}

func (h *Handler) Provision(ctx caddy.Context) error {
	h.ctx = ctx
	h.logger = ctx.Logger(h)
	if h.Hostname == "" {
		return fmt.Errorf("smtp_starttls: hostname is required")
	}
	if h.Upstream == "" {
		return fmt.Errorf("smtp_starttls: upstream is required")
	}
	if len(h.ConnectionPolicies) == 0 {
		h.ConnectionPolicies = append(h.ConnectionPolicies, new(caddytls.ConnectionPolicy))
	}
	if err := h.ConnectionPolicies.Provision(ctx); err != nil {
		return fmt.Errorf("smtp_starttls: setting up connection policies: %v", err)
	}
	h.conversation = NewConversation(h.Hostname, NewTcpUpstream(h.Upstream),
		h.tlsConfig, h.logger)
	return nil
}

// a name is only widened to a wildcard when the client sent one, so an unnamed
// connection is given the default. the hello is rebuilt between choosing the
// config and choosing the certificate, so the name has to be filled in on the
// second one to survive
func (h *Handler) tlsConfig() *tls.Config {
	config := h.ConnectionPolicies.TLSConfig(h.ctx)
	chooseConfig := config.GetConfigForClient
	config.GetConfigForClient = func(hello *tls.ClientHelloInfo) (*tls.Config, error) {
		chosen, err := chooseConfig(hello)
		if err != nil || hello.ServerName != "" || h.DefaultSni == "" {
			return chosen, err
		}
		named := chosen.Clone()
		chooseCertificate := chosen.GetCertificate
		named.GetCertificate = func(fresh *tls.ClientHelloInfo) (*tls.Certificate, error) {
			if fresh.ServerName == "" {
				fresh.ServerName = h.DefaultSni
			}
			return chooseCertificate(fresh)
		}
		return named, nil
	}
	return config
}

func (h *Handler) Handle(cx *layer4.Connection, _ layer4.Handler) error {
	return h.conversation.Serve(cx)
}

func (h *Handler) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	_, wrapper := d.Next(), d.Val()
	if d.CountRemainingArgs() > 0 {
		return d.ArgErr()
	}
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		option := d.Val()
		switch option {
		case "hostname":
			if !d.NextArg() {
				return d.ArgErr()
			}
			h.Hostname = d.Val()
		case "upstream":
			if !d.NextArg() {
				return d.ArgErr()
			}
			h.Upstream = d.Val()
		case "default_sni":
			if !d.NextArg() {
				return d.ArgErr()
			}
			h.DefaultSni = d.Val()
		default:
			return d.Errf("unrecognized %s option '%s'", wrapper, option)
		}
		if d.CountRemainingArgs() > 0 {
			return d.ArgErr()
		}
	}
	return nil
}

var (
	_ caddy.Provisioner     = (*Handler)(nil)
	_ caddyfile.Unmarshaler = (*Handler)(nil)
	_ layer4.NextHandler    = (*Handler)(nil)
)
