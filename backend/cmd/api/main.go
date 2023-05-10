package main

import (
	"fmt"
	"github.com/syncloud/redirect/db"
	"github.com/syncloud/redirect/ioc"
	"github.com/syncloud/redirect/log"
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
	err = c.Call(func(api *rest.Api, database *db.MySql) error {
		err := database.Connect()
		if err != nil {
			return err
		}
		return api.Start()
	})
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
