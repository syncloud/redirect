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
	stripesdk "github.com/stripe/stripe-go/v81"
	"github.com/syncloud/redirect/change"
	"github.com/syncloud/redirect/clock"
	"github.com/syncloud/redirect/db"
	"github.com/syncloud/redirect/dns"
	"github.com/syncloud/redirect/log"
	"github.com/syncloud/redirect/mail"
	"github.com/syncloud/redirect/mail/inbound"
	"github.com/syncloud/redirect/mail/outbound"
	"github.com/syncloud/redirect/metrics"
	"github.com/syncloud/redirect/payment"
	"github.com/syncloud/redirect/probe"
	"github.com/syncloud/redirect/product"
	"github.com/syncloud/redirect/relay"
	"github.com/syncloud/redirect/rest"
	"github.com/syncloud/redirect/service"
	"github.com/syncloud/redirect/smtp"
	"github.com/syncloud/redirect/subscription"
	"github.com/syncloud/redirect/user"
	"github.com/syncloud/redirect/utils"
	"go.uber.org/zap"
	"net/http"
	"time"
)

// route53 is a global service and signs with this region wherever it runs
const awsRoute53Region = "us-east-1"

func NewContainer(configPath string, secretPath string, mailPath string) (container.Container, error) {
	var logger = log.Default()

	c := container.New()

	config := utils.NewConfig()
	config.Load(configPath, secretPath)
	config.Merge(filepath.Join(filepath.Dir(secretPath), "payments.cfg"))

	if url := config.StripeUrl(); url != "" {
		stripesdk.SetBackend(stripesdk.APIBackend, stripesdk.GetBackendWithConfig(
			stripesdk.APIBackend, &stripesdk.BackendConfig{URL: stripesdk.String(url)}))
	}

	err := c.Singleton(func() *utils.Config {
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

	err = c.Singleton(func(config *utils.Config) *db.Migrator {
		return db.NewMigrator(config, logger)
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

	err = c.Singleton(func(session *session.Session, config *utils.Config) *route53.Route53 {
		return route53.New(session, aws.NewConfig().
			WithEndpoint(config.AwsEndpoint()).
			WithRegion(awsRoute53Region))
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

	err = c.Singleton(func(route53 *route53.Route53, metrics *metrics.Metrics, config *utils.Config) *dns.AmazonDns {
		return dns.New(route53, metrics, 255, config.Domain(), logger)
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
			config.MailSubjectPrefix(),
			config.MailDeviceErrorTo(),
			config.Domain(),
			logger,
		)
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(
		database *db.MySql,
		mailService *service.Mail,
		config *utils.Config,
	) (*product.Orders, error) {
		stripeCheckout := payment.NewStripe(
			config.StripeSecretKey(),
			fmt.Sprintf("https://www.%s/shop", config.Domain()),
			fmt.Sprintf("https://www.%s/shop", config.Domain()),
			logger)
		paypalCheckout, err := payment.NewPayPal(
			config.PayPalClientId(),
			config.PayPalSecretId(),
			config.PayPalUrl(),
			fmt.Sprintf("https://www.%s/shop", config.Domain()),
			fmt.Sprintf("https://www.%s/shop", config.Domain()),
			logger)
		if err != nil {
			return nil, err
		}
		orders := product.NewOrders(
			product.NewCatalog(product.Devices(), product.Shipping),
			product.NewCheckouts(paypalCheckout, stripeCheckout),
			database,
			mailService,
			logger,
		)
		return orders, nil
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(orders *product.Orders) *product.Reconciler {
		return product.NewReconciler(orders, 10*time.Minute, 2*time.Minute, logger)
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
			config.PayPalSdkUrl(),
			config.PayPalPlanMonthlyId(),
			config.PayPalPlanAnnualId(),
			config.PayPalPlanMaxMonthlyId(),
			config.PayPalPlanMaxAnnualId(),
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
			config.StripePriceMaxMonthlyId(),
			config.StripePriceMaxAnnualId(),
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
		mailService *service.Mail,
		actions *service.Actions,
		config *utils.Config,
		subscriptions *subscription.Router,
	) *service.Users {
		return service.NewUsers(
			database,
			config.ActivateByEmail(),
			actions,
			mailService,
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
		return service.NewDomains(amazonDns, database, users, metrics, config.Domain(), config.AwsHostedZoneId(),
			detector, config.GetRelayAddress())
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

	err = c.Singleton(func(config *utils.Config) *relay.FrpsMetrics {
		return relay.NewFrpsMetrics(config.GetFrpsMetricsUrl())
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(database *db.MySql, config *utils.Config) *relay.Tiers {
		return relay.NewTiers(database, config.GetRelayFreeLimitBytes(), config.GetRelayProLimitBytes(), config.GetRelayMaxLimitBytes(), logger)
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(database *db.MySql, tiers *relay.Tiers) *relay.Usage {
		return relay.NewUsage(database, tiers)
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(database *db.MySql, mailService *service.Mail) *relay.LimitWarner {
		return relay.NewLimitWarner(database, mailService)
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(
		frps *relay.FrpsMetrics,
		database *db.MySql,
		tiers *relay.Tiers,
		warner *relay.LimitWarner,
		config *utils.Config,
	) *relay.Accountant {
		interval := time.Duration(config.GetRelayPollIntervalSeconds()) * time.Second
		return relay.NewAccountant(frps, database, tiers, warner, interval, logger)
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(
		domains *service.Domains,
		accountant *relay.Accountant,
		config *utils.Config,
	) *relay.AuthServer {
		return relay.NewAuthServer(config.GetRelayPluginAddr(), domains, accountant, config.Domain(), logger)
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(database *db.MySql, config *utils.Config) *outbound.Tiers {
		return outbound.NewTiers(database,
			config.GetMailOutboundFreeLimitMessages(),
			config.GetMailOutboundProLimitMessages(),
			config.GetMailOutboundMaxLimitMessages(), logger)
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(database *db.MySql, tiers *outbound.Tiers) *outbound.AccountUsage {
		return outbound.NewAccountUsage(database, tiers)
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(database *db.MySql, mailService *service.Mail) *outbound.LimitWarner {
		return outbound.NewLimitWarner(database, mailService)
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(database *db.MySql) *outbound.DbStore {
		return outbound.NewDbStore(database)
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(
		domains *service.Domains,
		tiers *outbound.Tiers,
		store *outbound.DbStore,
		warner *outbound.LimitWarner,
	) *outbound.Relay {
		return outbound.New(domains, tiers, store, store, warner, logger)
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(store *outbound.DbStore, config *utils.Config) *outbound.Feedback {
		return outbound.NewFeedback(store, store,
			config.GetMailOutboundBounceRatio(), config.GetMailOutboundBounceMinimum(), logger)
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func() *clock.SystemClock {
		return clock.New()
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(database *db.MySql, systemClock *clock.SystemClock) *outbound.UsageMetrics {
		return outbound.NewUsageMetrics(database, systemClock, logger)
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(awsSession *session.Session, config *utils.Config) *outbound.Reputation {
		source := outbound.NewCloudWatch(awsSession, config.GetMailOutboundSesRegion(), time.Hour)
		return outbound.NewReputation(source, time.Duration(config.GetMailOutboundReputationIntervalSeconds())*time.Second, logger)
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(config *utils.Config) *outbound.Limiter {
		return outbound.NewLimiter(outbound.Limits{
			Minute:     config.GetMailOutboundLimitPerMinute(),
			Hour:       config.GetMailOutboundLimitPerHour(),
			Day:        config.GetMailOutboundLimitPerDay(),
			Recipients: config.GetMailOutboundMaxRecipients(),
		})
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(config *utils.Config) *outbound.Rspamd {
		return outbound.NewRspamd(config, 10*time.Second, logger)
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(awsSession *session.Session, config *utils.Config) outbound.Sender {
		return outbound.NewSesSender(awsSession, config.GetMailOutboundSesRegion(),
			config.GetMailOutboundSesEndpoint(), config.GetMailOutboundSesConfigurationSet(), logger)
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(config *utils.Config) *mail.Connections {
		return mail.NewConnections(config.GetMailOutboundMaxConnectionsPerPeer())
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(config *utils.Config) *mail.InFlight {
		return mail.NewInFlight(config.GetMailOutboundMaxConcurrentSends())
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(
		relayService *outbound.Relay,
		sender outbound.Sender,
		scanner *outbound.Rspamd,
		limiter *outbound.Limiter,
		connections *mail.Connections,
		inFlight *mail.InFlight,
		config *utils.Config,
	) *outbound.Server {
		return outbound.NewServer(config.GetMailOutboundAddress(), config.Domain(),
			relayService, sender, scanner, limiter, connections, inFlight,
			config.GetMailOutboundMaxMessageBytes(), logger)
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(database *db.MySql) *inbound.Router {
		return inbound.NewRouter(database)
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(config *utils.Config) inbound.DeviceDialer {
		return inbound.NewTunnelDialer(config.GetMailInboundMuxer(), inbound.DialTimeout)
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(router *inbound.Router, dialer inbound.DeviceDialer,
		accountant *relay.Accountant, config *utils.Config) *inbound.Server {
		return inbound.NewServer(
			config.GetMailInboundAddress(),
			config.GetMailInboundHostname(),
			router,
			dialer,
			mail.NewConnections(config.GetMailInboundMaxConnectionsPerPeer()),
			mail.NewInFlight(config.GetMailInboundMaxConcurrent()),
			accountant,
			inbound.NewPtrResolver(inbound.ResolveTimeout),
			config.GetMailInboundMaxMessageBytes(),
			logger)
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(
		domains *service.Domains,
		users *service.Users,
		mailService *service.Mail,
		prober *probe.Service,
		certbot *service.Certbot,
		metrics *metrics.Metrics,
		feedback *outbound.Feedback,
		config *utils.Config,
	) *rest.Api {
		return rest.NewApi(
			domains,
			users,
			mailService,
			prober,
			certbot,
			metrics,
			feedback,
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
		mailService *service.Mail,
		actions *service.Actions,
		stripe *subscription.Stripe,
		paypal *subscription.PayPal,
		orders *product.Orders,
		usage *relay.Usage,
		mailUsage *outbound.AccountUsage,
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
			mailService,
			stripe,
			orders,
			usage,
			mailUsage,
			paypal,
			metrics,
			config.Domain(),
			secretKey,
			config.GetWwwSocket(),
			config.GetWwwRateLimitPerMinute(),
			config.GetWwwRateLimitPerHour(),
			logger,
		), nil
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(
		database *db.MySql,
		domains *service.Domains,
		mailService *service.Mail,
		metrics *metrics.Metrics,
	) *dns.Cleaner {
		return dns.NewCleaner(
			database,
			domains,
			mailService,
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
		mailService *service.Mail,
		config *utils.Config,
		domains *service.Domains,
		router *subscription.Router,
	) *user.Cleaner {
		return user.NewCleaner(
			database,
			state,
			mailService,
			domains,
			router,
			config.UserCleanerEnabled(),
			logger,
		)
	})
	if err != nil {
		return nil, err
	}

	err = c.Singleton(func(
		database *db.MySql,
		mailService *service.Mail,
		config *utils.Config,
	) *user.ActivationSender {
		return user.NewActivationSender(
			database,
			mailService,
			config.ActivateByEmail(),
			logger,
		)
	})
	if err != nil {
		return nil, err
	}

	return c, nil
}
