package app_container

import "backend_go/internal/scheduler"

func (c *Container) InitScheduler() *Container {
	tasksContainer := scheduler.TasksContainer{
		BackupTask:  scheduler.GetBackupTask(),
		UpdateTasks: scheduler.GetUpdateTasks(),
	}

	sch := scheduler.Scheduler{
		TasksService:      c.ServicesContainer.TasksService,
		ServicesContainer: c.ServicesContainer,
		IsReady:           c.IsReady,
		TasksContainer:    &tasksContainer,
	}

	c.Scheduler = &sch
	return c
}
