package http

import (
	"github.com/labstack/echo/v4"

	http_handler "github.com/TranTheTuan/health-data-platform/internal/handler/http"
)

// RegisterRoutes wires all HTTP handlers to the Echo instance.
func RegisterRoutes(e *echo.Echo, ah *http_handler.AuthHandler, dh *http_handler.DeviceHandler) {
	e.GET("/", ah.Home)
	e.GET("/login", ah.GoogleLogin)
	e.GET("/auth/google/callback", ah.GoogleCallback)
	e.GET("/logout", ah.Logout)

	dashboardRoute := e.Group("/dashboard")
	dashboardRoute.Use(ah.AuthMiddleware)
	dashboardRoute.GET("", ah.Dashboard)

	protected := e.Group("/protected")
	protected.Use(ah.AuthMiddleware)

	protected.POST("/devices", dh.Register)
	protected.GET("/devices", dh.List)

	dashboardRoute.GET("/devices/:id/packets", dh.PacketInspectPage)
	protected.GET("/devices/:id/packets", dh.ListPacketsAPI)
}
