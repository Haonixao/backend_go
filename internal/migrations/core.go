package migrations

import (
	"backend_go/internal/app_container"
	"backend_go/internal/models/public"
	"backend_go/internal/models/scheduled_models"
	"backend_go/pkg/format_errors"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func GetMigrationsList(appContainer *app_container.Container) []*gormigrate.Migration {
	mList := []*gormigrate.Migration{
		{
			ID: "0",
			Migrate: func(tx *gorm.DB) error {
				return tx.Transaction(func(tx *gorm.DB) error {
					sql := `ALTER TABLE migrations ADD IF NOT EXISTS applied_at timestamptz DEFAULT current_timestamp NOT NULL`
					err := tx.Exec(sql).Error
					if err != nil {
						err = format_errors.Wrap(err, "ошибка миграции")
					}
					return err
				})
			},

			Rollback: func(tx *gorm.DB) error {
				return tx.Transaction(func(tx *gorm.DB) error {
					sql := `ALTER TABLE public.migrations DROP COLUMN applied_at`
					err := tx.Exec(sql).Error
					if err != nil {
						err = format_errors.Wrap(err, "ошибка миграции")
					}
					return err
				})
			},
		},
		{
			ID: "1",
			Migrate: func(tx *gorm.DB) error {
				var errs []error
				errs = append(errs,
					tx.AutoMigrate(
						&public.Backup{},
					),
					tx.Exec("create schema if not exists scheduled;").Error,
					tx.AutoMigrate(
						&scheduled_models.Task{},
						&scheduled_models.TaskLog{},
						&scheduled_models.TaskLock{},
						&public.Backup{},
					),
				)

				for _, err := range errs {
					if err != nil {
						return format_errors.Wrap(err, "ошибка миграции")
					}
				}
				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Transaction(func(tx *gorm.DB) error {
					var errs []error
					errs = append(
						errs,
						tx.Exec("drop schema if exists scheduled cascade;").Error,
						tx.Exec("drop table if exists public.backup;").Error,
					)

					for _, err := range errs {
						if err != nil {
							return format_errors.Wrap(err, "ошибка миграции")
						}
					}
					return nil
				})
			},
		},
	}
	return mList
}
