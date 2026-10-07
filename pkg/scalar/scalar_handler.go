package scalar

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

const scalarHTMLTemplate = `<!doctype html>
<html>
<head>
    <title>%s</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
</head>
<body>
    <div id="app"></div>
    <script src="%s"></script>
    <script>
        Scalar.createApiReference('#app', {
            url: '%s',
        })
    </script>
</body>
</html>`

// ScalarUIHandler returns gin handlers for serving Scalar API UI offline.
//
// Скрипт Scalar скачивается вручную с https://cdn.jsdelivr.net/npm/@scalar/api-reference
// и сохраняется как ./docs/scalar.js Для уменьшения размера файл сжимается gzip
// и сохраняется как ./docs/scalar.js.gz
// curl -o scalar.js https://cdn.jsdelivr.net/npm/@scalar/api-reference && gzip scalar.js
// specData и scriptGzip встраиваются в бинарник через //go:embed в docs/static.go
//
// Routes registered:
//
// GET /docs/*any         -> HTML page
// GET /docs/openapi.json -> OpenAPI spec (embedded)
// GET /scalar.js         -> Scalar script gzip-compressed (embedded)
func ScalarUIHandler(group *gin.RouterGroup, title string, specData, scriptGzip []byte, specURL, scriptURL string) {
	html := fmt.Sprintf(scalarHTMLTemplate, title, scriptURL, specURL)

	group.GET("/docs/*any", func(c *gin.Context) {
		if c.Param("any") == "/openapi.json" {
			c.Data(http.StatusOK, "application/json; charset=utf-8", specData)
			return
		}
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, html)
	})

	group.GET("/scalar.js", func(c *gin.Context) {
		c.Header("Content-Encoding", "gzip")
		c.Data(http.StatusOK, "application/javascript; charset=utf-8", scriptGzip)
	})
}
