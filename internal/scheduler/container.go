package scheduler

import (
	"backend_go/internal/models/scheduled_models"
)

type TasksContainer struct {
	BackupTask  *scheduled_models.Task
	UpdateTasks *scheduled_models.Task
}
