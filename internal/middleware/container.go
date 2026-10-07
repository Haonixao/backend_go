package middleware

import (
	"sync/atomic"

	"backend_go/internal/handlers"
	"backend_go/internal/services"
)

type Container struct {
	ServicesContainer *services.Container
	HandlersContainer *handlers.Container
	IsReady *atomic.Bool
}
