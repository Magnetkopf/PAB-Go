package httpapi

import (
	"github.com/Magnetkopf/PAB-Go/internal/httpapi/handler"
	"github.com/gin-gonic/gin"
)

func Register(router *gin.Engine, store handler.Store, config handler.Config) {
	handler.New(store, config).Register(router.Group("/api"))
}
