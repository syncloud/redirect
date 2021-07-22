package utils

import (
	"github.com/bigkevmcd/go-configparser"
	"log"
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
	secretParser, err := configparser.NewConfigParserFromFile(secret)
	if err != nil {
		log.Fatalln("Cannot load secret config: ", secret, err)
	}
	for _, section := range secretParser.Sections() {
		items, _ := secretParser.Items(section)
		for key, value := range items {
			err := parser.Set(section, key, value)
			if err != nil {
				log.Fatalln("Cannot apply secret config: ", err)
			}
		}
	}
	config.parser = parser
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

func (config *Config) StatsdServer() string {
	value, err := config.parser.Get("stats", "server")
	if err != nil {
		log.Fatalln("Cannot read config: ", err)
	}
	return value
}
func (config *Config) StatsdPrefix() string {
	value, err := config.parser.Get("stats", "prefix")
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

func (config *Config) PayPalPlanId() string {
	result, err := config.parser.Get("paypal", "plan_id")
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
