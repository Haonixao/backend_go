package repositories

import (
	"backend_go/internal/repositories/backup_repositories"
	"backend_go/internal/repositories/scheduled_repositories"
)

type Container struct {
	TaskLogsRep      scheduled_repositories.TaskLogsRepository
	BackupRepository backup_repositories.BackupRepository
}
