package tasks

import (
	"errors"
	"net/http"

	"github.com/rs/zerolog"

	"backend_go/internal/methods"
	"backend_go/internal/models"
	"backend_go/internal/models/scheduled_models"
	"backend_go/internal/scheduler"
	"backend_go/internal/services/scheduled_services"
	"backend_go/pkg/format_errors"
	"backend_go/pkg/gorm_extra"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	TasksService     gorm_extra.BaseService[scheduled_models.Task, string]
	TaskLogsService  scheduled_services.TaskLogsService
	TaskLocksService gorm_extra.BaseService[scheduled_models.TaskLock, string]
	Logger           *zerolog.Logger
}

func (h *Handler) getTaskById(c *gin.Context, taskID string) (*scheduled_models.Task, error) {
	pathP := &struct {
		ID string `uri:"id" binding:"required"`
	}{ID: taskID}
	if err := methods.BindAndValidate(c, pathP, true); err != nil {
		errS := "ошибка валидации: " + err.Error()
		c.JSON(http.StatusBadRequest, models.MessageResponse{Message: errS})
		return nil, format_errors.Wrap(err, "ошибка валидации")
	}
	task, err := h.TasksService.Get(&scheduled_models.Task{ID: pathP.ID})
	if err != nil {
		errS := "ошибка получения задачи: " + err.Error()
		c.JSON(http.StatusInternalServerError, models.MessageResponse{Message: errS})
		return nil, format_errors.Wrap(err, "ошибка получения задачи")
	}
	if task == nil {
		errS := "задача не найдена: " + taskID
		c.JSON(http.StatusNotFound, models.MessageResponse{Message: errS})
		return nil, format_errors.Wrap(errors.New(errS), "задача не найдена")
	}
	return task, nil
}

// GetTaskLogs
//
// @Tags		tasks
// @Produce	json
// @Param		id	path		string	true	"Task ID"
// @Success	200		{array}		scheduled_models.TaskLog
// @Failure	404		{object}	models.MessageResponse
// @Failure	500		{object}	models.MessageResponse
// @Failure	401		{object}	models.MessageResponse
// @Failure	403		{object}	models.MessageResponse
// @Router		/api/v1/tasks/{id}/logs [get]
func (h *Handler) GetTaskLogs(c *gin.Context) {
	taskID := c.Param("id")
	task, err := h.getTaskById(c, taskID)
	if err != nil {
		h.Logger.Error().Msg(format_errors.FormatTree(err))
		return
	}
	var taskLogs []*scheduled_models.TaskLog
	taskLogs, err = h.TaskLogsService.GetMany(&scheduled_models.TaskLog{TaskID: task.ID})
	if err != nil {
		errS := "ошибка получения логов задач: " + err.Error()
		h.Logger.Error().Msg(format_errors.FormatTree(err))
		c.JSON(http.StatusInternalServerError, models.MessageResponse{Message: errS})
		return
	}
	c.JSON(http.StatusOK, taskLogs)
}

// GetTaskById
//
// @Tags		tasks
// @Produce	json
// @Param		id	path		string	true	"Task ID"
// @Success	200	{object}	scheduled_models.Task
// @Failure	404	{object}	models.MessageResponse
// @Failure	500	{object}	models.MessageResponse
// @Failure	401	{object}	models.MessageResponse
// @Failure	403	{object}	models.MessageResponse
// @Router		/api/v1/tasks/{id} [get]
func (h *Handler) GetTaskById(c *gin.Context) {
	taskID := c.Param("id")
	task, err := h.getTaskById(c, taskID)
	if err != nil {
		h.Logger.Error().Msg(format_errors.FormatTree(err))
		return
	}
	c.JSON(http.StatusOK, task)
}

// GetTasks
//
// @Tags		tasks
// @Produce	json
// @Success	200	{array}		scheduled_models.Task
// @Failure	404	{object}	models.MessageResponse
// @Failure	500	{object}	models.MessageResponse
// @Failure	401	{object}	models.MessageResponse
// @Failure	403	{object}	models.MessageResponse
// @Router		/api/v1/tasks [get]
func (h *Handler) GetTasks(c *gin.Context) {
	tasks, err := h.TasksService.GetAll()
	if err != nil {
		errS := "ошибка получения задач: " + err.Error()
		h.Logger.Error().Msg(format_errors.FormatTree(err))
		c.JSON(http.StatusInternalServerError, models.MessageResponse{Message: errS})
		return
	}
	c.JSON(http.StatusOK, tasks)
}

// UpdateTaskById
//
// @Tags		tasks
// @Produce	json
// @Param		id				path		string				true	"Task ID"
// @Param		updateTaskBody	body		UpdateTaskRequest	true	"Task update body"
// @Success	200				{object}	string				"Task ID"
// @Failure	400				{object}	models.MessageResponse
// @Failure	404				{object}	models.MessageResponse
// @Failure	500				{object}	models.MessageResponse
// @Failure	401				{object}	models.MessageResponse
// @Failure	403				{object}	models.MessageResponse
// @Router		/api/v1/tasks/{id} [patch]
func (h *Handler) UpdateTaskById(scheduler *scheduler.Scheduler) gin.HandlerFunc {
	return func(c *gin.Context) {
		taskID := c.Param("id")
		task, err := h.getTaskById(c, taskID)
		if err != nil {
			h.Logger.Error().Msg(format_errors.FormatTree(err))
			return
		}
		updateRequest := &UpdateTaskRequest{}
		if err := methods.BindAndValidate(c, updateRequest, false); err != nil {
			errS := "ошибка валидации: " + err.Error()
			h.Logger.Error().Msg(format_errors.FormatTree(err))
			c.JSON(http.StatusBadRequest, models.MessageResponse{Message: errS})
			return
		}
		if updateRequest.Schedule != nil {
			task.Schedule = *updateRequest.Schedule
		}
		if updateRequest.IsEnabled != nil {
			task.IsEnabled = *updateRequest.IsEnabled
		}
		if err = h.TasksService.Update(task); err != nil {
			err, status := methods.DuplicationError(err)
			errS := "ошибка обновления: " + err.Error()
			h.Logger.Error().Msg(format_errors.FormatTree(err))
			c.JSON(status, models.MessageResponse{Message: errS})
			return
		}
		if err = scheduler.UpdateTaskJob(task); err != nil {
			errS := "ошибка обновления задачи планировщика: " + err.Error()
			h.Logger.Error().Msg(format_errors.FormatTree(err))
			c.JSON(http.StatusInternalServerError, models.MessageResponse{Message: errS})
			return
		}
		c.JSON(http.StatusOK, task.ID)
	}
}
