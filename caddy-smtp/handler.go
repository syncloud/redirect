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
	ProxyProtocol      bool                        `json:"proxy_protocol,omitempty"`
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
		h.ConnectionPolicies = append(h.ConnectionPolicies,
			&caddytls.ConnectionPolicy{DefaultSNI: h.DefaultSni})
	}
	if err := h.ConnectionPolicies.Provision(ctx); err != nil {
		return fmt.Errorf("smtp_starttls: setting up connection policies: %v", err)
	}
	h.conversation = NewConversation(h.Hostname, NewTcpUpstream(h.Upstream, h.ProxyProtocol),
		func() *tls.Config { return h.ConnectionPolicies.TLSConfig(h.ctx) }, h.logger)
	return nil
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
		case "proxy_protocol":
			h.ProxyProtocol = true
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
