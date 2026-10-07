package main

import (
	"go-auth-api/internal/handler"
	"go-auth-api/internal/router"
	"go-auth-api/internal/store"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("Warning: .env file not found, error: %v", err)
	}

	s := store.NewUserStore()
	authH := handler.NewAuthHandler(s)
	r := router.New(authH)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
