package httphandler

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/login/application"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

type LoginUseCaseInterface interface {
	Execute(ctx context.Context, input application.LoginDTO) (*application.LoginResult, error)
}

type LoginHTTPHandler struct {
	UseCase LoginUseCaseInterface
}

func NewLoginHTTPHandler(useCase LoginUseCaseInterface) *LoginHTTPHandler {
	return &LoginHTTPHandler{UseCase: useCase}
}

// LoginSpecialist godoc
// @Summary      Login as specialist
// @Description  Authenticate a specialist with email and password. Returns access and refresh tokens.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "Login payload"
// @Success      200 {object} LoginResponse
// @Failure      400 {object} ErrorResponse "Invalid payload"
// @Failure      401 {object} ErrorResponse "Invalid credentials or locked account"
// @Failure      500 {object} ErrorResponse "Infrastructure failure"
// @Router       /api/v1/auth/specialist/login [post]
func (h *LoginHTTPHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/auth/specialist/login", h.loginAs(role.Specialist))
	r.POST("/auth/patient/login", h.loginAs(role.Patient))
	r.POST("/auth/admin/login", h.loginAs(role.Admin))
}

func (h *LoginHTTPHandler) loginAs(expectedRole role.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}

		dto := application.LoginDTO{
			Email:        req.Email,
			Password:     req.Password,
			ExpectedRole: expectedRole.String(),
			DeviceInfo:   req.DeviceInfo,
			IPAddress:    c.ClientIP(),
			UserAgent:    c.Request.UserAgent(),
		}

		result, err := h.UseCase.Execute(c.Request.Context(), dto)
		if err != nil {
			status, msg := mapLoginError(err)
			c.JSON(status, ErrorResponse{Error: msg})
			return
		}

		c.JSON(http.StatusOK, ToLoginResponse(result))
	}
}

func mapLoginError(err error) (int, string) {
	switch {
	case errors.Is(err, application.ErrInvalidCredentials):
		return http.StatusUnauthorized, "invalid credentials"
	case errors.Is(err, application.ErrCredentialLocked):
		return http.StatusUnauthorized, "account locked"
	case errors.Is(err, application.ErrIssueTokens),
		errors.Is(err, application.ErrPersistSession),
		errors.Is(err, application.ErrCacheRefreshToken):
		return http.StatusInternalServerError, "internal error"
	default:
		return http.StatusInternalServerError, "internal error"
	}
}
