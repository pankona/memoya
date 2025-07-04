package models

import (
	"time"
)

// User represents a user in the system with minimal privacy-focused data
type User struct {
	ID        string    `firestore:"id" json:"id"`               // Random UUID for user identification
	GoogleID  string    `firestore:"google_id" json:"google_id"` // Google OAuth user ID for authentication only
	CreatedAt time.Time `firestore:"created_at" json:"created_at"`
	IsActive  bool      `firestore:"is_active" json:"is_active"` // Account status
}

// DeviceAuthSession represents a temporary authentication session for OAuth Device Flow
type DeviceAuthSession struct {
	DeviceCode      string    `firestore:"device_code" json:"device_code"`
	UserCode        string    `firestore:"user_code" json:"user_code"`
	VerificationURI string    `firestore:"verification_uri" json:"verification_uri"`
	ExpiresAt       time.Time `firestore:"expires_at" json:"expires_at"`
	UserID          string    `firestore:"user_id,omitempty" json:"user_id,omitempty"` // Set after authorization
	Status          string    `firestore:"status" json:"status"`                       // "pending", "authorized", "expired"
	CreatedAt       time.Time `firestore:"created_at" json:"created_at"`
}

// DeviceAuthStatus constants
const (
	DeviceAuthStatusPending    = "pending"
	DeviceAuthStatusAuthorized = "authorized"
	DeviceAuthStatusExpired    = "expired"
)
