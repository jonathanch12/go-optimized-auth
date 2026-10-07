package router

import (
	"go-auth-api/internal/handler"
	"go-auth-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

func New(h *handler.AuthHandler) *gin.Engine {
	engine := gin.Default()
	api := engine.Group("/api/auth")
	{
		api.POST("/login", h.Login)
		api.GET("/me", middleware.RequireAuth(), h.Me)
	}
	return engine
}
