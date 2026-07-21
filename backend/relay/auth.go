package relay

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/syncloud/redirect/model"
	"go.uber.org/zap"
)

type Domains interface {
	GetDomain(token string) (*model.Domain, error)
}

type Limiter interface {
	OverLimit(name string) bool
}

type AuthServer struct {
	address string
	domains Domains
	limiter Limiter
	suffix  string
	logger  *zap.Logger
}

func NewAuthServer(address string, domains Domains, limiter Limiter, suffix string, logger *zap.Logger) *AuthServer {
	return &AuthServer{address: address, domains: domains, limiter: limiter, suffix: suffix, logger: logger}
}

type pluginRequest struct {
	Version string          `json:"version"`
	Op      string          `json:"op"`
	Content json.RawMessage `json:"content"`
}

type pluginUser struct {
	User  string            `json:"user"`
	Metas map[string]string `json:"metas"`
	RunId string            `json:"run_id"`
}

type newProxyContent struct {
	User          pluginUser        `json:"user"`
	ProxyName     string            `json:"proxy_name"`
	ProxyType     string            `json:"proxy_type"`
	Subdomain     string            `json:"subdomain"`
	CustomDomains []string          `json:"custom_domains"`
	Metas         map[string]string `json:"metas"`
}

type newUserConnContent struct {
	User      pluginUser `json:"user"`
	ProxyName string     `json:"proxy_name"`
	ProxyType string     `json:"proxy_type"`
}

type pluginResponse struct {
	Reject       bool   `json:"reject"`
	RejectReason string `json:"reject_reason,omitempty"`
	Unchange     bool   `json:"unchange"`
}

func (s *AuthServer) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/plugin", s.handle)
	srv := &http.Server{Addr: s.address, Handler: mux}
	s.logger.Info("relay auth plugin listening", zap.String("address", s.address))
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("relay auth plugin stopped", zap.Error(err))
		}
	}()
	return nil
}

func allow() pluginResponse { return pluginResponse{Reject: false, Unchange: true} }
func deny(r string) pluginResponse {
	return pluginResponse{Reject: true, RejectReason: r}
}

func (s *AuthServer) handle(w http.ResponseWriter, r *http.Request) {
	var req pluginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.write(w, deny("invalid plugin request"))
		return
	}
	switch req.Op {
	case "NewProxy":
		var content newProxyContent
		if err := json.Unmarshal(req.Content, &content); err != nil {
			s.write(w, deny("invalid new proxy content"))
			return
		}
		s.write(w, s.Authorize(content))
	case "NewUserConn":
		var content newUserConnContent
		if err := json.Unmarshal(req.Content, &content); err != nil {
			s.write(w, deny("invalid new user conn content"))
			return
		}
		s.write(w, s.Enforce(content))
	default:
		s.write(w, allow())
	}
}

func (s *AuthServer) Enforce(content newUserConnContent) pluginResponse {
	if s.limiter != nil && s.limiter.OverLimit(content.ProxyName) {
		s.logger.Warn("relay connection blocked, over monthly limit", zap.String("proxy", content.ProxyName))
		return deny("monthly traffic limit exceeded")
	}
	return allow()
}

func (s *AuthServer) Authorize(content newProxyContent) pluginResponse {
	token := content.User.Metas["token"]
	if token == "" {
		token = content.Metas["token"]
	}
	if token == "" {
		return deny("missing token")
	}
	domain, err := s.domains.GetDomain(token)
	if err != nil || domain == nil {
		s.logger.Warn("relay authorize rejected", zap.String("proxy", content.ProxyName), zap.Error(err))
		return deny("unknown token")
	}
	for _, requested := range s.requestedNames(content) {
		if strings.EqualFold(requested, domain.Name) {
			return allow()
		}
	}
	s.logger.Warn("relay authorize domain mismatch",
		zap.String("owned", domain.Name),
		zap.Strings("custom_domains", content.CustomDomains),
		zap.String("subdomain", content.Subdomain))
	return deny("domain not owned by token")
}

func (s *AuthServer) requestedNames(content newProxyContent) []string {
	names := append([]string{}, content.CustomDomains...)
	if content.Subdomain != "" && s.suffix != "" {
		names = append(names, fmt.Sprintf("%s.%s", content.Subdomain, s.suffix))
	}
	return names
}

func (s *AuthServer) write(w http.ResponseWriter, resp pluginResponse) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
