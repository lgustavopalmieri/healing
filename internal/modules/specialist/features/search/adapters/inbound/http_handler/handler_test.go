package httphandler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	cursor "github.com/lgustavopalmieri/healing-specialist/internal/commom/value-objects/pagination/cursor"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	searchoutput "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search/search_output"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/search/application"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/search/adapters/inbound/http_handler/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func searchRequestFactory(overrides ...func(*SearchSpecialistsRequest)) SearchSpecialistsRequest {
	req := SearchSpecialistsRequest{
		SearchTerm: "cardiology",
		Filters:    []FilterRequest{},
		Sort: []SortRequest{
			{Field: "rating", Order: "desc"},
		},
		PageSize:  10,
		Cursor:    "",
		Direction: "next",
	}
	for _, o := range overrides {
		o(&req)
	}
	return req
}

func specialistFactory(overrides ...func(*domain.Specialist)) *domain.Specialist {
	now := time.Now().UTC()
	s := &domain.Specialist{
		ID:            "550e8400-e29b-41d4-a716-446655440000",
		Name:          "Dr. João Silva",
		Email:         "joao@exemplo.com",
		Phone:         "+5511999999999",
		Specialty:     "Cardiology",
		LicenseNumber: "CRM-123456",
		Description:   "Experienced cardiologist",
		Keywords:      []string{"heart", "cardiology"},
		AgreedToShare: true,
		Rating:        4.8,
		Status:        domain.StatusActive,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	for _, o := range overrides {
		o(s)
	}
	return s
}

func searchOutputFactory(specialists []*domain.Specialist, nextCursor *string, prevCursor *string, hasNext bool, hasPrev bool) *searchoutput.ListSearchOutput {
	return searchoutput.NewListSearchOutput(
		specialists,
		cursor.NewCursorPaginationOutput(nextCursor, prevCursor, hasNext, hasPrev, len(specialists)),
	)
}

func setupRouter(handler *SpecialistSearchHTTPHandler) *gin.Engine {
	router := gin.New()
	api := router.Group("/api/v1")
	handler.RegisterRoutes(api)
	return router
}

func TestSpecialistSearchHTTPHandler_SearchSpecialists(t *testing.T) {
	tests := []struct {
		name           string
		body           any
		setupMocks     func(*mocks.MockSpecialistSearchCommandInterface)
		expectedStatus int
		validateBody   func(*testing.T, map[string]any)
	}{
		{
			name: "success - searches specialists and returns 200 with paginated results",
			body: searchRequestFactory(),
			setupMocks: func(mockCommand *mocks.MockSpecialistSearchCommandInterface) {
				specialists := []*domain.Specialist{specialistFactory()}
				output := searchOutputFactory(specialists, nil, nil, false, false)
				mockCommand.EXPECT().
					Execute(gomock.Any(), gomock.Any()).
					Return(output, nil).
					Times(1)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body map[string]any) {
				specialists := body["specialists"].([]any)
				assert.Len(t, specialists, 1)
				first := specialists[0].(map[string]any)
				assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", first["id"])
				assert.Equal(t, "Dr. João Silva", first["name"])
				assert.Equal(t, "Cardiology", first["specialty"])
				assert.Equal(t, 4.8, first["rating"])
				assert.Equal(t, "active", first["status"])

				pagination := body["pagination"].(map[string]any)
				assert.Equal(t, false, pagination["has_next_page"])
				assert.Equal(t, false, pagination["has_previous_page"])
				assert.Equal(t, float64(1), pagination["total_items_in_page"])
			},
		},
		{
			name: "success - returns 200 with empty results when no specialists match",
			body: searchRequestFactory(),
			setupMocks: func(mockCommand *mocks.MockSpecialistSearchCommandInterface) {
				output := searchOutputFactory([]*domain.Specialist{}, nil, nil, false, false)
				mockCommand.EXPECT().
					Execute(gomock.Any(), gomock.Any()).
					Return(output, nil).
					Times(1)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body map[string]any) {
				specialists := body["specialists"].([]any)
				assert.Empty(t, specialists)
				pagination := body["pagination"].(map[string]any)
				assert.Equal(t, float64(0), pagination["total_items_in_page"])
				assert.Equal(t, false, pagination["has_next_page"])
			},
		},
		{
			name: "success - returns 200 with pagination cursors when has more pages",
			body: searchRequestFactory(),
			setupMocks: func(mockCommand *mocks.MockSpecialistSearchCommandInterface) {
				specialists := []*domain.Specialist{specialistFactory()}
				nextCur := "eyJzb3J0IjpbNC44XX0="
				output := searchOutputFactory(specialists, &nextCur, nil, true, false)
				mockCommand.EXPECT().
					Execute(gomock.Any(), gomock.Any()).
					Return(output, nil).
					Times(1)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body map[string]any) {
				pagination := body["pagination"].(map[string]any)
				assert.Equal(t, true, pagination["has_next_page"])
				assert.Equal(t, false, pagination["has_previous_page"])
				assert.NotEmpty(t, pagination["next_cursor"])
			},
		},
		{
			name: "failure - returns 400 when request body is invalid JSON",
			body: "not-valid-json{{{",
			setupMocks: func(mockCommand *mocks.MockSpecialistSearchCommandInterface) {
				mockCommand.EXPECT().Execute(gomock.Any(), gomock.Any()).Times(0)
			},
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body map[string]any) {
				assert.NotEmpty(t, body["error"])
			},
		},
		{
			name: "failure - returns 422 when command returns ErrSearchExecution",
			body: searchRequestFactory(),
			setupMocks: func(mockCommand *mocks.MockSpecialistSearchCommandInterface) {
				mockCommand.EXPECT().
					Execute(gomock.Any(), gomock.Any()).
					Return(nil, application.ErrSearchExecution).
					Times(1)
			},
			expectedStatus: http.StatusUnprocessableEntity,
			validateBody: func(t *testing.T, body map[string]any) {
				assert.Equal(t, application.ErrSearchExecution.Error(), body["error"])
			},
		},
		{
			name: "failure - returns 422 when command returns ErrInvalidSearchInput",
			body: searchRequestFactory(),
			setupMocks: func(mockCommand *mocks.MockSpecialistSearchCommandInterface) {
				mockCommand.EXPECT().
					Execute(gomock.Any(), gomock.Any()).
					Return(nil, application.ErrInvalidSearchInput).
					Times(1)
			},
			expectedStatus: http.StatusUnprocessableEntity,
			validateBody: func(t *testing.T, body map[string]any) {
				assert.Equal(t, application.ErrInvalidSearchInput.Error(), body["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockCommand := mocks.NewMockSpecialistSearchCommandInterface(ctrl)
			tt.setupMocks(mockCommand)

			handler := NewSpecialistSearchHTTPHandler(mockCommand)
			router := setupRouter(handler)

			var bodyBytes []byte
			switch v := tt.body.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				bodyBytes, _ = json.Marshal(v)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/specialists/search", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			var responseBody map[string]any
			err := json.Unmarshal(rec.Body.Bytes(), &responseBody)
			assert.NoError(t, err)

			tt.validateBody(t, responseBody)
		})
	}
}
