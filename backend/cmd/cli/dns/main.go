package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/syncloud/redirect/db"
	"github.com/syncloud/redirect/ioc"
	"github.com/syncloud/redirect/log"
	"os"
	"time"
)

func main() {
	var configFile string
	var secretFile string
	var dryRun bool
	var beforeString string
	var cmd = &cobra.Command{
		Use: "dns",
		RunE: func(cmd *cobra.Command, args []string) error {
			log.EnableStdOutLog()
			c, err := ioc.NewContainer(configFile, secretFile, "")
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			return c.Call(func(database *db.MySql) error {
				err := database.Connect()
				if err != nil {
					return err
				}
				before, err := time.Parse(time.DateOnly, beforeString)
				if err != nil {
					return err
				}

				domain, err := database.GetOldestDomainBefore(before)
				if err != nil {
					return err
				}
				fmt.Printf("id: %d, name: %s, update: %s\n", domain.Id, domain.Name, domain.LastUpdate.String())
				if !dryRun {
					fmt.Printf("will remove\n")
				}
				return nil
			})
		},
	}
	cmd.Flags().StringVar(&configFile, "config-file", "", "config file")
	_ = cmd.MarkFlagRequired("config-file")
	cmd.Flags().StringVar(&secretFile, "secret-file", "", "secret file")
	_ = cmd.MarkFlagRequired("secret-file")
	cmd.Flags().BoolVar(&dryRun, "dry-run", true, "dry run")
	cmd.Flags().StringVar(&beforeString, "before", "", "before date")

	err := cmd.Execute()
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}

/*
sys.path.append(normpath(join(dirname(__file__), '..')))
from redirect import ioc
import argparse

if __name__=='__main__':
    parser = argparse.ArgumentParser(description='This is redirect unsubscribing email tool', usage='%(prog)s [options]')
    parser.add_argument('date', help='data before cleanup')
    parser.add_argument('limit', help='how many to cleanup')
    args = parser.parse_args()

    date = args.date
    limit = args.limit

    manager = ioc.manager()
    create_storage = manager.create_storage
    dns = manager.dns
    with create_storage() as storage:
        for domain in storage.get_domains_last_updated_before(date, limit):
            print('ip: {0}, last update: {1}'.format(domain.ip, domain.last_update))
            dns.delete_domain(manager.main_domain, domain)
            domain.ip = None

*/
