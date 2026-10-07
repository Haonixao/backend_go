package scheduler

import (
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"backend_go/internal/config"
	"backend_go/internal/models/public"
	"backend_go/internal/models/scheduled_models"
	"backend_go/internal/services"
	"backend_go/pkg/format_errors"
	"backend_go/pkg/log"
)

func GetBackupTask() *scheduled_models.Task {
	return &scheduled_models.Task{
		ID:          "create_periodic_backup",
		Schedule:    "0 0 * * 0",
		IsEnabled:   true,
		Description: "Периодическое создание бэкапа базы данных",
		TaskFunc: func(task *scheduled_models.Task, servicesContainer *services.Container, isReady *atomic.Bool) {
			logger := log.GetLogger(config.GetConfig().Log)
			taskLocksService := servicesContainer.TaskLocksService
			backupTaskLock, err := taskLocksService.Get(&scheduled_models.TaskLock{TaskID: task.ID})
			if err != nil {
				logger.Error().Err(errors.New("ошибка получения блокировки задачи")).Msg(format_errors.FormatTree(err))
				return
			}
			if backupTaskLock == nil {
				isReady.Store(false)
				defer isReady.Store(true)
				backupTaskLock := &scheduled_models.TaskLock{
					TaskID:    task.ID,
					CreatedAt: time.Now(),
				}
				err = taskLocksService.Create(backupTaskLock)
				if err != nil {
					logger.Error().Err(errors.New("ошибка создания блокировки задачи")).Msg(format_errors.FormatTree(err))
					return
				}
				defer func() {
					err = taskLocksService.Delete(backupTaskLock)
					if err != nil {
						logger.Error().Err(errors.New("ошибка удаления блокировки задачи")).Msg(format_errors.FormatTree(err))
						return
					}
				}()
				logger.Info().Str("task_id", task.ID).Msg("задача запущена")
				backupService := servicesContainer.BackupService
				var resErr error
				var backupCreated *public.Backup
				for i := range 3 {
					backupCreated, resErr = backupService.CreateBackup()
					if resErr == nil {
						break
					}
					logger.Error().Err(fmt.Errorf("не удалось создать бэкап %v раз. Повтор...", i+1)).Msg(format_errors.FormatTree(resErr))
				}
				taskIsGood := resErr == nil
				if !taskIsGood {
					logger.Error().Str("task_id", task.ID).Err(errors.New("ошибка создания бэкапа")).Msg(format_errors.FormatTree(resErr))
				} else {
					logger.Info().
						Str("task_id", task.ID).
						Str("backup_id", backupCreated.ID.String()).
						Str("size", backupCreated.Size).
						Str("migration_version", backupCreated.MigrationVersion).
						Msg("бэкап успешно создан")
					oneYearAgo := time.Now().AddDate(-1, 0, 0)
					err = backupService.DeleteOldBackups(oneYearAgo)
					if err != nil {
						logger.Error().Err(errors.New("ошибка удаления старых бэкапов")).Msg(format_errors.FormatTree(err))
					} else {
						logger.Info().Str("task_id", task.ID).Msg("старые бэкапы успешно удалены")
					}
				}
				logger.Info().Str("task_id", task.ID).Msg("задача завершена")
			} else {
				logger.Warn().Str("task_id", task.ID).Msg("задача уже заблокирована. пропуск.")
				if backupTaskLock.CreatedAt.Add(time.Hour * 24).Before(time.Now()) {
					logger.Warn().Str("task_id", task.ID).Msg("очистка устаревшей блокировки")
					err = taskLocksService.Delete(backupTaskLock)
					if err != nil {
						logger.Error().Err(errors.New("ошибка очистки устаревшей блокировки")).Msg(format_errors.FormatTree(err))
						return
					}
				}
			}
		},
	}
}
