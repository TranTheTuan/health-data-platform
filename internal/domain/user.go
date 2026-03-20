package domain

// User represents an authenticated owner of devices.
type User struct {
	ID    string // Google OAuth Unique ID
	Email string
}
