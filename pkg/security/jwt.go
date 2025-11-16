package security

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("your-super-secret-jwt-key-change-this-in-production")

type JWTClaims struct {
	UserID   string `json:"user_id"`
	TgUserID int64  `json:"tg_user_id"`
	jwt.RegisteredClaims
}

// JWT Claims
type Claims struct {
	UserID   string `json:"user_id"`
	TgUserID int64  `json:"tg_user_id"`
}

// GenerateToken creates a JWT token for a user
func GenerateToken(userID string, tgUserID int64) (string, error) {
	expirationTime := time.Now().Add(30 * 24 * time.Hour) // 30 days

	claims := &JWTClaims{
		UserID:   userID,
		TgUserID: tgUserID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidateToken validates a JWT token and returns the claims
func ValidateToken(tokenString string) (*Claims, error) {
	claims := &JWTClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return &Claims{
		UserID:   claims.UserID,
		TgUserID: claims.TgUserID,
	}, nil
}
