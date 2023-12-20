package main

import (
	"github.com/spf13/cobra"
	"github.com/syncloud/redirect/db"
	"github.com/syncloud/redirect/ioc"
	"github.com/syncloud/redirect/log"
	"github.com/syncloud/redirect/metrics"
	"github.com/syncloud/redirect/rest"
	"github.com/syncloud/redirect/service"
)

func main() {
	var configFile string
	var secretFile string
	var mailDir string
	cmd := &cobra.Command{
		Use: "www",
		RunE: func(cmd *cobra.Command, args []string) error {
			log.EnableStdOutLog()
			c, err := ioc.NewContainer(configFile, secretFile, mailDir)
			if err != nil {
				return err
			}
			return c.Call(func(
				www *rest.Www,
				database *db.MySql,
				metrics *metrics.Publisher,
				graphite *metrics.GraphiteClient,
			) error {
				services := []service.Startable{
					database,
					graphite,
					metrics,
					www,
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
	cmd.Flags().StringVar(&configFile, "config-file", ioc.ConfigFile, "config file")
	cmd.Flags().StringVar(&secretFile, "secret-file", ioc.SecretFile, "secret file")
	cmd.Flags().StringVar(&mailDir, "mail-dir", ioc.MailDir, "mail dir")

	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}
