package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/syncloud/redirect/db"
	"github.com/syncloud/redirect/ioc"
	"github.com/syncloud/redirect/log"
	"github.com/syncloud/redirect/utils"
)

func migrator(configFile string, secretFile string) *db.Migrator {
	config := utils.NewConfig()
	config.Load(configFile, secretFile)
	return db.NewMigrator(config, log.Default())
}

func main() {
	var configFile string
	var secretFile string

	root := &cobra.Command{Use: "migrate", SilenceUsage: true}
	root.PersistentFlags().StringVar(&configFile, "config-file", ioc.ConfigFile, "config file")
	root.PersistentFlags().StringVar(&secretFile, "secret-file", ioc.SecretFile, "secret file")

	root.AddCommand(&cobra.Command{
		Use:   "up",
		Short: "apply pending migrations",
		RunE: func(_ *cobra.Command, _ []string) error {
			return migrator(configFile, secretFile).Start()
		},
	})

	root.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "print the applied version",
		RunE: func(_ *cobra.Command, _ []string) error {
			version, dirty, err := migrator(configFile, secretFile).Version()
			if err != nil {
				return err
			}
			fmt.Printf("%d dirty=%v\n", version, dirty)
			return nil
		},
	})

	root.AddCommand(&cobra.Command{
		Use:   "force [version]",
		Short: "record a version as applied without running it",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			var version int
			if _, err := fmt.Sscanf(args[0], "%d", &version); err != nil {
				return err
			}
			return migrator(configFile, secretFile).Force(version)
		},
	})

	if err := root.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
