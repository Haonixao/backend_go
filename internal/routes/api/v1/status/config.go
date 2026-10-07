package status

import (
	"backend_go/internal/app_container"
	"backend_go/internal/config"

	"github.com/gin-gonic/gin"
)

func Init(baseGroup *gin.RouterGroup, appContainer *app_container.Container) {
	statusV1 := baseGroup.Group("/status")

	handler := appContainer.HandlersContainer.StatusHandler

	statusV1.GET("/check", handler.GetStatus)
	statusV1.GET("/backups", handler.GetBackups)
	statusV1.GET("/backups/:id/data", handler.GetBackupData())
	statusV1.POST("/backups/:id/apply", handler.ApplyBackup(appContainer.IsReady))
	statusV1.POST("/backups", handler.CreateBackup(appContainer.IsReady))
	statusV1.GET("/versions", handler.GetVersions)
	appMode := config.GetConfig().App.Mode
	if appMode == "debug" {
		statusV1.GET("/logs", handler.GetLogs(appContainer.Config.Log.Dir))
	}
}
