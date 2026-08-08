package outbound

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

var (
	ErrRejectedAsSpam     = fmt.Errorf("message rejected by spam filter")
	ErrScannerUnavailable = fmt.Errorf("spam filter unavailable")
)

type Scanner interface {
	Scan(from string, recipients []string, domain string, message []byte) error
}

type RspamdConfig interface {
	GetMailOutboundRspamdUrl() string
}

type Rspamd struct {
	config RspamdConfig
	url    string
	client *http.Client
	logger *zap.Logger
}

func NewRspamd(config RspamdConfig, timeout time.Duration, logger *zap.Logger) *Rspamd {
	return &Rspamd{
		config: config,
		client: &http.Client{Timeout: timeout},
		logger: logger,
	}
}

func (r *Rspamd) Start() error {
	r.url = r.config.GetMailOutboundRspamdUrl()
	if r.url == "" {
		return fmt.Errorf("mail relay spam filter url is not configured")
	}
	return nil
}

func (r *Rspamd) Scan(from string, recipients []string, domain string, message []byte) error {
	request, err := http.NewRequest(http.MethodPost, r.url+"/checkv2", bytes.NewReader(message))
	if err != nil {
		return r.unavailable(err)
	}
	request.Header.Set("From", from)
	request.Header.Set("User", domain)
	for _, recipient := range recipients {
		request.Header.Add("Rcpt", recipient)
	}
	response, err := r.client.Do(request)
	if err != nil {
		return r.unavailable(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return r.unavailable(fmt.Errorf("spam filter returned %s", response.Status))
	}
	var result rspamdResult
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return r.unavailable(err)
	}
	if result.Action == "reject" {
		r.logger.Warn("outbound message rejected as spam",
			zap.String("domain", domain), zap.Float64("score", result.Score))
		return ErrRejectedAsSpam
	}
	return nil
}

func (r *Rspamd) unavailable(err error) error {
	r.logger.Error("spam filter failed", zap.Error(err))
	return ErrScannerUnavailable
}
