package pages

import (
	"net/http"

	"backend_go/internal/config"

	"github.com/gin-gonic/gin"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) GetNotReadyPage(c *gin.Context) {
	htmlContent := `
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Загрузка приложения...</title>
    <link href="https://fonts.googleapis.com/css2?family=Roboto:wght@400;700&display=swap" rel="stylesheet">
    <style>
        body {
            font-family: 'Roboto', sans-serif;
            background-color: #f4f6f9;
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
            margin: 0;
            color: #333;
            flex-direction: column;
        }
        .container {
            background: #fff;
            padding: 2rem;
            border-radius: 8px;
            box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
            max-width: 600px;
            width: 100%;
            text-align: center;
        }
        h1 {
            font-size: 1.5rem;
            margin-bottom: 1rem;
            color: #ff9800;
        }
        .status {
            font-style: italic;
            color: #555;
        }
        .loader {
            border: 4px solid #f3f3f3;
            border-top: 4px solid #ff9800;
            border-radius: 50%;
            width: 40px;
            height: 40px;
            animation: spin 1s linear infinite;
            margin: 1rem auto;
        }
        @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
        }
        .logs {
            background: #000;
            color: #fff;
            font-family: monospace;
            text-align: left;
            padding: 1rem;
            margin-top: 1.5rem;
            height: 200px;
            overflow-y: auto;
            border-radius: 4px;
            font-size: 0.9rem;
            display: none; /* скрыто по умолчанию */
        }
        .toggle-btn {
            margin-top: 1rem;
            padding: 0.5rem 1rem;
            font-size: 0.9rem;
            border: none;
            border-radius: 4px;
            background-color: #1a73e8;
            color: #fff;
            cursor: pointer;
            transition: background-color 0.2s ease;
        }
        .toggle-btn:hover {
            background-color: #155ab6;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Загрузка приложения..</h1>
        <div class="loader"></div>
        <p class="status">Сервис запущен, но ещё выполняет необходимые операции...</p>
        <p>Пожалуйста, попробуйте снова через некоторое время</p>`

	if config.GetConfig().App.Mode == "debug" {
		htmlContent += `<p><a href="/api/v1/status/logs" class="toggle-btn" style="text-decoration: none; display: inline-block;">Скачать логи</a></p>`
	}
	htmlContent += `
    </div>
</body>
</html>
`
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, htmlContent)
}

func (h *Handler) StartPage(c *gin.Context) {
	htmlContent := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Backend Go</title>
</head>
<body>
    <h1>Backend Go</h1>
    <p>Scalar: http://localhost:8080/api/scalar/docs</p>
</body>
</html>
`
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, htmlContent)
}
