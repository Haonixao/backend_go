package scheduled_repositories

import (
	"errors"
	"time"

	"backend_go/pkg/format_errors"

	"backend_go/internal/models/scheduled_models"
	"backend_go/pkg/gorm_extra"

	"gorm.io/gorm"
)

type TaskLogsRepository interface {
	gorm_extra.BaseRepository[scheduled_models.TaskLog, string]
	GetLatestByTaskIdAndIsGood(taskId string, isGood *bool) (*scheduled_models.TaskLog, error)
	DeleteManyByTaskIdAndBeforeDate(task *scheduled_models.Task, date *time.Time) error
}

type TaskLogsRepositoryImpl struct {
	gorm_extra.BaseRepository[scheduled_models.TaskLog, string]
}

func (s *TaskLogsRepositoryImpl) GetLatestByTaskIdAndIsGood(taskId string,
	isGood *bool,
) (*scheduled_models.TaskLog, error) {
	var taskLog scheduled_models.TaskLog
	conn, err := s.GetConn()
	if err != nil {
		return nil, format_errors.Wrap(err, "ошибка репозитория")
	}
	query := conn.
		Where("task_id = ?", taskId)
	if isGood != nil {
		query = query.Where("is_good = ?", isGood)
	}
	query = query.Order("date_start DESC")
	err = query.First(&taskLog).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, format_errors.Wrap(err, "ошибка репозитория")
	}
	return &taskLog, nil
}

func (s *TaskLogsRepositoryImpl) DeleteManyByTaskIdAndBeforeDate(task *scheduled_models.Task, date *time.Time) error {
	conn, err := s.GetConn()
	if err != nil {
		return format_errors.Wrap(err, "ошибка репозитория")
	}
	query := conn
	if task != nil {
		query = query.Where("task_id = ?", task.ID)
	}
	if date != nil {
		query = query.Where("date_end < ?", date)
	}
	err = query.Delete(new(scheduled_models.TaskLog)).Error
	if err != nil {
		err = format_errors.Wrap(err, "ошибка репозитория")
	}
	return err
}
