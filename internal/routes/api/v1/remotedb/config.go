package remotedb

import (
	"backend_go/internal/app_container"
	"backend_go/internal/config"
	"backend_go/pkg/http_tcp_connector"

	"github.com/gin-gonic/gin"
)

func Init(baseGroup *gin.RouterGroup, appContainer *app_container.Container) {
	appMode := config.GetConfig().App.Mode
	if appMode != "debug" {
		return
	}
	remotedb := baseGroup.Group("/remotedb")

	http_tcp_connector.RemoteDBAddr = appContainer.Config.Host + ":" + appContainer.Config.Port
	http_tcp_connector.Metadata["user"] = appContainer.Config.User
	http_tcp_connector.Metadata["password"] = appContainer.Config.Password
	http_tcp_connector.Metadata["db"] = appContainer.Config.Db

	remotedb.POST("/send_data", http_tcp_connector.HandleSendDataGin)
	remotedb.POST("/get_data", http_tcp_connector.HandleGetDataGin)
	remotedb.GET("/get_meta", http_tcp_connector.HandleGetMetaGin)
}
