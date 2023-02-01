package cmd

import (
	"encoding/base64"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/smira/go-statsd"
	"github.com/syncloud/redirect/change"
	"github.com/syncloud/redirect/db"
	"github.com/syncloud/redirect/dns"
	"github.com/syncloud/redirect/metrics"
	"github.com/syncloud/redirect/probe"
	"github.com/syncloud/redirect/rest"
	"github.com/syncloud/redirect/service"
	"github.com/syncloud/redirect/smtp"
	"github.com/syncloud/redirect/utils"
	"log"
	"os"
	"time"
)

type Main struct {
	config         *utils.Config
	api            *rest.Api
	www            *rest.Www
	graphiteClient *metrics.GraphiteClient
	database       *db.MySql
}

type logWriter struct {
}

func (writer logWriter) Write(bytes []byte) (int, error) {
	return fmt.Print(string(bytes))
}

func NewMain() *Main {
	log.SetFlags(0)
	log.SetOutput(new(logWriter))

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
	graphiteClient := metrics.New(config.GraphitePrefix(), config.GraphiteHost(), 2003)
	mySession := session.Must(session.NewSession(&aws.Config{Credentials: credentials.NewStaticCredentials(config.AwsAccessKeyId(), config.AwsSecretAccessKey(), "")}))
	client := route53.New(mySession)
	amazonDns := dns.New(statsdClient, client, 255)
	actions := service.NewActions(database)
	smtpClient := smtp.NewSmtp(config.SmtpHost(), config.SmtpPort(), config.SmtpTls(),
		config.SmtpLogin(), config.SmtpPassword())
	mail := service.NewMail(smtpClient, mailPath, config.MailFrom(), config.MailDeviceErrorTo(), config.Domain())
	users := service.NewUsers(database, config.ActivateByEmail(), actions, mail)
	detector := change.New()
	domains := service.NewDomains(amazonDns, database, users, config.Domain(), config.AwsHostedZoneId(), detector)
	probeClient := probe.NewClient()
	prober := probe.New(database, probeClient)
	api := rest.NewApi(statsdClient, domains, users, mail, prober, service.NewCertbot(database, amazonDns), config.Domain())
	secretKey, err := base64.StdEncoding.DecodeString(config.AuthSecretSey())
	if err != nil {
		log.Fatalf("unable to decode secre key %v", err)
	}
	www := rest.NewWww(statsdClient, domains, users, actions, mail, config.Domain(), config.PayPalPlanId(), config.PayPalClientId(), secretKey)
	return &Main{
		config:         config,
		api:            api,
		www:            www,
		graphiteClient: graphiteClient,
		database:       database,
	}

}

func (m *Main) StartApi() {
	m.api.StartApi(m.config.GetApiSocket())
}

func (m *Main) StartWww() {
	m.StartMetrics()
	m.www.StartWww(m.config.GetWwwSocket())
}

func (m *Main) StartMetrics() {
	m.graphiteClient.Start()
	devicesGauge := m.graphiteClient.Graphite.NewGauge("db.devices")
	usersGauge := m.graphiteClient.Graphite.NewGauge("db.users")
	go func() {
		for {
			count, err := m.database.GetOnlineDevicesCount()
			if err != nil {
				log.Printf("db error %v", err)
			} else {
				devicesGauge.Set(float64(count))
			}
			count, err = m.database.GetUsersCount()
			if err != nil {
				log.Printf("db error %v", err)
			} else {
				usersGauge.Set(float64(count))
			}
			time.Sleep(10 * time.Second)
		}
	}()
}
