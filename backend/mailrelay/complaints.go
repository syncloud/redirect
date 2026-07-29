package mailrelay

import (
	"encoding/json"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

type Blocker interface {
	Block(domain string, reason string) error
}

// snsNotification is the envelope SNS wraps SES event publishing in.
type snsNotification struct {
	Type    string `json:"Type"`
	Message string `json:"Message"`
}

type sesEvent struct {
	NotificationType string `json:"notificationType"`
	EventType        string `json:"eventType"`
	Mail             struct {
		Source string `json:"source"`
	} `json:"mail"`
}

type Complaints struct {
	blocker Blocker
	logger  *zap.Logger
}

func NewComplaints(blocker Blocker, logger *zap.Logger) *Complaints {
	return &Complaints{blocker: blocker, logger: logger}
}

// Handle blocks a device the moment SES reports a spam complaint against it, so
// one bad sender cannot burn the reputation the whole relay depends on.
func (c *Complaints) Handle(w http.ResponseWriter, r *http.Request) {
	var notification snsNotification
	if err := json.NewDecoder(r.Body).Decode(&notification); err != nil {
		http.Error(w, "cannot parse notification", http.StatusBadRequest)
		return
	}
	var event sesEvent
	if err := json.Unmarshal([]byte(notification.Message), &event); err != nil {
		http.Error(w, "cannot parse event", http.StatusBadRequest)
		return
	}
	kind := event.NotificationType
	if kind == "" {
		kind = event.EventType
	}
	if !strings.EqualFold(kind, "Complaint") {
		w.WriteHeader(http.StatusOK)
		return
	}
	domain := senderDomain(event.Mail.Source)
	if domain == "" {
		w.WriteHeader(http.StatusOK)
		return
	}
	c.logger.Warn("mail relay complaint, blocking", zap.String("domain", domain))
	if err := c.blocker.Block(domain, "spam complaint"); err != nil {
		c.logger.Error("unable to block", zap.String("domain", domain), zap.Error(err))
		http.Error(w, "cannot block", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func senderDomain(source string) string {
	at := strings.LastIndex(source, "@")
	if at < 0 {
		return ""
	}
	return strings.ToLower(source[at+1:])
}
