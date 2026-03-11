package db

import "github.com/golang-migrate/migrate/v4"

func RunMigrations(migrationURL string, dbSource string) error {
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		return err
	}

	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}
