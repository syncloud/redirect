package mailrelay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

var ErrRejectedAsSpam = fmt.Errorf("message rejected by spam filter")

type Scanner interface {
	Scan(from string, recipients []string, domain string, message []byte) error
}

type rspamdResult struct {
	Action string  `json:"action"`
	Score  float64 `json:"score"`
}

// Rspamd scans outbound mail before it reaches SES, so a compromised device is
// stopped here rather than by AWS pausing the whole account.
type Rspamd struct {
	url    string
	client *http.Client
	failed error
	logger *zap.Logger
}

// NewRspamd fails closed when rejectOnError is set, so a scanner outage stops
// mail rather than silently letting unscanned mail through to SES.
func NewRspamd(url string, timeout time.Duration, rejectOnError bool, logger *zap.Logger) *Rspamd {
	var failed error
	if rejectOnError {
		failed = fmt.Errorf("spam filter unavailable")
	}
	return &Rspamd{
		url:    url,
		client: &http.Client{Timeout: timeout},
		failed: failed,
		logger: logger,
	}
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
	return r.failed
}
