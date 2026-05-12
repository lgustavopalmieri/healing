package httphandler

import (
	"time"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/login/application"
)

const bearerTokenType = "Bearer"

type LoginRequest struct {
	Email      string `json:"email" binding:"required"`
	Password   string `json:"password" binding:"required"`
	DeviceInfo string `json:"device_info"`
}

type TokenPairResponse struct {
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	TokenType        string    `json:"token_type"`
	AccessExpiresAt  time.Time `json:"access_expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
}

type LoginResponse struct {
	TokenPair TokenPairResponse `json:"token_pair"`
	SubjectID string            `json:"subject_id"`
	Role      string            `json:"role"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func ToLoginResponse(result *application.LoginResult) LoginResponse {
	return LoginResponse{
		TokenPair: TokenPairResponse{
			AccessToken:      result.TokenPair.AccessToken,
			RefreshToken:     result.TokenPair.RefreshToken,
			TokenType:        bearerTokenType,
			AccessExpiresAt:  result.TokenPair.AccessExpiresAt,
			RefreshExpiresAt: result.TokenPair.RefreshExpiresAt,
		},
		SubjectID: result.SubjectID,
		Role:      result.Role.String(),
	}
}
