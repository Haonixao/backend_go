package handlers

import (
	"backend_go/internal/handlers/pages"
	"backend_go/internal/handlers/status"
	"backend_go/internal/handlers/tasks"
)

type Container struct {
	StatusHandler *status.Handler
	TasksHandler  *tasks.Handler
	PagesHandler  *pages.Handler
}
