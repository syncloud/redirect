package service

import (
	"fmt"
	"github.com/syncloud/redirect/smtp"
	"go.uber.org/zap"
	"log"
	"os"
	"strings"
)

type Mail struct {
	smtp                        *smtp.Smtp
	resetPasswordTemplatePath   string
	setPasswordTemplatePath     string
	activateTemplatePath        string
	planSubscribeTemplatePath   string
	planUnSubscribeTemplatePath string
	releaseAnnouncementPath     string
	dnsCleanPath                string
	subscriptionTrialPath       string
	accountLockSoonPath         string
	accountLockedPath           string
	accountRemovedPath          string
	from                        string
	deviceErrorTo               string
	mainDomain                  string
	logger                      *zap.Logger
}

func NewMail(smtp *smtp.Smtp,
	mailPath string,
	from string,
	deviceErrorTo string,
	mainDomain string,
	logger *zap.Logger,
) *Mail {

	return &Mail{
		smtp:                        smtp,
		resetPasswordTemplatePath:   mailPath + "/reset_password.txt",
		setPasswordTemplatePath:     mailPath + "/set_password.txt",
		activateTemplatePath:        mailPath + "/activate.txt",
		planSubscribeTemplatePath:   mailPath + "/plan_subscribe.txt",
		planUnSubscribeTemplatePath: mailPath + "/plan_unsubscribe.txt",
		releaseAnnouncementPath:     mailPath + "/release_announcement.txt",
		dnsCleanPath:                mailPath + "/dns_clean.txt",
		subscriptionTrialPath:       mailPath + "/subscription_trial.txt",
		accountLockSoonPath:         mailPath + "/account_lock_soon.txt",
		accountLockedPath:           mailPath + "/account_locked.txt",
		accountRemovedPath:          mailPath + "/account_removed.txt",
		from:                        from,
		deviceErrorTo:               deviceErrorTo,
		mainDomain:                  mainDomain,
		logger:                      logger,
	}
}

func (m *Mail) SendResetPassword(to string, token string) error {
	return m.SendNotification(m.resetPasswordTemplatePath, map[string]string{
		"token":  token,
		"domain": m.mainDomain,
	}, to)
}

func (m *Mail) SendSetPassword(to string) error {
	return m.SendNotification(m.setPasswordTemplatePath, map[string]string{}, to)
}

func (m *Mail) SendActivate(to string, token string) error {
	return m.SendNotification(m.activateTemplatePath, map[string]string{
		"token":  token,
		"domain": m.mainDomain,
	}, to, m.deviceErrorTo)
}

func (m *Mail) SendPlanSubscribed(to string) error {
	return m.SendNotification(m.planSubscribeTemplatePath, map[string]string{
		"domain": m.mainDomain,
	}, to, m.deviceErrorTo)
}

func (m *Mail) SendPlanUnSubscribed(to string) error {
	return m.SendNotification(m.planUnSubscribeTemplatePath, map[string]string{
		"domain": m.mainDomain,
	}, to, m.deviceErrorTo)
}

func (m *Mail) SendReleaseAnnouncement(to string) error {
	return m.SendNotification(m.releaseAnnouncementPath, map[string]string{
		"domain": m.mainDomain,
	}, to)
}

func (m *Mail) SendDnsCleanNotification(to string, userDomain string) error {
	return m.SendNotification(m.dnsCleanPath, map[string]string{
		"main_domain": m.mainDomain,
		"user_domain": userDomain,
	}, to, m.deviceErrorTo)
}

func (m *Mail) SendTrial(to string) error {
	return m.SendNotification(m.subscriptionTrialPath, map[string]string{
		"main_domain": m.mainDomain,
	}, to, m.deviceErrorTo)
}

func (m *Mail) SendAccountLockSoon(to string) error {
	return m.SendNotification(m.accountLockSoonPath, map[string]string{
		"main_domain": m.mainDomain,
	}, to, m.deviceErrorTo)
}

func (m *Mail) SendAccountLocked(to string) error {
	return m.SendNotification(m.accountLockedPath, map[string]string{
		"main_domain": m.mainDomain,
	}, to, m.deviceErrorTo)
}

func (m *Mail) SendAccountRemoved(to string) error {
	return m.SendNotification(m.accountRemovedPath, map[string]string{
		"main_domain": m.mainDomain,
	}, to, m.deviceErrorTo)
}

func (m *Mail) SendNotification(template string, subs map[string]string, to ...string) error {
	m.logger.Info("send email notification", zap.String("template", template), zap.Strings("to", to))
	buf, err := os.ReadFile(template)
	if err != nil {
		m.logger.Error("unable to read email template", zap.String("template", template), zap.Error(err))
		return err
	}
	subject, body, err := ParseBody(string(buf), subs)
	if err != nil {
		m.logger.Error("unable to parse email template", zap.String("template", template), zap.Error(err))
		return err
	}
	err = m.smtp.Send(m.from, "text/plain", body, subject, to...)
	if err != nil {
		m.logger.Error("unable to send email", zap.Strings("to", to), zap.Error(err))
		return err
	}
	return nil
}

func (m *Mail) SendLogs(to string, data string, includeSupport bool) error {
	body := "Thank you for sharing Syncloud device error info, Syncloud support will get back to you shortly.\n"
	body += "If you need to add more details just reply to this email.\n\n"
	body += data

	log.Printf("sending logs, include support: %v\n", includeSupport)
	recipients := []string{to}
	if includeSupport {
		recipients = append(recipients, m.deviceErrorTo)
	}
	return m.smtp.Send(m.from, "text/plain", body, "Device error report", recipients...)
}

func ParseBody(template string, substitution map[string]string) (string, string, error) {
	for k, v := range substitution {
		template = strings.ReplaceAll(template, "{"+k+"}", v)
	}
	parts := strings.SplitN(template, "\n", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("cannot parse template")
	}
	subjectLine := parts[0]
	subject := strings.ReplaceAll(subjectLine, "Subject: ", "")
	return subject, parts[1], nil
}
