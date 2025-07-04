package auth

import (
	"context"
	"net/http"
	"strings"
)

// UserContextKey is the key used to store user ID in request context
type UserContextKey string

const (
	UserIDKey UserContextKey = "user_id"
)

// AuthMiddleware validates JWT tokens and adds user ID to request context
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error":"missing authorization header","success":false}`, http.StatusUnauthorized)
			return
		}

		// Check Bearer token format
		tokenParts := strings.SplitN(authHeader, " ", 2)
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			http.Error(w, `{"error":"invalid authorization header format","success":false}`, http.StatusUnauthorized)
			return
		}

		token := tokenParts[1]

		// Validate JWT token
		userID, err := ValidateJWT(token)
		if err != nil {
			http.Error(w, `{"error":"invalid or expired token","success":false}`, http.StatusUnauthorized)
			return
		}

		// Add user ID to request context
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserIDFromContext extracts user ID from request context
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

// RequireAuth is a helper function to check if user is authenticated in context
func RequireAuth(ctx context.Context) (string, error) {
	userID, ok := GetUserIDFromContext(ctx)
	if !ok || userID == "" {
		return "", &AuthError{Message: "authentication required"}
	}
	return userID, nil
}

// AuthError represents authentication-related errors
type AuthError struct {
	Message string
}

func (e *AuthError) Error() string {
	return e.Message
}
