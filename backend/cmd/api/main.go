package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/syncloud/redirect/db"
	"github.com/syncloud/redirect/dns"
	"github.com/syncloud/redirect/ioc"
	"github.com/syncloud/redirect/log"
	"github.com/syncloud/redirect/metrics"
	"github.com/syncloud/redirect/rest"
	"github.com/syncloud/redirect/service"
	"os"
)

func main() {
	var configFile string
	var secretFile string
	var mailDir string
	var cmd = &cobra.Command{
		Use: "api",
		RunE: func(cmd *cobra.Command, args []string) error {
			log.EnableStdOutLog()
			c, err := ioc.NewContainer(configFile, secretFile, mailDir)
			if err != nil {
				return err
			}
			return c.Call(func(
				api *rest.Api,
				database *db.MySql,
				dnsCleaner *dns.Cleaner,
				graphite *metrics.GraphiteClient,
			) error {
				services := []service.Startable{
					database,
					dnsCleaner,
					graphite,
					api,
				}
				for _, s := range services {
					err := s.Start()
					if err != nil {
						return err
					}
				}
				return nil
			})
		},
	}
	cmd.Flags().StringVar(&configFile, "config-file", "", "config file")
	_ = cmd.MarkFlagRequired("config-file")
	cmd.Flags().StringVar(&secretFile, "secret-file", "", "secret file")
	_ = cmd.MarkFlagRequired("secret-file")
	cmd.Flags().StringVar(&mailDir, "mail-dir", "", "mail dir")
	_ = cmd.MarkFlagRequired("mail-dir")

	err := cmd.Execute()
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}
