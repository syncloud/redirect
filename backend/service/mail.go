package service

import (
	"fmt"
	"github.com/syncloud/redirect/smtp"
	"log"
	"os"
	"strings"
)

type Mail struct {
	smtp                      *smtp.Smtp
	resetPasswordTemplatePath string
	setPasswordTemplatePath   string
	activateTemplatePath      string
	planSubscribeTemplatePath string
	releaseAnnouncementPath   string
	dnsCleanPath              string
	from                      string
	deviceErrorTo             string
	mainDomain                string
}

func NewMail(smtp *smtp.Smtp,
	mailPath string,
	from string,
	deviceErrorTo string,
	mainDomain string) *Mail {

	return &Mail{
		smtp:                      smtp,
		resetPasswordTemplatePath: mailPath + "/reset_password.txt",
		setPasswordTemplatePath:   mailPath + "/set_password.txt",
		activateTemplatePath:      mailPath + "/activate.txt",
		planSubscribeTemplatePath: mailPath + "/plan_subscribe.txt",
		releaseAnnouncementPath:   mailPath + "/release_announcement.txt",
		dnsCleanPath:              mailPath + "/dns_clean.txt",
		from:                      from,
		deviceErrorTo:             deviceErrorTo,
		mainDomain:                mainDomain,
	}
}

func (m *Mail) SendResetPassword(to string, token string) error {
	buf, err := os.ReadFile(m.resetPasswordTemplatePath)
	if err != nil {
		return err
	}
	template := string(buf)
	subject, body, err := ParseBody(template, map[string]string{"domain": m.mainDomain, "token": token})
	if err != nil {
		return err
	}
	err = m.smtp.Send(m.from, "text/plain", body, subject, to)
	return err
}

func (m *Mail) SendSetPassword(to string) error {
	buf, err := os.ReadFile(m.setPasswordTemplatePath)
	if err != nil {
		return err
	}
	template := string(buf)
	subject, body, err := ParseBody(template, map[string]string{})
	if err != nil {
		return err
	}
	err = m.smtp.Send(m.from, "text/plain", body, subject, to)
	return err
}

func (m *Mail) SendActivate(to string, token string) error {
	buf, err := os.ReadFile(m.activateTemplatePath)
	if err != nil {
		return err
	}
	template := string(buf)
	subject, body, err := ParseBody(template, map[string]string{"token": token, "domain": m.mainDomain})
	if err != nil {
		return err
	}
	err = m.smtp.Send(m.from, "text/plain", body, subject, to)
	return err
}

func (m *Mail) SendPlanSubscribed(to string) error {
	buf, err := os.ReadFile(m.planSubscribeTemplatePath)
	if err != nil {
		return err
	}
	template := string(buf)
	subject, body, err := ParseBody(template, map[string]string{"domain": m.mainDomain})
	if err != nil {
		return err
	}
	err = m.smtp.Send(m.from, "text/plain", body, subject, to, m.deviceErrorTo)
	return err
}

func (m *Mail) SendReleaseAnnouncement(to string) error {
	buf, err := os.ReadFile(m.releaseAnnouncementPath)
	if err != nil {
		return err
	}
	template := string(buf)
	subject, body, err := ParseBody(template, map[string]string{"domain": m.mainDomain})
	if err != nil {
		return err
	}
	err = m.smtp.Send(m.from, "text/plain", body, subject, to)
	return err
}

func (m *Mail) SendDnsCleanNotification(to string, userDomain string) error {
	buf, err := os.ReadFile(m.dnsCleanPath)
	if err != nil {
		return err
	}
	template := string(buf)
	subject, body, err := ParseBody(template,
		map[string]string{
			"main_domain": m.mainDomain,
			"user_domain": userDomain,
		},
	)
	if err != nil {
		return err
	}
	err = m.smtp.Send(m.from, "text/plain", body, subject, to)
	return err
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
