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
	socket, err := config.parser.Get("api", "socket")
	if err != nil {
		log.Fatalln("Cannot read config: ", err)
	}
	return socket
}

func (config *Config) AwsAccessKeyId() string {
	socket, err := config.parser.Get("aws", "access_key_id")
	if err != nil {
		log.Fatalln("Cannot read config: ", err)
	}
	return socket
}

func (config *Config) AwsSecretAccessKey() string {
	socket, err := config.parser.Get("aws", "secret_access_key")
	if err != nil {
		log.Fatalln("Cannot read config: ", err)
	}
	return socket
}

func (config *Config) AwsHostedZoneId() string {
	socket, err := config.parser.Get("aws", "hosted_zone_id")
	if err != nil {
		log.Fatalln("Cannot read config: ", err)
	}
	return socket
}

func (config *Config) StatsdServer() string {
	socket, err := config.parser.Get("stats", "server")
	if err != nil {
		log.Fatalln("Cannot read config: ", err)
	}
	return socket
}
func (config *Config) StatsdPrefix() string {
	socket, err := config.parser.Get("stats", "prefix")
	if err != nil {
		log.Fatalln("Cannot read config: ", err)
	}
	return socket
}

func (config *Config) Domain() string {
	socket, err := config.parser.Get("redirect", "domain")
	if err != nil {
		log.Fatalln("Cannot read config: ", err)
	}
	return socket
}

func (config *Config) MockDns() bool {
	result, err := config.parser.GetBool("redirect", "mock_dns")
	if err != nil {
		log.Fatalln("Cannot read config: ", err)
	}
	return result
}
