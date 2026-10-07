package model

import "time"

type AccessToken struct {
	ID        string     `json:"id"`
	UserID    int        `json:"user_id"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	ExpiresAt time.Time  `json:"expires_at"`
}
