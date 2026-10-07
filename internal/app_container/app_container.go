package app_container

import (
	"sync/atomic"

	"github.com/rs/zerolog"

	"backend_go/internal/config"
	"backend_go/internal/handlers"
	"backend_go/internal/middleware"
	"backend_go/internal/repositories"
	"backend_go/internal/scheduler"
	"backend_go/internal/services"
	"backend_go/pkg/log"
)

type Container struct {
	Config                *config.Config
	HandlersContainer     *handlers.Container
	ServicesContainer     *services.Container
	RepositoriesContainer *repositories.Container
	IsReady               *atomic.Bool
	Scheduler             *scheduler.Scheduler
	Logger                *zerolog.Logger
	MiddlewareContainer   *middleware.Container
}

func GetContainer() *Container {
	logger := log.GetLogger(config.GetConfig().Log)

	isReady := atomic.Bool{}

	cfg := config.GetConfig()

	appContainer := &Container{
		Config:  cfg,
		IsReady: &isReady,
		Logger:  logger,
	}

	appContainer = appContainer.InitRepositories().InitServices().InitHandlers()

	appContainer.MiddlewareContainer = &middleware.Container{
		ServicesContainer: appContainer.ServicesContainer,
		HandlersContainer: appContainer.HandlersContainer,
		IsReady:           appContainer.IsReady,
	}

	appContainer = appContainer.InitScheduler()

	return appContainer
}
