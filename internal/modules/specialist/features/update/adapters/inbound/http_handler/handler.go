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

type SpecialistUpdateCommandInterface interface {
	Execute(ctx context.Context, input application.UpdateSpecialistDTO) (*domain.Specialist, error)
}

type SpecialistUpdateHTTPHandler struct {
	Command SpecialistUpdateCommandInterface
}

func NewSpecialistUpdateHTTPHandler(command SpecialistUpdateCommandInterface) *SpecialistUpdateHTTPHandler {
	return &SpecialistUpdateHTTPHandler{
		Command: command,
	}
}

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
	specialist, err := h.Command.Execute(c.Request.Context(), dto)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"specialist": ToSpecialistResponse(specialist)})
}

func (h *SpecialistUpdateHTTPHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.PATCH("/specialists/:id", h.UpdateSpecialist)
}
