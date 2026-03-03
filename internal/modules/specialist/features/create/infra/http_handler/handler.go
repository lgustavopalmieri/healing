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

type SpecialistCreateCommandInterface interface {
	Execute(ctx context.Context, input application.CreateSpecialistDTO) (*domain.Specialist, error)
}

type SpecialistCreateHTTPHandler struct {
	Command SpecialistCreateCommandInterface
}

func NewSpecialistCreateHTTPHandler(command SpecialistCreateCommandInterface) *SpecialistCreateHTTPHandler {
	return &SpecialistCreateHTTPHandler{
		Command: command,
	}
}

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
	specialist, err := h.Command.Execute(c.Request.Context(), dto)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"specialist": ToSpecialistResponse(specialist)})
}

func (h *SpecialistCreateHTTPHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/specialists", h.CreateSpecialist)
}
