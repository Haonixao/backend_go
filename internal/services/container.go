package services

import (
	"backend_go/internal/models/scheduled_models"
	"backend_go/internal/services/backup_services"
	"backend_go/internal/services/scheduled_services"
	"backend_go/pkg/gorm_extra"
)

type Container struct {
	TasksService     gorm_extra.BaseService[scheduled_models.Task, string]
	TaskLogsService  scheduled_services.TaskLogsService
	TaskLocksService gorm_extra.BaseService[scheduled_models.TaskLock, string]
	BackupService    backup_services.BackupService
}
