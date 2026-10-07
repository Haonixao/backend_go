package gorm_extra

import (
	"errors"
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"

	"backend_go/pkg/format_errors"
)

var squashMigration = "squash_migration"

type Migrator struct {
	migrations     *gormigrate.Gormigrate
	migrationsList []*gormigrate.Migration
	conn           *gorm.DB
}

func GetMigrator(migrationsList []*gormigrate.Migration, cfg *Postgres) (*Migrator, error) {
	migrationsList = append(migrationsList,
		&gormigrate.Migration{
			// Миграция для очистки
			// Упаковать несколько миграций в одну и потом удалить записи старых миграций из бд
			ID: squashMigration,
			Migrate: func(tx *gorm.DB) error {
				return tx.Transaction(func(tx *gorm.DB) error {
					mToDel := []string{
						"-temp-",
					}
					return format_errors.Wrap(tx.Exec("delete from public.migrations WHERE id in ?;", mToDel).Error, "ошибка удаления temp миграций")
				})
			},
			Rollback: func(tx *gorm.DB) error {
				return nil
			},
		})
	m := &Migrator{
		migrationsList: migrationsList,
	}
	conn, err := GetPostgresConn(cfg)
	if err != nil {
		return nil, format_errors.Wrap(err, "ошибка соединения с бд")
	}
	m.conn = conn
	m.migrations = gormigrate.New(m.conn, gormigrate.DefaultOptions, migrationsList)
	return m, nil
}

func (m *Migrator) ApplyMigrations() error {
	migrationsIDs := []string{}
	for _, migration := range m.migrationsList {
		if err := m.MigrateToMigration(migration.ID); err != nil {
			backErr := m.migrations.RollbackMigration(migration)
			if backErr != nil {
				return format_errors.Wrap(backErr, "ошибка rollback миграции "+migration.ID)
			}
			return format_errors.Wrap(err, "ошибка миграции "+migration.ID)
		}
		if migration.ID == squashMigration {
			if err := m.conn.Exec("delete from public.migrations WHERE id = ?;", migration.ID).Error; err != nil {
				return format_errors.Wrap(err, "ошибка удаления squashMigration")
			}
		}
		migrationsIDs = append(migrationsIDs, migration.ID)
	}
	var notValidMigrationsIDs []string
	err := m.conn.Raw("select id from public.migrations where id not in (?)", migrationsIDs).Scan(&notValidMigrationsIDs).Error
	if err != nil {
		return format_errors.Wrap(err, "ошибка поиска невалидных миграций")
	}
	if len(notValidMigrationsIDs) > 0 {
		err = errors.New("найдены ID невалидных миграций. лучше пересоздать БД: " + fmt.Sprintf("%v", notValidMigrationsIDs))
		return format_errors.Wrap(err, "невалидные миграции")
	}
	return nil
}

func (m *Migrator) RollbackToMigration(migrationID string) error {
	err := m.migrations.RollbackTo(migrationID)
	if err != nil {
		err = format_errors.Wrap(err, "ошибка rollback")
	}
	return err
}

func (m *Migrator) MigrateToMigration(migrationID string) error {
	err := m.migrations.MigrateTo(migrationID)
	if err != nil {
		err = format_errors.Wrap(err, "ошибка миграции")
	}
	return err
}
