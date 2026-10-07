package app_container

import (
	"github.com/google/uuid"

	"backend_go/internal/config"
	"backend_go/internal/models/public"
	"backend_go/internal/models/scheduled_models"
	"backend_go/internal/repositories"
	"backend_go/internal/repositories/backup_repositories"
	"backend_go/internal/repositories/scheduled_repositories"
	"backend_go/pkg/gorm_extra"
)

func (c *Container) InitRepositories() *Container {
	repositoriesContainer := repositories.Container{
		TaskLogsRep: &scheduled_repositories.TaskLogsRepositoryImpl{
			BaseRepository: &gorm_extra.BaseRepositoryImpl[scheduled_models.TaskLog, string]{
				PCfg: config.GetConfig().Postgres,
			},
		},
		BackupRepository: &backup_repositories.BackupRepositoryImpl{
			BaseRepository: &gorm_extra.BaseRepositoryImpl[public.Backup, uuid.UUID]{
				PCfg: config.GetConfig().Postgres,
			},
		},
	}
	c.RepositoriesContainer = &repositoriesContainer
	return c
}
