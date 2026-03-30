package httphandler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	searchoutput "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search/search_output"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/search/application"
)

type SpecialistSearchUseCaseInterface interface {
	Execute(ctx context.Context, dto *application.SearchSpecialistsDTO) (*searchoutput.ListSearchOutput, error)
}

type SpecialistSearchHTTPHandler struct {
	UseCase SpecialistSearchUseCaseInterface
}

func NewSpecialistSearchHTTPHandler(useCase SpecialistSearchUseCaseInterface) *SpecialistSearchHTTPHandler {
	return &SpecialistSearchHTTPHandler{
		UseCase: useCase,
	}
}

// SearchSpecialists godoc
// @Summary      Search specialists
// @Description  Searches specialists using full-text search, filters, sorting and cursor-based pagination. Powered by OpenSearch.
// @Tags         specialists
// @Accept       json
// @Produce      json
// @Param        request body     SearchSpecialistsRequest true "Search criteria"
// @Success      200     {object} SearchSpecialistsResponse
// @Failure      400     {object} ErrorResponse "Invalid request body or search parameters"
// @Failure      422     {object} ErrorResponse "Search execution error"
// @Router       /api/v1/specialists/search [post]
func (h *SpecialistSearchHTTPHandler) SearchSpecialists(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req SearchSpecialistsRequest
	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dto, err := ToSearchDTO(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	output, err := h.UseCase.Execute(c.Request.Context(), dto)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ToSearchResponse(output))
}

func (h *SpecialistSearchHTTPHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/specialists/search", h.SearchSpecialists)
}
