package scheduler

import (
	"errors"
	"fmt"
	"slices"
	"sync/atomic"
	"time"

	"backend_go/internal/config"
	"backend_go/internal/models/scheduled_models"
	"backend_go/internal/services"
	"backend_go/pkg/format_errors"
	"backend_go/pkg/gorm_extra"
	"backend_go/pkg/log"

	"github.com/go-co-op/gocron"
)

type Scheduler struct {
	tasksList         []*scheduled_models.Task
	TasksService      gorm_extra.BaseService[scheduled_models.Task, string]
	ServicesContainer *services.Container
	cronScheduler     *gocron.Scheduler
	IsReady           *atomic.Bool
	TasksContainer    *TasksContainer
}

func (s *Scheduler) getCroneJob(task, taskWithFunc *scheduled_models.Task) (*gocron.Job, error) {
	switch taskWithFunc.ID {
	case s.TasksContainer.BackupTask.ID:
		return s.cronScheduler.Cron(task.Schedule).Do(taskWithFunc.TaskFunc, task, s.ServicesContainer, s.IsReady)
	case s.TasksContainer.UpdateTasks.ID:
		return s.cronScheduler.Cron(task.Schedule).Do(taskWithFunc.TaskFunc, task, s.ServicesContainer, s)
	default:
		return s.cronScheduler.Cron(task.Schedule).Do(taskWithFunc.TaskFunc, task, s.ServicesContainer)
	}
}

func (s *Scheduler) GetTasks() []*scheduled_models.Task {
	return s.tasksList
}

func (s *Scheduler) Init() {
	logger := log.GetLogger(config.GetConfig().Log)
	s.tasksList = []*scheduled_models.Task{
		s.TasksContainer.UpdateTasks,
		s.TasksContainer.BackupTask,
	}
	s.cronScheduler = gocron.NewScheduler(time.UTC)
	for _, task := range s.tasksList {
		var err error
		err = s.checkTask(task)
		if err != nil {
			logger.Error().Str("task_id", task.ID).Err(errors.New("ошибка проверки задачи")).Msg(format_errors.FormatTree(err))
			continue
		}
		if task.IsEnabled {
			task.CronJob, err = s.getCroneJob(task, task)
			if err != nil {
				logger.Error().Str("task_id", task.ID).Err(errors.New("ошибка добавления задачи в планировщик")).Msg(format_errors.FormatTree(err))
			}
		} else {
			logger.Warn().Str("task_id", task.ID).Msg("задача выключена")
		}
	}
	currentIDs := []string{}
	for _, task := range s.tasksList {
		currentIDs = append(currentIDs, task.ID)
	}
	notFoundedTasks, err := s.TasksService.GetManyByIdsNotIn(currentIDs)
	if err != nil {
		logger.Error().Err(errors.New("ошибка получения ненайденных задач")).Msg(format_errors.FormatTree(err))
		return
	}
	// Удаление задач из бд которые не добавлены в список планировщика, но есть в бд
	for _, task := range notFoundedTasks {
		task.IsEnabled = false
		err = s.TasksService.Delete(task)
		if err != nil {
			logger.Error().Err(errors.New("ошибка удаления ненайденной задачи")).Msg(format_errors.FormatTree(err))
			continue
		}
	}
	s.cronScheduler.StartAsync()
}

func (s *Scheduler) checkTask(task *scheduled_models.Task) error {
	checkedTask, err := s.TasksService.Get(&scheduled_models.Task{ID: task.ID})
	if err != nil {
		return format_errors.Wrap(err, "ошибка проверки задачи")
	}
	if checkedTask != nil {
		task.IsEnabled = checkedTask.IsEnabled
		task.Schedule = checkedTask.Schedule
	} else {
		err = s.TasksService.Create(task)
		if err != nil {
			return format_errors.Wrap(err, "ошибка проверки задачи")
		}
	}
	return nil
}

func (s *Scheduler) UpdateTaskJob(task *scheduled_models.Task) error {
	taskForUpdateIndex := slices.IndexFunc(s.tasksList, func(t *scheduled_models.Task) bool { return t.ID == task.ID })
	if taskForUpdateIndex == -1 {
		return format_errors.Wrap(fmt.Errorf("задача с id %s не найдена в списке исполняемых задач", task.ID), "ошибка обновления задачи")
	}
	taskForUpdate := s.tasksList[taskForUpdateIndex]
	if task.IsEnabled != taskForUpdate.IsEnabled {
		if taskForUpdate.IsEnabled {
			s.cronScheduler.RemoveByReference(taskForUpdate.CronJob)
		} else {
			var err error
			taskForUpdate.CronJob, err = s.getCroneJob(task, taskForUpdate)
			if err != nil {
				return format_errors.Wrap(err, "ошибка добавления задачи в планировщик")
			}
		}
	} else if task.Schedule != taskForUpdate.Schedule {
		var err error
		s.cronScheduler.RemoveByReference(taskForUpdate.CronJob)
		taskForUpdate.CronJob, err = s.getCroneJob(task, taskForUpdate)
		if err != nil {
			return format_errors.Wrap(err, "ошибка обновления задачи в планировщике")
		}
	}
	taskForUpdate.Schedule = task.Schedule
	taskForUpdate.IsEnabled = task.IsEnabled
	return nil
}
