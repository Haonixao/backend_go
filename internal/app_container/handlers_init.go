package app_container

import (
	"backend_go/internal/handlers"
	"backend_go/internal/handlers/pages"
	"backend_go/internal/handlers/status"
	"backend_go/internal/handlers/tasks"
)

func (c *Container) InitHandlers() *Container {
	handlersContainer := handlers.Container{
		PagesHandler: &pages.Handler{},
		StatusHandler: &status.Handler{
			BackupService:    c.ServicesContainer.BackupService,
			TaskLocksService: c.ServicesContainer.TaskLocksService,
			Logger:           c.Logger,
		},
		TasksHandler: &tasks.Handler{
			TasksService:     c.ServicesContainer.TasksService,
			TaskLogsService:  c.ServicesContainer.TaskLogsService,
			TaskLocksService: c.ServicesContainer.TaskLocksService,
			Logger:           c.Logger,
		},
	}
	c.HandlersContainer = &handlersContainer
	return c
}
