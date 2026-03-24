package domain

import "time"

// Device represents a smartwatch registered to the platform.
type Device struct {
	ID         string
	IMEI       string // 10-15 digit device identifier (10-digit Wonlex ID or legacy 15-digit IMEI)
	UserID     string // Owner of the device (Google OAuth ID)
	Name       string // Optional user-assigned name
	LastSeenAt *time.Time
	CreatedAt  time.Time
}
