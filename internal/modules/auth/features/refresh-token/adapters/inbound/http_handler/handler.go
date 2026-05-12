package httphandler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/refresh-token/application"
)

type RefreshTokenUseCaseInterface interface {
	Execute(ctx context.Context, input application.RefreshTokenDTO) (*application.RefreshTokenResult, error)
}

type RefreshTokenHTTPHandler struct {
	UseCase RefreshTokenUseCaseInterface
}

func NewRefreshTokenHTTPHandler(useCase RefreshTokenUseCaseInterface) *RefreshTokenHTTPHandler {
	return &RefreshTokenHTTPHandler{UseCase: useCase}
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type TokenPairResponse struct {
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	TokenType        string    `json:"token_type"`
	AccessExpiresAt  time.Time `json:"access_expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
}

type RefreshTokenResponse struct {
	TokenPair TokenPairResponse `json:"token_pair"`
	SubjectID string            `json:"subject_id"`
	Role      string            `json:"role"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// RefreshToken godoc
// @Summary      Refresh access token
// @Description  Exchange a valid refresh token for a new access/refresh token pair. The old refresh token is invalidated.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body RefreshTokenRequest true "Refresh token payload"
// @Success      200 {object} RefreshTokenResponse
// @Failure      400 {object} ErrorResponse "Invalid payload"
// @Failure      401 {object} ErrorResponse "Invalid or expired refresh token"
// @Failure      500 {object} ErrorResponse "Infrastructure failure"
// @Router       /api/v1/auth/refresh [post]
func (h *RefreshTokenHTTPHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/auth/refresh", h.handle)
}

func (h *RefreshTokenHTTPHandler) handle(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	result, err := h.UseCase.Execute(c.Request.Context(), application.RefreshTokenDTO{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		status, msg := mapRefreshError(err)
		c.JSON(status, ErrorResponse{Error: msg})
		return
	}

	c.JSON(http.StatusOK, RefreshTokenResponse{
		TokenPair: TokenPairResponse{
			AccessToken:      result.TokenPair.AccessToken,
			RefreshToken:     result.TokenPair.RefreshToken,
			TokenType:        "Bearer",
			AccessExpiresAt:  result.TokenPair.AccessExpiresAt,
			RefreshExpiresAt: result.TokenPair.RefreshExpiresAt,
		},
		SubjectID: result.SubjectID,
		Role:      result.Role.String(),
	})
}

func mapRefreshError(err error) (int, string) {
	switch {
	case errors.Is(err, application.ErrInvalidRefreshToken):
		return http.StatusUnauthorized, "invalid refresh token"
	case errors.Is(err, application.ErrDeleteOldRefresh),
		errors.Is(err, application.ErrIssueNewTokens),
		errors.Is(err, application.ErrCacheNewRefresh):
		return http.StatusInternalServerError, "internal error"
	default:
		return http.StatusInternalServerError, "internal error"
	}
}
