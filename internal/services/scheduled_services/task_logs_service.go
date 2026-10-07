package scheduled_services

import (
	"time"

	"backend_go/pkg/format_errors"

	"backend_go/internal/models/scheduled_models"
	"backend_go/internal/repositories/scheduled_repositories"
	"backend_go/pkg/gorm_extra"
)

type TaskLogsService interface {
	gorm_extra.BaseService[scheduled_models.TaskLog, string]
	GetLatestTaskLogByTaskIdAndIsGood(taskId string, isGood *bool) (*scheduled_models.TaskLog, error)
	DeleteTaskLogsByTaskIdAndBeforeDate(task *scheduled_models.Task, date *time.Time) error
}

type TaskLogsServiceImpl struct {
	gorm_extra.BaseService[scheduled_models.TaskLog, string]
	TaskLogsRep scheduled_repositories.TaskLogsRepository
}

func (s *TaskLogsServiceImpl) GetLatestTaskLogByTaskIdAndIsGood(taskId string,
	isGood *bool,
) (*scheduled_models.TaskLog, error) {
	r, err := s.TaskLogsRep.GetLatestByTaskIdAndIsGood(taskId, isGood)
	if err != nil {
		return r, format_errors.Wrap(err, "ошибка сервиса")
	}
	return r, nil
}

func (s *TaskLogsServiceImpl) DeleteTaskLogsByTaskIdAndBeforeDate(task *scheduled_models.Task, date *time.Time) error {
	err := s.TaskLogsRep.DeleteManyByTaskIdAndBeforeDate(task, date)
	if err != nil {
		return format_errors.Wrap(err, "ошибка сервиса")
	}
	return nil
}
