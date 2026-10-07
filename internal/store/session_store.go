package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrSessionNotFound = errors.New("session not found")

const sessionKeyPrefix = "auth:session:"

type SessionStore interface {
	Create(ctx context.Context, tokenID string, userID int, expiration time.Duration) error
	GetUserID(ctx context.Context, tokenID string) (int, error)
	Delete(ctx context.Context, tokenID string) error
}

type RedisSessionStore struct {
	client *redis.Client
}

func NewRedisSessionStore(client *redis.Client) *RedisSessionStore {
	return &RedisSessionStore{client: client}
}

func sessionKey(tokenID string) string {
	return sessionKeyPrefix + tokenID
}

func (s *RedisSessionStore) Create(ctx context.Context, tokenID string, userID int, expiration time.Duration) error {
	if err := s.client.Set(ctx, sessionKey(tokenID), userID, expiration).Err(); err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

func (s *RedisSessionStore) GetUserID(ctx context.Context, tokenID string) (int, error) {
	userID, err := s.client.Get(ctx, sessionKey(tokenID)).Int()
	if errors.Is(err, redis.Nil) {
		return 0, ErrSessionNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("get session: %w", err)
	}
	return userID, nil
}

func (s *RedisSessionStore) Delete(ctx context.Context, tokenID string) error {
	if err := s.client.Del(ctx, sessionKey(tokenID)).Err(); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}
