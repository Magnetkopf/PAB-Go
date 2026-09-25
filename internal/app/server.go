package app

import (
	"github.com/Magnetkopf/PAB-Go/internal/config"
	"github.com/Magnetkopf/PAB-Go/internal/httpapi"
	"github.com/Magnetkopf/PAB-Go/internal/httpapi/handler"
	"github.com/Magnetkopf/PAB-Go/internal/web"
	"github.com/gin-gonic/gin"
)

func NewServer(cfg config.Config, store handler.Store) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	httpapi.Register(router, store, handler.Config{Username: cfg.Username, PasswordHash: cfg.PasswordHash, SessionSecret: cfg.SessionSecret})
	web.Serve(router)
	return router
}
