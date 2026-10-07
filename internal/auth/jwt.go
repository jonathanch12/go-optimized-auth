package auth

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

type GeneratedToken struct {
	Token     string
	JTI       string
	TTL       time.Duration
	ExpiresAt time.Time
}

func jwtTTL() time.Duration {
	minutes := 60
	if v := os.Getenv("JWT_EXPIRES_MINUTES"); v != "" {
		if m, err := strconv.Atoi(v); err == nil && m > 0 {
			minutes = m
		}
	}
	return time.Duration(minutes) * time.Minute
}

func GenerateToken(userID int) (GeneratedToken, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return GeneratedToken{}, errors.New("JWT_SECRET not set")
	}

	ttl := jwtTTL()
	now := time.Now()
	expiresAt := now.Add(ttl)
	jti := uuid.NewString()

	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return GeneratedToken{}, err
	}

	return GeneratedToken{
		Token:     signed,
		JTI:       jti,
		TTL:       ttl,
		ExpiresAt: expiresAt,
	}, nil
}

func ParseToken(tokenString string) (*Claims, error) {
	secret := os.Getenv("JWT_SECRET")

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	if claims.ID == "" {
		return nil, errors.New("missing jti claim")
	}
	if claims.UserID == 0 {
		return nil, errors.New("missing user_id claim")
	}
	return claims, nil
}
