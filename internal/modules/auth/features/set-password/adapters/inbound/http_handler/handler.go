package httphandler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/password"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/set-password/application"
)

type SetPasswordUseCaseInterface interface {
	Execute(ctx context.Context, input application.SetPasswordDTO) (*application.SetPasswordResult, error)
}

type SetPasswordHTTPHandler struct {
	UseCase SetPasswordUseCaseInterface
}

func NewSetPasswordHTTPHandler(useCase SetPasswordUseCaseInterface) *SetPasswordHTTPHandler {
	return &SetPasswordHTTPHandler{UseCase: useCase}
}

// SetPassword godoc
// @Summary      Set initial password for a pending credential
// @Description  Consumes a single-use set-password token, activates the credential, and returns access/refresh tokens. The token is issued during onboarding and can be used only once.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body SetPasswordRequest true "Set-password payload"
// @Success      200 {object} SetPasswordResponse
// @Failure      400 {object} ErrorResponse "Invalid payload or weak password"
// @Failure      401 {object} ErrorResponse "Invalid or already used token"
// @Failure      409 {object} ErrorResponse "Credential not in pending state"
// @Failure      500 {object} ErrorResponse "Infrastructure failure"
// @Router       /api/v1/auth/set-password [post]
func (h *SetPasswordHTTPHandler) SetPassword(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	var req SetPasswordRequest
	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	dto := application.SetPasswordDTO{
		Token:      req.Token,
		Password:   req.Password,
		DeviceInfo: req.DeviceInfo,
		IPAddress:  c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	}

	result, err := h.UseCase.Execute(c.Request.Context(), dto)
	if err != nil {
		status, msg := mapSetPasswordError(err)
		c.JSON(status, ErrorResponse{Error: msg})
		return
	}

	c.JSON(http.StatusOK, ToSetPasswordResponse(result))
}

func (h *SetPasswordHTTPHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/auth/set-password", h.SetPassword)
}

func mapSetPasswordError(err error) (int, string) {
	switch {
	case errors.Is(err, application.ErrInvalidSetPasswordToken),
		errors.Is(err, application.ErrSingleUseTokenAlreadyUsed):
		return http.StatusUnauthorized, err.Error()
	case errors.Is(err, application.ErrCredentialNotFound):
		return http.StatusNotFound, err.Error()
	case errors.Is(err, application.ErrCredentialNotPending):
		return http.StatusConflict, err.Error()
	case errors.Is(err, password.ErrTooShort),
		errors.Is(err, password.ErrMissingRequiredChars):
		return http.StatusBadRequest, err.Error()
	case errors.Is(err, application.ErrFailedToConsumeSingleUse),
		errors.Is(err, application.ErrFailedToFindCredential),
		errors.Is(err, application.ErrFailedToHashPassword),
		errors.Is(err, application.ErrFailedToActivateCredential),
		errors.Is(err, application.ErrFailedToPersistCredential),
		errors.Is(err, application.ErrFailedToIssueTokenPair),
		errors.Is(err, application.ErrFailedToPersistSession),
		errors.Is(err, application.ErrFailedToCacheRefreshToken):
		return http.StatusInternalServerError, err.Error()
	default:
		return http.StatusInternalServerError, err.Error()
	}
}
