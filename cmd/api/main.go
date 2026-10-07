package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"go-auth-api/internal/handler"
	"go-auth-api/internal/router"
	"go-auth-api/internal/store"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("Warning: .env file not found, error: %v", err)
	}

	redisClient := newRedisClient()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	sessions := store.NewRedisSessionStore(redisClient)

	s := store.NewUserStore()
	authH := handler.NewAuthHandler(s, sessions)
	r := router.New(authH, sessions)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func newRedisClient() *redis.Client {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	db := 0
	if v := os.Getenv("REDIS_DB"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			db = n
		}
	}

	return redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       db,
	})
}
