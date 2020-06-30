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

func (config *Config) Load(path string) {
	parser, err := configparser.NewConfigParserFromFile(path)
	if err != nil {
		log.Fatalln("Cannot load config: ", path, err)
	}
	config.parser = parser
}

func (config *Config) GetMySqlLogin() string {
	login, err := config.parser.Get("mysql", "login")
	if err != nil {
		log.Fatalln("Cannot read login: ", err)
	}
	return login
}

func (config *Config) GetMySqlPassword() string {
	password, err := config.parser.Get("mysql", "password")
	if err != nil {
		log.Fatalln("Cannot read password: ", err)
	}
	return password
}

func (config *Config) GetMySqlDB() string {
	db, err := config.parser.Get("mysql", "db")
	if err != nil {
		log.Fatalln("Cannot read db: ", err)
	}
	return db
}

func (config *Config) GetApiSocket() string {
	socket, err := config.parser.Get("api", "socket")
	if err != nil {
		log.Fatalln("Cannot read api socket: ", err)
	}
	return socket
}
