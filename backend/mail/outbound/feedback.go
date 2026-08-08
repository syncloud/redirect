package outbound

import (
	"encoding/json"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

type Blocker interface {
	Block(domain string, reason string) error
}

type Bounces interface {
	Sent(domain string) (int64, error)
	Bounced(domain string) (int64, error)
	Bounce(domain string, count int64) error
}

type Feedback struct {
	blocker       Blocker
	bounces       Bounces
	bounceRatio   float64
	bounceMinimum int64
	logger        *zap.Logger
}

func NewFeedback(blocker Blocker, bounces Bounces, bounceRatio float64, bounceMinimum int64, logger *zap.Logger) *Feedback {
	return &Feedback{
		blocker:       blocker,
		bounces:       bounces,
		bounceRatio:   bounceRatio,
		bounceMinimum: bounceMinimum,
		logger:        logger,
	}
}

func (f *Feedback) Handle(w http.ResponseWriter, r *http.Request) {
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
	domain := senderDomain(event.Mail.Source)
	if domain == "" {
		w.WriteHeader(http.StatusOK)
		return
	}
	kind := event.NotificationType
	if kind == "" {
		kind = event.EventType
	}

	var err error
	switch {
	case strings.EqualFold(kind, "Complaint"):
		err = f.complaint(domain)
	case strings.EqualFold(kind, "Bounce"):
		err = f.bounce(domain, event)
	}
	if err != nil {
		f.logger.Error("unable to handle feedback", zap.String("domain", domain), zap.Error(err))
		http.Error(w, "cannot handle feedback", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (f *Feedback) complaint(domain string) error {
	f.logger.Warn("mail relay complaint, blocking", zap.String("domain", domain))
	return f.blocker.Block(domain, "spam complaint")
}

func (f *Feedback) bounce(domain string, event sesEvent) error {
	if event.Bounce == nil || !strings.EqualFold(event.Bounce.BounceType, "Permanent") {
		return nil
	}
	count := int64(len(event.Bounce.BouncedRecipients))
	if count == 0 {
		count = 1
	}
	if err := f.bounces.Bounce(domain, count); err != nil {
		return err
	}
	bounced, err := f.bounces.Bounced(domain)
	if err != nil {
		return err
	}
	sent, err := f.bounces.Sent(domain)
	if err != nil {
		return err
	}
	if sent < f.bounceMinimum {
		return nil
	}
	ratio := float64(bounced) / float64(sent)
	if ratio < f.bounceRatio {
		return nil
	}
	f.logger.Warn("mail relay bounce rate too high, blocking",
		zap.String("domain", domain), zap.Float64("ratio", ratio))
	return f.blocker.Block(domain, "bounce rate too high")
}

func senderDomain(source string) string {
	at := strings.LastIndex(source, "@")
	if at < 0 {
		return ""
	}
	return strings.ToLower(source[at+1:])
}
