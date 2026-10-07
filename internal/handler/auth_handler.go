package handler

import (
	"go-auth-api/internal/auth"
	"go-auth-api/internal/store"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	store *store.UserStore
}

func NewAuthHandler(s *store.UserStore) *AuthHandler {
	return &AuthHandler{store: s}
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	user, ok := h.store.FindByUsername(req.Username)
	if !ok {
		c.JSON(401, gin.H{"error": "Invalid credentials"})
		return
	}

	if !auth.CheckPassword(user.Password, req.Password) {
		c.JSON(401, gin.H{"error": "Invalid credentials"})
		return
	}

	tokenString, err := auth.GenerateToken(user.ID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Server error"})
		return
	}

	c.JSON(200, gin.H{"token": tokenString})
}

func (h *AuthHandler) Me(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}
	userID, ok := val.(int)
	if !ok {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}
	user, found := h.store.FindByID(userID)
	if !found {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}
	c.JSON(200, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"name":     user.Name,
	})
}
