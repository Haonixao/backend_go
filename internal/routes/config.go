package routes

import (
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	"backend_go/docs"
	"backend_go/internal/app_container"
	"backend_go/internal/config"
	"backend_go/internal/routes/api"
	"backend_go/pkg/scalar"

	"github.com/gin-gonic/gin"
)

func Init(engine *gin.Engine, appContainer *app_container.Container) {
	// Создание валидатора который при ошибке возвращает не имя поля, а значение из тега
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "" {
				name = strings.SplitN(fld.Tag.Get("form"), ",", 2)[0]
			}
			if name == "-" {
				return ""
			}
			return name
		})
	}
	baseGroup := engine.Group("")
	baseGroup.Use(appContainer.MiddlewareContainer.IsReadyMiddleware())
	baseGroup.GET("/", appContainer.HandlersContainer.PagesHandler.StartPage)
	apiGroup := baseGroup.Group("/api")
	api.Init(apiGroup, appContainer)
	if config.GetConfig().App.Mode == "debug" {
		// Scalar
		scalarGroup := apiGroup.Group("/scalar")
		scalar.ScalarUIHandler(scalarGroup, "Backend API",
			docs.ScalarSwaggerJSON,
			docs.ScalarScriptGzip,
			"/api/scalar/docs/openapi.json",
			"/api/scalar/scalar.js",
		)
	}
}
