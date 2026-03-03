package httphandler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	searchoutput "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search/search_output"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/search/application"
)

type SpecialistSearchCommandInterface interface {
	Execute(ctx context.Context, dto *application.SearchSpecialistsDTO) (*searchoutput.ListSearchOutput, error)
}

type SpecialistSearchHTTPHandler struct {
	Command SpecialistSearchCommandInterface
}

func NewSpecialistSearchHTTPHandler(command SpecialistSearchCommandInterface) *SpecialistSearchHTTPHandler {
	return &SpecialistSearchHTTPHandler{
		Command: command,
	}
}

func (h *SpecialistSearchHTTPHandler) SearchSpecialists(c *gin.Context) {
	var req SearchSpecialistsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dto, err := ToSearchDTO(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	output, err := h.Command.Execute(c.Request.Context(), dto)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ToSearchResponse(output))
}

func (h *SpecialistSearchHTTPHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/specialists/search", h.SearchSpecialists)
}
