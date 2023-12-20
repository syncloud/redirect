package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/syncloud/redirect/ioc"
	"github.com/syncloud/redirect/log"
	"github.com/syncloud/redirect/subscription"
)

func main() {
	var configFile string
	var secretFile string
	var mailDir string
	cli := &cobra.Command{
		Use:   "cli",
		Short: "debugging tool",
	}
	cli.PersistentFlags().StringVar(&configFile, "config-file", ioc.ConfigFile, "config file")
	cli.PersistentFlags().StringVar(&secretFile, "secret-file", ioc.SecretFile, "secret file")
	cli.PersistentFlags().StringVar(&mailDir, "mail-dir", ioc.MailDir, "mail dir")

	subs := &cobra.Command{
		Use: "subscription",
	}
	cli.AddCommand(subs)

	subs.AddCommand(&cobra.Command{
		Use:  "details [id]",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			log.EnableStdOutLog()
			c, err := ioc.NewContainer(configFile, secretFile, mailDir)
			if err != nil {
				return err
			}
			return c.Call(func(paypal *subscription.PayPal) error {
				details, err := paypal.GetSubscriptionDetails(args[0])
				if err != nil {
					return err
				}
				fmt.Printf("failed payments: %d\n", details.BillingInfo.FailedPaymentsCount)
				return nil
			})
		},
	})

	err := cli.Execute()
	if err != nil {
		panic(err)
	}
}
