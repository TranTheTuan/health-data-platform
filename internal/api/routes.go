package api

import (
	"github.com/labstack/echo/v4"

	"github.com/TranTheTuan/health-data-platform/configs"
	"github.com/TranTheTuan/health-data-platform/internal/api/handlers"
)

func RegisterRoutes(e *echo.Echo, cfg *configs.Config) {
	// Initialize handlers
	ah := handlers.NewAuthHandler(cfg)

	e.GET("/", ah.Home)
	e.GET("/login", ah.GoogleLogin)
	e.GET("/auth/google/callback", ah.GoogleCallback)

	// Protected routes
	protected := e.Group("/protected")
	protected.Use(ah.AuthMiddleware)
	protected.GET("", ah.ProtectedEndpoint)
}
