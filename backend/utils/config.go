package utils

import (
	"github.com/bigkevmcd/go-configparser"
	"log"
	"os"
)

type Config struct {
	parser *configparser.ConfigParser
}

func NewConfig() *Config {
	return &Config{}
}

func (config *Config) Load(path string, secret string) {
	parser, err := configparser.NewConfigParserFromFile(path)
	if err != nil {
		log.Fatalln("Cannot load config: ", path, err)
	}
	config.parser = parser
	config.mergeSecret(secret, true)
}

func (config *Config) Merge(path string) {
	config.mergeSecret(path, true)
}

func (config *Config) mergeSecret(path string, required bool) {
	if _, err := os.Stat(path); err != nil {
		if required {
			log.Fatalln("Cannot load secret config: ", path, err)
		}
		return
	}
	secretParser, err := configparser.NewConfigParserFromFile(path)
	if err != nil {
		log.Fatalln("Cannot load secret config: ", path, err)
	}
	for _, section := range secretParser.Sections() {
		if !config.parser.HasSection(section) {
			if err := config.parser.AddSection(section); err != nil {
				log.Fatalln("Cannot apply secret config: ", err)
			}
		}
		items, _ := secretParser.Items(section)
		for key, value := range items {
			err := config.parser.Set(section, key, value)
			if err != nil {
				log.Fatalln("Cannot apply secret config: ", err)
			}
		}
	}
}

func (config *Config) GetMySqlLogin() string {
	login, err := config.parser.Get("mysql", "user")
	if err != nil {
		log.Fatalln("Cannot read config: ", err)
	}
	return login
}

func (config *Config) GetMySqlHost() string {
	host, err := config.parser.Get("mysql", "host")
	if err != nil {
		log.Fatalln("Cannot read config: ", err)
	}
	return host
}

func (config *Config) GetMySqlPassword() string {
	password, err := config.parser.Get("mysql", "passwd")
	if err != nil {
		log.Fatalln("Cannot read config: ", err)
	}
	return password
}

func (config *Config) GetMySqlDB() string {
	db, err := config.parser.Get("mysql", "db")
	if err != nil {
		log.Fatalln("Cannot read config: ", err)
	}
	return db
}

func (config *Config) GetApiSocket() string {
	value, err := config.parser.Get("api", "socket")
	if err != nil {
		log.Fatalln("Cannot read config: ", err)
	}
	return value
}

func (config *Config) GetWwwSocket() string {
	value, err := config.parser.Get("www", "socket")
	if err != nil {
		log.Fatalln("Cannot read config: ", err)
	}
	return value
}

func (config *Config) GetApiMetricsAddr() string {
	if value, err := config.parser.Get("metrics", "api_addr"); err == nil {
		return value
	}
	return ":9091"
}

func (config *Config) GetWwwMetricsAddr() string {
	if value, err := config.parser.Get("metrics", "www_addr"); err == nil {
		return value
	}
	return ":9092"
}

func (config *Config) GetRelayPluginAddr() string {
	if value, err := config.parser.Get("relay", "plugin_addr"); err == nil {
		return value
	}
	return "127.0.0.1:7501"
}

func (config *Config) relayLimitBytes(key string, gib int64) int64 {
	if value, err := config.parser.GetInt64("relay", key); err == nil {
		return value
	}
	return gib * 1024 * 1024 * 1024
}

func (config *Config) GetRelayFreeLimitBytes() int64 {
	return config.relayLimitBytes("free_limit_bytes", 1)
}

func (config *Config) GetRelayProLimitBytes() int64 {
	return config.relayLimitBytes("pro_limit_bytes", 10)
}

func (config *Config) GetRelayMaxLimitBytes() int64 {
	return config.relayLimitBytes("max_limit_bytes", 100)
}

func (config *Config) GetRelayPollIntervalSeconds() int {
	if value, err := config.parser.GetInt64("relay", "poll_interval_seconds"); err == nil {
		return int(value)
	}
	return 60
}

func (config *Config) GetFrpsMetricsUrl() string {
	if value, err := config.parser.Get("relay", "frps_metrics_url"); err == nil {
		return value
	}
	return "http://127.0.0.1:7500/metrics"
}

func (config *Config) GetRelayAddress() string {
	if value, err := config.parser.Get("relay", "address"); err == nil {
		return value
	}
	return ""
}

func (config *Config) GetMailRelayAddress() string {
	if value, err := config.parser.Get("mail_relay", "address"); err == nil {
		return value
	}
	return ":587"
}

func (config *Config) GetMailRelayCertFile() string {
	if value, err := config.parser.Get("mail_relay", "cert_file"); err == nil {
		return value
	}
	return ""
}

func (config *Config) GetMailRelayKeyFile() string {
	if value, err := config.parser.Get("mail_relay", "key_file"); err == nil {
		return value
	}
	return ""
}

func (config *Config) GetMailRelaySesRegion() string {
	if value, err := config.parser.Get("mail_relay", "ses_region"); err == nil {
		return value
	}
	return "us-east-1"
}

func (config *Config) GetMailRelaySesConfigurationSet() string {
	if value, err := config.parser.Get("mail_relay", "ses_configuration_set"); err == nil {
		return value
	}
	return ""
}

func (config *Config) GetMailRelayMaxMessageBytes() int64 {
	if value, err := config.parser.GetInt64("mail_relay", "max_message_bytes"); err == nil {
		return value
	}
	return 25 * 1024 * 1024
}

func (config *Config) mailRelayLimitMessages(key string, messages int64) int64 {
	if value, err := config.parser.GetInt64("mail_relay", key); err == nil {
		return value
	}
	return messages
}

