package service

import (
	"fmt"
	"github.com/syncloud/redirect/smtp"
	"io/ioutil"
	"log"
	"strings"
)

type Mail struct {
	smtp                       *smtp.Smtp
	resetPasswordTemplatePath  string
	activateTemplatePath       string
	premiumRequestTemplatePath string
	from                       string
	passwordUrlTemplate        string
	activateUrlTemplate        string
	deviceErrorTo              string
	mainDomain                 string
}

func NewMail(smtp *smtp.Smtp,
	mailPath string,
	from string,
	passwordUrlTemplate string,
	activateUrlTemplate string,
	deviceErrorTo string,
	mainDomain string) *Mail {

	return &Mail{
		smtp:                       smtp,
		resetPasswordTemplatePath:  mailPath + "/reset_password.txt",
		activateTemplatePath:       mailPath + "/activate.txt",
		premiumRequestTemplatePath: mailPath + "/premium_request.txt",
		from:                       from,
		passwordUrlTemplate:        passwordUrlTemplate,
		activateUrlTemplate:        activateUrlTemplate,
		deviceErrorTo:              deviceErrorTo,
		mainDomain:                 mainDomain,
	}
}

func (m *Mail) SendResetPassword(to string, token string) error {
	url := ParseUrl(m.passwordUrlTemplate, token)
	buf, err := ioutil.ReadFile(m.resetPasswordTemplatePath)
	if err != nil {
		return err
	}
	template := string(buf)
	subject, body, err := ParseBody(template, map[string]string{"url": url})
	if err != nil {
		return err
	}
	err = m.smtp.Send(m.from, to, "text/plain", body, subject)
	return err
}

func (m *Mail) SendActivate(to string, token string) error {
	url := ParseUrl(m.activateUrlTemplate, token)
	buf, err := ioutil.ReadFile(m.activateTemplatePath)
	if err != nil {
		return err
	}
	template := string(buf)
	subject, body, err := ParseBody(template, map[string]string{"url": url, "main_domain": m.mainDomain})
	if err != nil {
		return err
	}
	err = m.smtp.Send(m.from, to, "text/plain", body, subject)
	return err
}

func (m *Mail) SendPremiumRequest(to string) error {
	buf, err := ioutil.ReadFile(m.premiumRequestTemplatePath)
	if err != nil {
		return err
	}
	template := string(buf)
	subject, body, err := ParseBody(template, map[string]string{})
	if err != nil {
		return err
	}
	err = m.smtp.Send(m.from, to, "text/plain", body, subject)
	return err
}

func (m *Mail) SendLogs(to string, data string, includeSupport bool) error {
	data += "Thank you for sharing Syncloud device error info, Syncloud support will get back to you shortly.\n"
	data += "If you need to add more details just reply to this email.\n\n"

	log.Printf("sending logs, include support: %v\n", includeSupport)
	if includeSupport {
		to += ", " + m.deviceErrorTo
	}
	return m.smtp.Send(m.from, to, "text/plain", data, "Device error report")
}

func ParseUrl(template string, token string) string {
	return strings.ReplaceAll(template, "{0}", token)
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
