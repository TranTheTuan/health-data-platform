package http

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/TranTheTuan/health-data-platform/configs"
	"github.com/TranTheTuan/health-data-platform/internal/auth"
	"github.com/TranTheTuan/health-data-platform/internal/dto"
	"github.com/TranTheTuan/health-data-platform/internal/service"
)

// AuthHandler coordinates OAuth logins with the service.
type AuthHandler struct {
	cfg *configs.Config
	svc service.AuthService
}

func NewAuthHandler(cfg *configs.Config, svc service.AuthService) *AuthHandler {
	return &AuthHandler{cfg: cfg, svc: svc}
}

func (h *AuthHandler) Home(c echo.Context) error {
	return c.Render(http.StatusOK, "login.html", nil)
}

func (h *AuthHandler) GoogleLogin(c echo.Context) error {
	// Simple static state for now based on original logic, but ideally use random generator
	state := "randomstate" 
	cookie := new(http.Cookie)
	cookie.Name = "oauthstate"
	cookie.Value = state
	cookie.Expires = time.Now().Add(365 * 24 * time.Hour)
	cookie.HttpOnly = true
	cookie.Path = "/"
	c.SetCookie(cookie)

	url := auth.GoogleOAuthConfig.AuthCodeURL(state)
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *AuthHandler) GoogleCallback(c echo.Context) error {
	oauthState, err := c.Cookie("oauthstate")
	if err != nil {
		return c.String(http.StatusBadRequest, "Missing oauth state cookie")
	}

	if c.QueryParam("state") != oauthState.Value {
		return c.String(http.StatusBadRequest, "Invalid oauth state")
	}

	userDto, err := h.svc.ExchangeCodeForUser(c.Request().Context(), c.QueryParam("code"))
	if err != nil {
		return c.String(http.StatusInternalServerError, "Auth failed: "+err.Error())
	}

	cookieVal := fmt.Sprintf("%s|%s", userDto.ID, userDto.Email)
	signedValue := auth.Sign(cookieVal, h.cfg.SessionSecret)

	sessionCookie := new(http.Cookie)
	sessionCookie.Name = "session"
	sessionCookie.Value = signedValue
	sessionCookie.Expires = time.Now().Add(24 * time.Hour)
	sessionCookie.HttpOnly = true
	sessionCookie.Path = "/"
	c.SetCookie(sessionCookie)

	return c.Redirect(http.StatusTemporaryRedirect, "/dashboard")
}

func (h *AuthHandler) Dashboard(c echo.Context) error {
	email := ""
	if e, ok := c.Get("user_email").(string); ok {
		email = e
	}
	return c.Render(http.StatusOK, "dashboard.html", dto.DashboardData{
		UserID: c.Get("user_id").(string),
		Email:  email,
	})
}

func (h *AuthHandler) Logout(c echo.Context) error {
	cookie := new(http.Cookie)
	cookie.Name = "session"
	cookie.Value = ""
	cookie.Expires = time.Unix(0, 0)
	cookie.MaxAge = -1
	cookie.HttpOnly = true
	cookie.Path = "/"
	c.SetCookie(cookie)
	return c.Redirect(http.StatusTemporaryRedirect, "/")
}

// Keep existing middleware attached to handler
func (h *AuthHandler) AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sessionCookie, err := c.Cookie("session")
		if err != nil {
			return c.String(http.StatusUnauthorized, "Unauthorized: Please login")
		}

		raw, err := auth.Verify(sessionCookie.Value, h.cfg.SessionSecret)
		if err != nil {
			return c.String(http.StatusUnauthorized, "Unauthorized: Invalid session")
		}

		parts := strings.SplitN(raw, "|", 2)
		c.Set("user_id", parts[0])
		if len(parts) > 1 {
			c.Set("user_email", parts[1])
		}
		return next(c)
	}
}
