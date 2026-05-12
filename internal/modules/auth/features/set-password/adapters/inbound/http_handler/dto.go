package httphandler

import (
	"time"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/set-password/application"
)

const bearerTokenType = "Bearer"

type SetPasswordRequest struct {
	Token      string `json:"token"`
	Password   string `json:"password"`
	DeviceInfo string `json:"device_info"`
}

type TokenPairResponse struct {
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	TokenType        string    `json:"token_type"`
	AccessExpiresAt  time.Time `json:"access_expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
}

type SetPasswordResponse struct {
	TokenPair TokenPairResponse `json:"token_pair"`
	SubjectID string            `json:"subject_id"`
	Role      string            `json:"role"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func ToSetPasswordResponse(result *application.SetPasswordResult) SetPasswordResponse {
	return SetPasswordResponse{
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
