package scheduler

import (
	"errors"

	"backend_go/internal/config"
	"backend_go/internal/services"
	"backend_go/pkg/format_errors"

	"backend_go/internal/models/scheduled_models"
	"backend_go/pkg/log"
	"backend_go/pkg/utils/maps"
)

func GetUpdateTasks() *scheduled_models.Task {
	return &scheduled_models.Task{
		ID:          "update_tasks",
		Schedule:    "*/5 * * * *",
		IsEnabled:   true,
		Description: "Обновление задач и их параметров в приложении",
		TaskFunc: func(task *scheduled_models.Task, servicesContainer *services.Container, sch *Scheduler) {
			logger := log.GetLogger(config.GetConfig().Log)
			s := servicesContainer.TasksService
			logger.Info().Str("task_id", task.ID).Msg("задача запущена")
			tasksIds := []string{}
			for _, task := range sch.GetTasks() {
				tasksIds = append(tasksIds, task.ID)
			}
			existedTasks, err := s.GetManyByIds(tasksIds)
			if err != nil {
				logger.Error().Err(errors.New("ошибка получения существующих задач")).Msg(format_errors.FormatTree(err))
				return
			}
			existedTasksMap := maps.MapFromHash(existedTasks)
			for _, task := range sch.GetTasks() {
				if existedTask, ok := existedTasksMap[task.ID]; ok {
					if task.IsEnabled != existedTask.IsEnabled || task.Schedule != existedTask.Schedule {
						err = sch.UpdateTaskJob(existedTask)
						if err != nil {
							logger.Error().Err(errors.New("ошибка обновления задачи")).Msg(format_errors.FormatTree(err))
							return
						}
					}
				}
			}
			logger.Info().Str("task_id", task.ID).Msg("задача завершена")
		},
	}
}
