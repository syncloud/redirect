package main

import (
	"fmt"
	"github.com/smira/go-statsd"
	"github.com/syncloud/redirect/db"
	"github.com/syncloud/redirect/dns"
	"github.com/syncloud/redirect/rest"
	"github.com/syncloud/redirect/service"
	"github.com/syncloud/redirect/utils"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		log.Println("usage: ", os.Args[0], "config.cfg", "secret.cfg")
		return
	}

	config := utils.NewConfig()
	config.Load(os.Args[1], os.Args[2])
	database := db.NewMySql()
	database.Connect(config.GetMySqlHost(), config.GetMySqlDB(), config.GetMySqlLogin(), config.GetMySqlPassword())

	statsdClient := statsd.NewClient(fmt.Sprintf("%s:8125", config.StatsdServer()),
		statsd.MaxPacketSize(1400),
		statsd.MetricPrefix(fmt.Sprintf("%s.", config.StatsdPrefix())))

	dnsImp := dns.New(statsdClient, config.AwsAccessKeyId(), config.AwsSecretAccessKey(), config.AwsHostedZoneId())

	dynamicDns := service.New(dnsImp, database, config.Domain())

	api := rest.NewApi(statsdClient, dynamicDns)
	api.Start(config.GetApiSocket())
}

