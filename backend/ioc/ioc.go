package ioc

import (
	"encoding/base64"
	"fmt"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/golobby/container/v3"
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
		config.Merge(filepath.Join(filepath.Dir(secretPath), "payments.cfg"))
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

	err = c.Singleton(func() *metrics.Metrics {
		return metrics.New()
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(database *db.MySql) *metrics.DbGauges {
		return metrics.NewDbGauges(database, logger)
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(route53 *route53.Route53, metrics *metrics.Metrics) *dns.AmazonDns {
		return dns.New(route53, metrics, 255, logger)
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func() *dns.PublicResolver {
		return dns.NewPublicResolver()
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
		config *utils.Config,
	) *subscription.Stripe {
		return subscription.NewStripe(
			config.StripeSecretKey(),
			config.StripePriceMonthlyId(),
			config.StripePriceAnnualId(),
			fmt.Sprintf("https://www.%s/account?stripe_session_id={CHECKOUT_SESSION_ID}", config.Domain()),
			fmt.Sprintf("https://www.%s/account", config.Domain()),
			logger,
		)
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(
		paypal *subscription.PayPal,
		stripe *subscription.Stripe,
	) *subscription.Router {
		return subscription.NewRouter(paypal, stripe)
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(
		database *db.MySql,
		mail *service.Mail,
		actions *service.Actions,
		config *utils.Config,
		subscriptions *subscription.Router,
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
		metrics *metrics.Metrics,
		config *utils.Config,
	) *service.Domains {
		return service.NewDomains(amazonDns, database, users, metrics, config.Domain(), config.AwsHostedZoneId(), detector)
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(
		database *db.MySql,
		amazonDns *dns.AmazonDns,
		resolver *dns.PublicResolver,
		config *utils.Config,
	) *service.NsChecker {
		return service.NewNsChecker(database, amazonDns, resolver, config.AwsHostedZoneId())
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
		domains *service.Domains,
		users *service.Users,
		mail *service.Mail,
		prober *probe.Service,
		certbot *service.Certbot,
		metrics *metrics.Metrics,
		config *utils.Config,
	) *rest.Api {
		return rest.NewApi(
			domains,
			users,
			mail,
			prober,
			certbot,
			metrics,
			config.Domain(),
			config.GetApiSocket(),
			logger,
		)
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(
		domains *service.Domains,
		nsChecker *service.NsChecker,
		users *service.Users,
		mail *service.Mail,
		actions *service.Actions,
		stripe *subscription.Stripe,
		metrics *metrics.Metrics,
		config *utils.Config,
	) (*rest.Www, error) {
		secretKey, err := base64.StdEncoding.DecodeString(config.AuthSecretSey())
		if err != nil {
			logger.Error("unable to decode secret key", zap.Error(err))
			return nil, err
		}
		return rest.NewWww(
			domains,
			nsChecker,
			users,
			actions,
			mail,
			stripe,
			metrics,
			config.Domain(),
			config.PayPalPlanMonthlyId(),
			config.PayPalPlanAnnualId(),
			config.PayPalClientId(),
			secretKey,
			config.GetWwwSocket(),
			logger,
		), nil
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(
		database *db.MySql,
		domains *service.Domains,
		mail *service.Mail,
		metrics *metrics.Metrics,
	) *dns.Cleaner {
		return dns.NewCleaner(
			database,
			domains,
			mail,
			metrics,
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
		config *utils.Config,
		domains *service.Domains,
		router *subscription.Router,
	) *user.Cleaner {
		return user.NewCleaner(
			database,
			state,
			mail,
			domains,
			router,
			config.UserCleanerEnabled(),
			logger,
		)
	})
	if err != nil {
		return nil, err
	}

	return c, nil
}
