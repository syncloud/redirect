package mailrelay

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/emersion/go-smtp"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func rspamdReturning(t *testing.T, status int, body string) (*httptest.Server, *http.Request) {
	t.Helper()
	var received *http.Request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	return server, received
}

func scan(t *testing.T, status int, body string, rejectOnError bool) error {
	t.Helper()
	server, _ := rspamdReturning(t, status, body)
	scanner := NewRspamd(server.URL, time.Second, rejectOnError, zap.NewNop())
	return scanner.Scan("user@device.syncloud.it", []string{"to@example.com"}, "device.syncloud.it", []byte("message"))
}

func TestRspamd_AllowsClean(t *testing.T) {
	assert.NoError(t, scan(t, http.StatusOK, `{"action":"no action","score":0.1}`, true))
}

func TestRspamd_RejectsSpam(t *testing.T) {
	assert.ErrorIs(t, scan(t, http.StatusOK, `{"action":"reject","score":15.0}`, true), ErrRejectedAsSpam)
}

func TestRspamd_FailsClosedWhenUnavailable(t *testing.T) {
	assert.Error(t, scan(t, http.StatusInternalServerError, "", true))
}

func TestRspamd_FailsOpenWhenConfigured(t *testing.T) {
	assert.NoError(t, scan(t, http.StatusInternalServerError, "", false))
}

func TestScanRejectionIsPermanent(t *testing.T) {
	err := permanent(ErrRejectedAsSpam, smtp.EnhancedCode{5, 7, 1})
	var smtpErr *smtp.SMTPError
	assert.True(t, errors.As(err, &smtpErr))
	assert.Equal(t, 550, smtpErr.Code)
}

func TestInfrastructureFailureIsTemporary(t *testing.T) {
	err := tryAgain(errors.New("ses is throttling"), smtp.EnhancedCode{4, 4, 0})
	var smtpErr *smtp.SMTPError
	assert.True(t, errors.As(err, &smtpErr))
	assert.Equal(t, 451, smtpErr.Code)
}
