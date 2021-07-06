package cmd

import (
	"fmt"
	"github.com/smira/go-statsd"
	"github.com/syncloud/redirect/db"
	"github.com/syncloud/redirect/dns"
	"github.com/syncloud/redirect/rest"
	"github.com/syncloud/redirect/service"
	"github.com/syncloud/redirect/smtp"
	"github.com/syncloud/redirect/utils"
	"log"
	"os"
)

type Main struct {
	config *utils.Config
	api    *rest.Api
	www    *rest.Www
}

func NewMain() *Main {
	if len(os.Args) < 4 {
		log.Println("usage: ", os.Args[0], "config.cfg", "secret.cfg", "mail_dir")
		return nil
	}

	config := utils.NewConfig()
	config.Load(os.Args[1], os.Args[2])
	mailPath := os.Args[3]
	database := db.NewMySql()
	database.Connect(config.GetMySqlHost(), config.GetMySqlDB(), config.GetMySqlLogin(), config.GetMySqlPassword())

	statsdClient := statsd.NewClient(fmt.Sprintf("%s:8125", config.StatsdServer()),
		statsd.MaxPacketSize(1400),
		statsd.MetricPrefix(fmt.Sprintf("%s.", config.StatsdPrefix())))
	dnsImp := dns.New(statsdClient, config.AwsAccessKeyId(), config.AwsSecretAccessKey())
	actions := service.NewActions(database)
	smtpClient := smtp.NewSmtp(config.SmtpHost(), config.SmtpPort(), config.SmtpTls(),
		config.SmtpLogin(), config.SmtpPassword())
	mail := service.NewMail(smtpClient, mailPath, config.MailFrom(), config.MailPasswordUrlTemplate(),
		config.MailActivateUrlTemplate(), config.MailDeviceErrorTo(), config.Domain())
	users := service.NewUsers(database, config.ActivateByEmail(), actions, mail)
	domains := service.NewDomains(dnsImp, database, users, config.Domain(), config.AwsHostedZoneId())
	probe := service.NewPortProbe(database)
	api := rest.NewApi(statsdClient, domains, users, actions, mail, probe, config.Domain())
	www := rest.NewWww(statsdClient, domains, users, actions, mail, probe, config.Domain())
	return &Main{config: config, api: api, www: www}

}

func (m *Main) StartApi() {
	m.api.StartApi(m.config.GetApiSocket())
}

func (m *Main) StartWww() {
	m.www.StartWww(m.config.GetWwwSocket())
}
