package domain

import "time"

// Device represents a smartwatch registered to the platform.
type Device struct {
	ID         string
	IMEI       string // 15-digit unique identifier
	UserID     string // Owner of the device (Google OAuth ID)
	Name       string // Optional user-assigned name
	LastSeenAt *time.Time
	CreatedAt  time.Time
}
