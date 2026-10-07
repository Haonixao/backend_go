package tasks

import (
	"backend_go/internal/app_container"

	"github.com/gin-gonic/gin"
)

func Init(baseGroup *gin.RouterGroup, appContainer *app_container.Container) {
	tasksV1 := baseGroup.Group("/tasks")

	handler := appContainer.HandlersContainer.TasksHandler

	tasksV1.GET("/:id", handler.GetTaskById)
	tasksV1.GET("/:id/logs", handler.GetTaskLogs)
	tasksV1.PATCH("/:id", handler.UpdateTaskById(appContainer.Scheduler))
	tasksV1.GET("", handler.GetTasks)
}
