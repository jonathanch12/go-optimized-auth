package handler

import (
	"log"
	"net/http"

	"go-auth-api/internal/auth"
	"go-auth-api/internal/middleware"
	"go-auth-api/internal/store"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	store    *store.UserStore
	sessions store.SessionStore
}

func NewAuthHandler(s *store.UserStore, sessions store.SessionStore) *AuthHandler {
	return &AuthHandler{store: s, sessions: sessions}
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	user, ok := h.store.FindByUsername(req.Username)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if !auth.CheckPassword(user.Password, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Credentials verified: mint the JWT (with jti) first, then record the
	// active session in Redis with a matching TTL.
	gen, err := auth.GenerateToken(user.ID)
	if err != nil {
		log.Printf("login: generate token for user %d: %v", user.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	if err := h.sessions.Create(c.Request.Context(), gen.JTI, user.ID, gen.TTL); err != nil {
		log.Printf("login: create session for user %d: %v", user.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": gen.Token})
}

func (h *AuthHandler) Me(c *gin.Context) {
	val, exists := c.Get(middleware.ContextUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, ok := val.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user, found := h.store.FindByID(userID)
	if !found {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"name":     user.Name,
	})
}

// Logout revokes the current session by deleting it from the session store.
// If the session is already gone (prior logout or TTL expiry),
// it still returns 204. Only a store/connectivity failure yields 500,
// because then we cannot confirm the session was revoked.
func (h *AuthHandler) Logout(c *gin.Context) {
	val, exists := c.Get(middleware.ContextJTI)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	jti, ok := val.(string)
	if !ok || jti == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if err := h.sessions.Delete(c.Request.Context(), jti); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not revoke session"})
		return
	}

	c.Status(http.StatusNoContent)
}
