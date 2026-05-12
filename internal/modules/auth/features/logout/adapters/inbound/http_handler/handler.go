package httphandler

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/logout/application"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/claims"
)

type LogoutUseCaseInterface interface {
	Execute(ctx context.Context, input application.LogoutDTO) error
}

type LogoutHTTPHandler struct {
	UseCase LogoutUseCaseInterface
}

func NewLogoutHTTPHandler(useCase LogoutUseCaseInterface) *LogoutHTTPHandler {
	return &LogoutHTTPHandler{UseCase: useCase}
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// Logout godoc
// @Summary      Logout and revoke session
// @Description  Invalidates the refresh token and blacklists the current access token. Requires authentication.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body LogoutRequest true "Logout payload"
// @Success      204 "No Content"
// @Failure      400 {object} ErrorResponse "Invalid payload"
// @Failure      401 {object} ErrorResponse "Unauthenticated"
// @Failure      500 {object} ErrorResponse "Infrastructure failure"
// @Router       /api/v1/auth/logout [post]
func (h *LogoutHTTPHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/auth/logout", h.handle)
}

func (h *LogoutHTTPHandler) handle(c *gin.Context) {
	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	userClaims, ok := claims.FromContext(c.Request.Context())
	if !ok || userClaims == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthenticated"})
		return
	}

	dto := application.LogoutDTO{
		RefreshToken:   req.RefreshToken,
		AccessTokenJTI: userClaims.TokenID,
		AccessTokenExp: userClaims.ExpireAt,
		SubjectID:      userClaims.Subject,
		Role:           userClaims.Role.String(),
		IPAddress:      c.ClientIP(),
		UserAgent:      c.Request.UserAgent(),
	}

	if err := h.UseCase.Execute(c.Request.Context(), dto); err != nil {
		status, msg := mapLogoutError(err)
		c.JSON(status, ErrorResponse{Error: msg})
		return
	}

	c.Status(http.StatusNoContent)
}

func mapLogoutError(err error) (int, string) {
	switch {
	case errors.Is(err, application.ErrDeleteRefreshToken),
		errors.Is(err, application.ErrBlacklistAccessToken):
		return http.StatusInternalServerError, "internal error"
	default:
		return http.StatusInternalServerError, "internal error"
	}
}
