package v1

import (
	"backend_go/internal/app_container"
	"backend_go/internal/routes/api/v1/remotedb"
	"backend_go/internal/routes/api/v1/status"
	"backend_go/internal/routes/api/v1/tasks"

	"github.com/gin-gonic/gin"
)

func Init(baseGroup *gin.RouterGroup, appContainer *app_container.Container) {
	v1 := baseGroup.Group("/v1")

	v1.Use(
		appContainer.MiddlewareContainer.LogResponseMiddleware(),
	)

	status.Init(v1, appContainer)
	remotedb.Init(v1, appContainer)
	tasks.Init(v1, appContainer)
}
