package main

import (
	"github.com/syncloud/redirect/db"
	"github.com/syncloud/redirect/service"
	"github.com/syncloud/redirect/smtp"
	"github.com/syncloud/redirect/utils"
	"log"
	"os"
	"time"
)

func main() {

	if len(os.Args) < 5 {
		log.Println("usage: ", os.Args[0], "config", "secret", "mail_path", "sql_email_filter")
		return
	}

	configPath := os.Args[1]
	configSecretPath := os.Args[2]
	mailPath := os.Args[3]
	sqlEmailFilter := os.Args[4]

	config := utils.NewConfig()
	config.Load(configPath, configSecretPath)
	database := db.NewMySql()
	database.Connect(config.GetMySqlHost(), config.GetMySqlDB(), config.GetMySqlLogin(), config.GetMySqlPassword())
	smtpClient := smtp.NewSmtp(config.SmtpHost(), config.SmtpPort(), config.SmtpTls(),
		config.SmtpLogin(), config.SmtpPassword())
	mail := service.NewMail(smtpClient, mailPath, config.MailFrom(), config.MailDeviceErrorTo(), config.Domain())
	users, err := database.GetUsersByField("email", sqlEmailFilter)
	if err != nil {
		return
	}
	for _, user := range users {
		if user.NotificationEnabled {
			log.Println("sending: ", user.Email)
			err := mail.SendReleaseAnnouncement(user.Email)
			if err != nil {
				log.Println("send error: ", err)
				return
			}
			time.Sleep(time.Second)
		} else {
			log.Println("skipping: ", user.Email)
		}
	}

}
