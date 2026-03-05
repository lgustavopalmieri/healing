package httphandler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/application"
)

type SpecialistCreateUseCaseInterface interface {
	Execute(ctx context.Context, input application.CreateSpecialistDTO) (*domain.Specialist, error)
}

type SpecialistCreateHTTPHandler struct {
	UseCase SpecialistCreateUseCaseInterface
}

func NewSpecialistCreateHTTPHandler(useCase SpecialistCreateUseCaseInterface) *SpecialistCreateHTTPHandler {
	return &SpecialistCreateHTTPHandler{
		UseCase: useCase,
	}
}

// CreateSpecialist godoc
// @Summary      Create a new specialist
// @Description  Registers a new healthcare specialist on the platform. The specialist starts with pending status and goes through license validation.
// @Tags         specialists
// @Accept       json
// @Produce      json
// @Param        request body     CreateSpecialistRequest true "Specialist data"
// @Success      201     {object} CreateSpecialistSuccessResponse
// @Failure      400     {object} ErrorResponse "Invalid request body"
// @Failure      422     {object} ErrorResponse "Domain validation error"
// @Router       /api/v1/specialists [post]
func (h *SpecialistCreateHTTPHandler) CreateSpecialist(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req CreateSpecialistRequest
	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dto := ToCreateSpecialistDTO(req)
	specialist, err := h.UseCase.Execute(c.Request.Context(), dto)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"specialist": ToSpecialistResponse(specialist)})
}

func (h *SpecialistCreateHTTPHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/specialists", h.CreateSpecialist)
}
