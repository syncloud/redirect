package main

import (
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/spf13/cobra"
	"github.com/syncloud/redirect/db"
	"github.com/syncloud/redirect/ioc"
	"github.com/syncloud/redirect/utils"
)

func migrator(configFile string, secretFile string) (*migrate.Migrate, error) {
	config := utils.NewConfig()
	config.Load(configFile, secretFile)

	source, err := iofs.New(db.Migrations, "migrations")
	if err != nil {
		return nil, err
	}
	dsn := fmt.Sprintf("mysql://%s:%s@tcp(%s:3306)/%s",
		config.GetMySqlLogin(), config.GetMySqlPassword(), config.GetMySqlHost(), config.GetMySqlDB())
	return migrate.NewWithSourceInstance("iofs", source, dsn)
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
			m, err := migrator(configFile, secretFile)
			if err != nil {
				return err
			}
			if err := m.Up(); err != nil && err != migrate.ErrNoChange {
				return err
			}
			return nil
		},
	})

	root.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "print the applied version",
		RunE: func(_ *cobra.Command, _ []string) error {
			m, err := migrator(configFile, secretFile)
			if err != nil {
				return err
			}
			version, dirty, err := m.Version()
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
			m, err := migrator(configFile, secretFile)
			if err != nil {
				return err
			}
			var version int
			if _, err := fmt.Sscanf(args[0], "%d", &version); err != nil {
				return err
			}
			return m.Force(version)
		},
	})

	if err := root.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
