package main

import (
	"bufio"
	"fmt"
	"github.com/syncloud/redirect/db"
	"github.com/syncloud/redirect/model"
	"github.com/syncloud/redirect/service"
	"github.com/syncloud/redirect/smtp"
	"github.com/syncloud/redirect/utils"
	"log"
	"os"
	"strings"
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
	database := db.NewMySql(config.GetMySqlHost(), config.GetMySqlDB(), config.GetMySqlLogin(), config.GetMySqlPassword())
	database.Connect()
	smtpClient := smtp.NewSmtp(config.SmtpHost(), config.SmtpPort(), config.SmtpTls(),
		config.SmtpLogin(), config.SmtpPassword())
	mail := service.NewMail(smtpClient, mailPath, config.MailFrom(), config.MailDeviceErrorTo(), config.Domain())
	notification := NewNotification(database, mail, "./sent", sqlEmailFilter)
	notification.Send()

}

type Db interface {
	GetUsersByField(field string, value string) ([]*model.User, error)
}

type Mail interface {
	SendReleaseAnnouncement(to string) error
}

type Notification struct {
	db             Db
	mail           Mail
	sentFile       string
	sqlEmailFilter string
}

func NewNotification(db Db, mail Mail, sentFile string, sqlEmailFilter string) *Notification {
	return &Notification{
		db:             db,
		mail:           mail,
		sentFile:       sentFile,
		sqlEmailFilter: sqlEmailFilter,
	}
}

func (n *Notification) LoadSentEmails() map[string]bool {
	file, err := os.Open(n.sentFile)
	if err != nil {
		return nil
	}
	defer file.Close()
	sentEmails := make(map[string]bool)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) != "" {
			sentEmails[scanner.Text()] = true
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Print(err)
		return nil
	}
	return sentEmails
}

func (n *Notification) Send() {
	sentEmails := n.LoadSentEmails()
	fmt.Printf("previously sent to %d emails\n", len(sentEmails))

	users, err := n.db.GetUsersByField("email", n.sqlEmailFilter)
	if err != nil {
		return
	}

	file, err := os.OpenFile(n.sentFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("")
		return
	}
	defer file.Close()

	for _, user := range users {
		if !user.NotificationEnabled {
			fmt.Println("notification disabled: ", user.Email)
			continue
		}
		if sentEmails[user.Email] == true {
			fmt.Println("already sent: ", user.Email)
			continue
		}
		fmt.Println("sending: ", user.Email)
		err := n.mail.SendReleaseAnnouncement(user.Email)
		if err != nil {
			fmt.Println("send error: ", err)
		} else {
			_, err = file.WriteString(fmt.Sprintf("%s\n", user.Email))
			if err != nil {
				fmt.Println("sent file write error: ", err)
				return
			}
		}
		time.Sleep(time.Second)
	}
}
