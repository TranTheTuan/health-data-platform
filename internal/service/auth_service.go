package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/TranTheTuan/health-data-platform/internal/auth"
	"github.com/TranTheTuan/health-data-platform/internal/dto"
)

type AuthService interface {
	ExchangeCodeForUser(ctx context.Context, code string) (dto.GoogleUserResponse, error)
}

type authService struct{}

func NewAuthService() AuthService {
	return &authService{}
}

func (s *authService) ExchangeCodeForUser(ctx context.Context, code string) (dto.GoogleUserResponse, error) {
	token, err := auth.GoogleOAuthConfig.Exchange(ctx, code)
	if err != nil {
		return dto.GoogleUserResponse{}, fmt.Errorf("exchange failed: %w", err)
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		return dto.GoogleUserResponse{}, fmt.Errorf("failed getting user info: %w", err)
	}
	defer response.Body.Close()

	contents, err := io.ReadAll(response.Body)
	if err != nil {
		return dto.GoogleUserResponse{}, fmt.Errorf("failed reading response: %w", err)
	}

	var user dto.GoogleUserResponse
	if err := json.Unmarshal(contents, &user); err != nil {
		return dto.GoogleUserResponse{}, fmt.Errorf("failed parsing user info: %w", err)
	}

	if user.ID == "" {
		return dto.GoogleUserResponse{}, errors.New("google ID missing")
	}

	return user, nil
}
