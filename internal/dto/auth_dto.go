package dto

// GoogleUserResponse represents the data we fetch from Google OAuth UserInfo API.
type GoogleUserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// DashboardData is the context passed to the dashboard HTML template.
type DashboardData struct {
	UserID string
	Email  string
}
