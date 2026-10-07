package app_container

import (
	"github.com/google/uuid"

	"backend_go/internal/models/public"
	"backend_go/internal/models/scheduled_models"
	"backend_go/internal/services"
	"backend_go/internal/services/backup_services"
	"backend_go/internal/services/scheduled_services"
	"backend_go/pkg/gorm_extra"
)

func (c *Container) InitServices() *Container {
	servicesContainer := services.Container{
		TasksService: &gorm_extra.BaseServiceImpl[scheduled_models.Task, string]{
			BaseRepository: &gorm_extra.BaseRepositoryImpl[scheduled_models.Task, string]{},
		},
		TaskLogsService: &scheduled_services.TaskLogsServiceImpl{
			BaseService: &gorm_extra.BaseServiceImpl[scheduled_models.TaskLog, string]{
				BaseRepository: &gorm_extra.BaseRepositoryImpl[scheduled_models.TaskLog, string]{},
			},
			TaskLogsRep: c.RepositoriesContainer.TaskLogsRep,
		},
		TaskLocksService: &gorm_extra.BaseServiceImpl[scheduled_models.TaskLock, string]{
			BaseRepository: &gorm_extra.BaseRepositoryImpl[scheduled_models.TaskLock, string]{},
		},
		BackupService: &backup_services.BackupServiceImpl{
			BaseService: &gorm_extra.BaseServiceImpl[public.Backup, uuid.UUID]{
				BaseRepository: &gorm_extra.BaseRepositoryImpl[public.Backup, uuid.UUID]{},
			},
			BackupRep: c.RepositoriesContainer.BackupRepository,
		},
	}

	c.ServicesContainer = &servicesContainer

	return c
}
