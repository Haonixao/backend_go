package middleware

import (
	"github.com/gin-gonic/gin"
)

func (mid *Container) IsReadyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !mid.IsReady.Load() {
			if c.Request.URL.Path == "/api/v1/status/logs" {
				c.Next()
				return
			}
			mid.HandlersContainer.PagesHandler.GetNotReadyPage(c)
			c.Abort()
			return
		}
		c.Next()
	}
}
