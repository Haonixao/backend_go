package api

import (
	"backend_go/internal/app_container"
	v1 "backend_go/internal/routes/api/v1"

	"github.com/gin-gonic/gin"
)

func Init(baseGroup *gin.RouterGroup, appContainer *app_container.Container) {
	v1.Init(baseGroup, appContainer)
}
