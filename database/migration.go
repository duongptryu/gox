package database

import (
	"database/sql"

	"github.com/duongptryu/gox/syserr"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type MigrationManager struct {
	db      *sql.DB
	migrate *migrate.Migrate
}

func NewMigrationManager(db *sql.DB, databaseConfig *Config, migrationPath string) (*MigrationManager, error) {
	var (
		driver database.Driver
		err    error
	)

	switch databaseConfig.Type {
	case "postgres":
		driver, err = postgres.WithInstance(db, &postgres.Config{})
		if err != nil {
			return nil, err
		}
	default:
		return nil, syserr.New(syserr.InvalidArgumentCode, "unsupported database type",
			syserr.F("database_type", databaseConfig.Type))
	}

	m, err := migrate.NewWithDatabaseInstance(
		migrationPath,
		databaseConfig.Name,
		driver,
	)
	if err != nil {
		return nil, syserr.WrapAsIs(err, "failed to create migrate instance")
	}

	return &MigrationManager{
		db:      db,
		migrate: m,
	}, nil
}

func (m *MigrationManager) Up() error {
	if err := m.migrate.Up(); err != nil {
		return syserr.WrapAsIs(err, "failed to migrate up")
	}
	return nil
}

func (m *MigrationManager) Down() error {
	if err := m.migrate.Down(); err != nil {
		return syserr.WrapAsIs(err, "failed to migrate down")
	}
	return nil
}

func (m *MigrationManager) Force(version int) error {
	if err := m.migrate.Force(version); err != nil {
		return syserr.WrapAsIs(err, "failed to force migrate", syserr.F("version", version))
	}
	return nil
}

func (m *MigrationManager) Version() (uint, bool, error) {
	version, dirty, err := m.migrate.Version()
	if err != nil {
		return 0, false, syserr.WrapAsIs(err, "failed to get version")
	}
	return version, dirty, nil
}

func (m *MigrationManager) Close() error {
	return m.db.Close()
}
