package db

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"go.uber.org/zap"
)

type MigratorConfig interface {
	GetMySqlHost() string
	GetMySqlDB() string
	GetMySqlLogin() string
	GetMySqlPassword() string
}

type Migrator struct {
	config MigratorConfig
	logger *zap.Logger
}

func NewMigrator(config MigratorConfig, logger *zap.Logger) *Migrator {
	return &Migrator{config: config, logger: logger}
}

func (m *Migrator) Start() error {
	migrator, err := m.migrator()
	if err != nil {
		return err
	}
	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	version, dirty, err := migrator.Version()
	if err != nil {
		return err
	}
	m.logger.Info("database schema", zap.Uint("version", version), zap.Bool("dirty", dirty))
	return nil
}

func (m *Migrator) Version() (uint, bool, error) {
	migrator, err := m.migrator()
	if err != nil {
		return 0, false, err
	}
	return migrator.Version()
}

func (m *Migrator) Force(version int) error {
	migrator, err := m.migrator()
	if err != nil {
		return err
	}
	return migrator.Force(version)
}

func (m *Migrator) migrator() (*migrate.Migrate, error) {
	source, err := iofs.New(Migrations, "migrations")
	if err != nil {
		return nil, err
	}
	return migrate.NewWithSourceInstance("iofs", source, fmt.Sprintf("mysql://%s:%s@tcp(%s:3306)/%s",
		m.config.GetMySqlLogin(), m.config.GetMySqlPassword(), m.config.GetMySqlHost(), m.config.GetMySqlDB()))
}
