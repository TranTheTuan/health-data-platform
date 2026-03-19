package api

import (
	"database/sql"

	"github.com/labstack/echo/v4"

	"github.com/TranTheTuan/health-data-platform/configs"
	"github.com/TranTheTuan/health-data-platform/internal/api/handlers"
)

// RegisterRoutes wires all HTTP handlers to the Echo instance.
// db is the shared PostgreSQL pool used by handlers that need storage (e.g. device API).
func RegisterRoutes(e *echo.Echo, cfg *configs.Config, db *sql.DB) {
	// Auth handler (no DB needed — session cookie only)
	ah := handlers.NewAuthHandler(cfg)

	e.GET("/", ah.Home)
	e.GET("/login", ah.GoogleLogin)
	e.GET("/auth/google/callback", ah.GoogleCallback)

	// Protected routes (require valid session cookie)
	protected := e.Group("/protected")
	protected.Use(ah.AuthMiddleware)
	protected.GET("", ah.ProtectedEndpoint)

	// Device API — Phase 5
	dh := handlers.NewDeviceHandler(db)
	protected.POST("/devices", dh.Register)
	protected.GET("/devices", dh.List)
}
