package httphandler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/update/application"
)

type SpecialistUpdateUseCaseInterface interface {
	Execute(ctx context.Context, input application.UpdateSpecialistDTO) (*domain.Specialist, error)
}

type SpecialistUpdateHTTPHandler struct {
	UseCase SpecialistUpdateUseCaseInterface
}

func NewSpecialistUpdateHTTPHandler(useCase SpecialistUpdateUseCaseInterface) *SpecialistUpdateHTTPHandler {
	return &SpecialistUpdateHTTPHandler{
		UseCase: useCase,
	}
}

// UpdateSpecialist godoc
// @Summary      Update a specialist
// @Description  Partially updates a specialist's profile data. Only provided fields are updated (PATCH semantics).
// @Tags         specialists
// @Accept       json
// @Produce      json
// @Param        id      path     string                  true "Specialist UUID"
// @Param        request body     UpdateSpecialistRequest  true "Fields to update"
// @Success      200     {object} UpdateSpecialistSuccessResponse
// @Failure      400     {object} ErrorResponse "Missing ID or invalid request body"
// @Failure      422     {object} ErrorResponse "Domain validation error"
// @Router       /api/v1/specialists/{id} [patch]
func (h *SpecialistUpdateHTTPHandler) UpdateSpecialist(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "specialist id is required"})
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req UpdateSpecialistRequest
	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dto := ToUpdateSpecialistDTO(id, req)
	specialist, err := h.UseCase.Execute(c.Request.Context(), dto)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"specialist": ToSpecialistResponse(specialist)})
}

func (h *SpecialistUpdateHTTPHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.PATCH("/specialists/:id", h.UpdateSpecialist)
}
