package ioc

import (
	"encoding/base64"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/golobby/container/v3"
	"github.com/smira/go-statsd"
	"github.com/syncloud/redirect/change"
	"github.com/syncloud/redirect/db"
	"github.com/syncloud/redirect/dns"
	"github.com/syncloud/redirect/log"
	"github.com/syncloud/redirect/metrics"
	"github.com/syncloud/redirect/probe"
	"github.com/syncloud/redirect/rest"
	"github.com/syncloud/redirect/service"
	"github.com/syncloud/redirect/smtp"
	"github.com/syncloud/redirect/subscription"
	"github.com/syncloud/redirect/user"
	"github.com/syncloud/redirect/utils"
	"go.uber.org/zap"
	"net/http"
)

func NewContainer(configPath string, secretPath string, mailPath string) (container.Container, error) {
	var logger = log.Default()

	c := container.New()

	err := c.Singleton(func() *utils.Config {
		config := utils.NewConfig()
		config.Load(configPath, secretPath)
		return config
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(config *utils.Config) *db.MySql {
		return db.NewMySql(
			config.GetMySqlHost(),
			config.GetMySqlDB(),
			config.GetMySqlLogin(),
			config.GetMySqlPassword(),
			logger,
		)
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(config *utils.Config) *statsd.Client {
		return statsd.NewClient(fmt.Sprintf("%s:8125", config.StatsdServer()),
			statsd.MaxPacketSize(1400),
			statsd.MetricPrefix(fmt.Sprintf("%s.", config.StatsdPrefix())))
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(config *utils.Config) *metrics.GraphiteClient {
		return metrics.New(config.GraphitePrefix(), config.GraphiteHost(), 2003)
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(config *utils.Config) *session.Session {
		return session.Must(session.NewSession(&aws.Config{
			Credentials: credentials.NewStaticCredentials(
				config.AwsAccessKeyId(),
				config.AwsSecretAccessKey(),
				"",
			),
		}))
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(session *session.Session) *route53.Route53 {
		return route53.New(session)
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(statsd *statsd.Client, route53 *route53.Route53) *dns.AmazonDns {
		return dns.New(statsd, route53, 255, logger)
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(database *db.MySql) *service.Actions {
		return service.NewActions(database)

	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(database *db.MySql, config *utils.Config) *smtp.Smtp {
		return smtp.NewSmtp(config.SmtpHost(), config.SmtpPort(), config.SmtpTls(),
			config.SmtpLogin(), config.SmtpPassword())
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(smtp *smtp.Smtp, config *utils.Config) *service.Mail {
		return service.NewMail(
			smtp,
			mailPath,
			config.MailFrom(),
			config.MailDeviceErrorTo(),
			config.Domain(),
			logger,
		)
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(
		config *utils.Config,
	) (*subscription.PayPal, error) {
		return subscription.New(
			config.PayPalClientId(),
			config.PayPalSecretId(),
			config.PayPalUrl(),
			logger,
		)
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(
		database *db.MySql,
		mail *service.Mail,
		actions *service.Actions,
		config *utils.Config,
		subscriptions *subscription.PayPal,
	) *service.Users {
		return service.NewUsers(
			database,
			config.ActivateByEmail(),
			actions,
			mail,
			subscriptions,
		)
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func() *change.RequestDetector {
		return change.New()
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(
		database *db.MySql,
		users *service.Users,
		detector *change.RequestDetector,
		amazonDns *dns.AmazonDns,
		config *utils.Config,
	) *service.Domains {
		return service.NewDomains(amazonDns, database, users, config.Domain(), config.AwsHostedZoneId(), detector)
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func() *http.Client {
		return probe.NewClient()
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(
		database *db.MySql,
		probeClient *http.Client,
	) *probe.Service {
		return probe.New(database, probeClient)
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(
		database *db.MySql,
		amazonDns *dns.AmazonDns,
	) *service.Certbot {
		return service.NewCertbot(database, amazonDns)
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(
		statsd *statsd.Client,
		domains *service.Domains,
		users *service.Users,
		mail *service.Mail,
		prober *probe.Service,
		certbot *service.Certbot,
		config *utils.Config,
	) *rest.Api {
		return rest.NewApi(
			statsd,
			domains,
			users,
			mail,
			prober,
			certbot,
			config.Domain(),
			config.GetApiSocket(),
			logger,
		)
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(
		statsd *statsd.Client,
		domains *service.Domains,
		users *service.Users,
		mail *service.Mail,
		actions *service.Actions,
		config *utils.Config,
	) (*rest.Www, error) {
		secretKey, err := base64.StdEncoding.DecodeString(config.AuthSecretSey())
		if err != nil {
			logger.Error("unable to decode secret key", zap.Error(err))
			return nil, err
		}
		return rest.NewWww(
			statsd,
			domains,
			users,
			actions,
			mail,
			config.Domain(),
			config.PayPalPlanMonthlyId(),
			config.PayPalPlanAnnualId(),
			config.PayPalClientId(),
			secretKey,
			config.GetWwwSocket(),
		), nil
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(
		database *db.MySql,
		graphite *metrics.GraphiteClient,
	) *metrics.Publisher {
		return metrics.NewPublisher(
			graphite,
			database,
		)
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(
		database *db.MySql,
		domains *service.Domains,
		mail *service.Mail,
		statsd *statsd.Client,
	) *dns.Cleaner {
		return dns.NewCleaner(
			database,
			domains,
			mail,
			statsd,
		)
	})

	if err != nil {
		return nil, err
	}
	err = c.Singleton(func() *user.CleanerState {
		return user.NewCleanerState(
			logger,
		)
	})
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}
	err = c.Singleton(func(
		database *db.MySql,
		state *user.CleanerState,
		mail *service.Mail,
		statsd *statsd.Client,
		config *utils.Config,
		domains *service.Domains,
	) *user.Cleaner {
		return user.NewCleaner(
			database,
			state,
			mail,
			domains,
			config.UserCleanerEnabled(),
			logger,
		)
	})
	if err != nil {
		return nil, err
	}

	return c, nil
}
