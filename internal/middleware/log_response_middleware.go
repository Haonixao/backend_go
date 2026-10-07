package middleware

import (
	"bytes"
	"io"
	"net/http"

	"backend_go/internal/config"
	"backend_go/pkg/log"

	"github.com/gin-gonic/gin"
)

func (mid *Container) LogResponseMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		body, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		query := c.Request.URL.RawQuery
		path := c.Request.URL.Path
		c.Next()
		if c.Writer.Status() >= http.StatusBadRequest {
			bodyStr := string(body)
			logger := log.GetLogger(config.GetConfig().Log)
			errEvent := logger.Error()
			if len(bodyStr) != 0 {
				errEvent = errEvent.Interface("body", bodyStr)
			}
			if len(query) != 0 {
				errEvent = errEvent.Str("query", query)
			}
			errEvent.Str("path", path).Msg("error request")
		}
	}
}