func (config *Config) GetMailRelayFreeLimitMessages() int64 {
	return config.mailRelayLimitMessages("free_limit_messages", 50)
}

func (config *Config) GetMailRelayProLimitMessages() int64 {
	return config.mailRelayLimitMessages("pro_limit_messages", 2000)
}

func (config *Config) GetMailRelayMaxLimitMessages() int64 {
	return config.mailRelayLimitMessages("max_limit_messages", 10000)
}

func (config *Config) AwsAccessKeyId() string {
	value, err := config.parser.Get("aws", "access_key_id")
	if err != nil {
		log.Fatalln("Cannot read config: ", err)
	}
	return value
}

func (config *Config) AwsSecretAccessKey() string {
	value, err := config.parser.Get("aws", "secret_access_key")
	if err != nil {
		log.Fatalln("Cannot read config: ", err)
	}
	return value
}

func (config *Config) AwsHostedZoneId() string {
	value, err := config.parser.Get("aws", "hosted_zone_id")
	if err != nil {
		log.Fatalln("Cannot read config: ", err)
	}
	return value
}

func (config *Config) Domain() string {
	value, err := config.parser.Get("redirect", "domain")
	if err != nil {
		log.Fatalln("Cannot read config: ", err)
	}
	return value
}

func (config *Config) SmtpHost() string {
	value, err := config.parser.Get("smtp", "host")
	if err != nil {
		log.Fatalln("Cannot read config: ", err)
	}
	return value
}

func (config *Config) SmtpPort() int {
	value, err := config.parser.GetInt64("smtp", "port")
	if err != nil {
		log.Fatalln("Cannot read config: ", err)
	}
	return int(value)
}

func (config *Config) SmtpTls() bool {
	contains, err := config.parser.HasOption("smtp", "use_tls")
	if err != nil {
		log.Fatalln("Cannot read config: ", err)
	}
	if !contains {
		return false
	}
	value, err := config.parser.GetBool("smtp", "use_tls")
	if err != nil {
		log.Fatalln("Cannot read config: ", err)
	}
	return value
}

func (config *Config) SmtpLogin() string {
	value, err := config.parser.Get("smtp", "login")
	if err != nil {
		log.Fatalln("Cannot read config: ", err)
	}
	return value
}

func (config *Config) SmtpPassword() string {
	value, err := config.parser.Get("smtp", "password")
	if err != nil {
		log.Fatalln("Cannot read config: ", err)
	}
	return value
}

func (config *Config) MailFrom() string {
	value, err := config.parser.Get("mail", "from")
	if err != nil {
		log.Fatalln("Cannot read config: ", err)
	}
	return value
}

func (config *Config) MailDeviceErrorTo() string {
	value, err := config.parser.Get("mail", "device_error")
	if err != nil {
		log.Fatalln("Cannot read config: ", err)
	}
	return value
}

func (config *Config) ActivateByEmail() bool {
	value, err := config.parser.GetBool("redirect", "activate_by_email")
	if err != nil {
		log.Fatalln("Cannot read config: ", err)
	}
	return value
}

func (config *Config) AuthSecretSey() string {
	result, err := config.parser.Get("redirect", "auth_secret_key")
	if err != nil {
		log.Fatalln("Cannot read config: ", err)
	}
	return result
}

func (config *Config) PayPalPlanMonthlyId() string {
	result, err := config.parser.Get("paypal", "plan_monthly_id")
	if err != nil {
		log.Fatalln("Cannot read config: ", err)
	}
	return result
}

func (config *Config) PayPalPlanAnnualId() string {
	result, err := config.parser.Get("paypal", "plan_annual_id")
	if err != nil {
		log.Fatalln("Cannot read config: ", err)
	}
	return result
}

func (config *Config) PayPalPlanMaxMonthlyId() string {
	result, _ := config.parser.Get("paypal", "plan_max_monthly_id")
	return result
}

func (config *Config) PayPalPlanMaxAnnualId() string {
	result, _ := config.parser.Get("paypal", "plan_max_annual_id")
	return result
}

func (config *Config) PayPalUrl() string {
	result, err := config.parser.Get("paypal", "url")
	if err != nil {
		log.Fatalln("Cannot read config: ", err)
	}
	return result
}

func (config *Config) PayPalClientId() string {
	result, err := config.parser.Get("paypal", "client_id")
	if err != nil {
		log.Fatalln("Cannot read config: ", err)
	}
	return result
}

func (config *Config) PayPalSecretId() string {
	result, err := config.parser.Get("paypal", "secret_id")
	if err != nil {
		log.Fatalln("Cannot read config: ", err)
	}
	return result
}

func (config *Config) StripeSecretKey() string {
	result, err := config.parser.Get("stripe", "secret_key")
	if err != nil {
		log.Fatalln("Cannot read config: ", err)
	}
	return result
}

func (config *Config) StripePriceMonthlyId() string {
	result, err := config.parser.Get("stripe", "price_monthly_id")
	if err != nil {
		log.Fatalln("Cannot read config: ", err)
	}
	return result
}

func (config *Config) StripePriceAnnualId() string {
	result, err := config.parser.Get("stripe", "price_annual_id")
	if err != nil {
		log.Fatalln("Cannot read config: ", err)
	}
	return result
}

func (config *Config) StripePriceMaxMonthlyId() string {
	result, _ := config.parser.Get("stripe", "price_max_monthly_id")
	return result
}

func (config *Config) StripePriceMaxAnnualId() string {
	result, _ := config.parser.Get("stripe", "price_max_annual_id")
	return result
}

func (config *Config) UserCleanerEnabled() bool {
	result, err := config.parser.GetBool("cleaner", "user")
	if err != nil {
		log.Fatalln("Cannot read config: ", err)
	}
	return result
}
