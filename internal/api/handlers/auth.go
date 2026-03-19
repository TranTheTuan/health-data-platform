package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/TranTheTuan/health-data-platform/configs"
	"github.com/TranTheTuan/health-data-platform/internal/auth"
	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	cfg *configs.Config
}

func NewAuthHandler(cfg *configs.Config) *AuthHandler {
	return &AuthHandler{cfg: cfg}
}

func generateStateOauthCookie(c echo.Context) string {
	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	cookie := new(http.Cookie)
	cookie.Name = "oauthstate"
	cookie.Value = state
	cookie.Expires = time.Now().Add(365 * 24 * time.Hour)
	cookie.HttpOnly = true
	cookie.Secure = false // Should be true in production with HTTPS
	cookie.Path = "/"
	c.SetCookie(cookie)

	return state
}

func (h *AuthHandler) Home(c echo.Context) error {
	return c.HTML(http.StatusOK, `<html><body><a href="/login">Login with Google</a></body></html>`)
}

func (h *AuthHandler) GoogleLogin(c echo.Context) error {
	state := generateStateOauthCookie(c)
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

	code := c.QueryParam("code")
	token, err := auth.GoogleOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Code exchange failed")
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed getting user info")
	}
	defer response.Body.Close()

	contents, err := io.ReadAll(response.Body)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed reading response")
	}

	var user map[string]interface{}
	if err := json.Unmarshal(contents, &user); err != nil {
		return c.String(http.StatusInternalServerError, "Failed parsing user info")
	}

	// Set a signed session cookie
	signedValue := auth.Sign(fmt.Sprintf("%v", user["id"]), h.cfg.SessionSecret)
	sessionCookie := new(http.Cookie)
	sessionCookie.Name = "session"
	sessionCookie.Value = signedValue
	sessionCookie.Expires = time.Now().Add(24 * time.Hour)
	sessionCookie.HttpOnly = true
	sessionCookie.Path = "/"
	c.SetCookie(sessionCookie)

	return c.Redirect(http.StatusTemporaryRedirect, "/protected")
}

func (h *AuthHandler) AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sessionCookie, err := c.Cookie("session")
		if err != nil {
			return c.String(http.StatusUnauthorized, "Unauthorized: Please login")
		}

		// Verify the signature
		userID, err := auth.Verify(sessionCookie.Value, h.cfg.SessionSecret)
		if err != nil {
			return c.String(http.StatusUnauthorized, "Unauthorized: Invalid session")
		}

		c.Set("user_id", userID)
		return next(c)
	}
}

func (h *AuthHandler) ProtectedEndpoint(c echo.Context) error {
	userID := c.Get("user_id").(string)
	return c.String(http.StatusOK, fmt.Sprintf("Welcome to the protected route! User ID: %s", userID))
}
