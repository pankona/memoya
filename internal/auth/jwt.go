package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	jwtSecret []byte
)

func init() {
	// Get JWT secret from environment or generate one
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// Generate a random secret for development
		randomBytes := make([]byte, 32)
		rand.Read(randomBytes)
		secret = base64.URLEncoding.EncodeToString(randomBytes)
	}
	jwtSecret = []byte(secret)
}

// Claims represents JWT claims with minimal user information
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateJWT creates a JWT token for the given user ID
func GenerateJWT(userID string) (string, error) {
	// Set expiration to 24 hours
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "memoya",
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT token: %w", err)
	}

	return tokenString, nil
}

// ValidateJWT validates a JWT token and returns the user ID
func ValidateJWT(tokenString string) (string, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to parse JWT token: %w", err)
	}

	if !token.Valid {
		return "", fmt.Errorf("invalid JWT token")
	}

	return claims.UserID, nil
}

// RefreshJWT generates a new JWT token with extended expiration
func RefreshJWT(tokenString string) (string, error) {
	userID, err := ValidateJWT(tokenString)
	if err != nil {
		return "", fmt.Errorf("cannot refresh invalid token: %w", err)
	}

	return GenerateJWT(userID)
}
