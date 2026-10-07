package router

import (
	"go-auth-api/internal/handler"
	"go-auth-api/internal/middleware"
	"go-auth-api/internal/store"

	"github.com/gin-gonic/gin"
)

func New(h *handler.AuthHandler, sessions store.SessionStore) *gin.Engine {
	engine := gin.Default()
	api := engine.Group("/api/auth")
	{
		api.POST("/login", h.Login)
		api.GET("/me", middleware.RequireAuth(sessions), h.Me)
		api.POST("/logout", middleware.RequireAuth(sessions), h.Logout)
	}
	return engine
}
