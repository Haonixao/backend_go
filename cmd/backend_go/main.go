package main

import (
	"backend_go/internal/app_container"
	"backend_go/internal/cors"
	"backend_go/internal/migrations"
	"backend_go/internal/routes"
	"backend_go/pkg/format_errors"
	"backend_go/pkg/gorm_extra"
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	appContainer := app_container.GetContainer()
	logger := appContainer.Logger
	appContainer.IsReady.Store(false)
	go func() {
		defer appContainer.IsReady.Store(true)
		migrator, err := gorm_extra.GetMigrator(migrations.GetMigrationsList(appContainer), appContainer.Config.Postgres)
		if err != nil {
			logger.Fatal().Msg(format_errors.FormatTree(err))
		}
		err = migrator.ApplyMigrations()
		if err != nil {
			logger.Fatal().Msg(format_errors.FormatTree(err))
		}
		// before_start.CreateTempUser(appContainer.ServicesContainer)
		if appContainer.Config.App.SchedulerEnabled {
			appContainer.Scheduler.Init()
		}
	}()
	appMode := appContainer.Config.App.Mode
	if appMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}
	ginEngine := gin.Default()
	ginEngine.Use(cors.GetCors())
	routes.Init(ginEngine, appContainer)
	srv := &http.Server{
		Addr:    ":" + appContainer.Config.App.Port,
		Handler: ginEngine,
	}
	logger.Info().Msg("Scalar включен: " + appContainer.Config.App.ServiceUrl + "/api/scalar/docs")
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal().Msg(format_errors.FormatTree(err))
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info().Msg("завершение работы...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		err := format_errors.Wrap(err, "принудительное завершение работы http сервера")
		logger.Error().Msg(format_errors.FormatTree(err))
	}
	logger.Info().Msg("работа завершена")
}
