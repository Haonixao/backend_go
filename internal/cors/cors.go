package cors

import (
	"net/http"
	"strings"

	"backend_go/internal/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func GetCors() gin.HandlerFunc {
	allowedOrigins := strings.Split(config.GetConfig().Cors.AllowedOrigins, ",")
	return cors.New(cors.Config{
		AllowOrigins: allowedOrigins,
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut,
			http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	})
}
