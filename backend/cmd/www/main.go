package main

import (
	"fmt"
	"github.com/syncloud/redirect/db"
	"github.com/syncloud/redirect/ioc"
	"github.com/syncloud/redirect/log"
	"github.com/syncloud/redirect/metrics"
	"github.com/syncloud/redirect/rest"
	"os"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Printf("usage: %s config.cfg secret.cfg mail_dir\n", os.Args[0])
		os.Exit(1)
	}
	log.EnableStdOutLog()
	c, err := ioc.NewContainer(os.Args[1], os.Args[2], os.Args[3])
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	err = c.Call(func(www *rest.Www, database *db.MySql, metrics *metrics.Publisher) error {
		err := database.Connect()
		if err != nil {
			return err
		}
		metrics.Start()
		return www.Start()
	})
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
