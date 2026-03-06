package httphandler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/create"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/adapters/inbound/http_handler/mocks"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/application"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func createRequestFactory(overrides ...func(*CreateSpecialistRequest)) CreateSpecialistRequest {
	req := CreateSpecialistRequest{
		Name:          "Dr. João Silva",
		Email:         "joao@exemplo.com",
		Phone:         "+5511999999999",
		Specialty:     "Cardiology",
		LicenseNumber: "CRM-123456",
		Description:   "Experienced cardiologist",
		Keywords:      []string{"heart", "cardiology"},
		AgreedToShare: true,
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
		Rating:        0.0,
		Status:        domain.StatusPending,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	for _, o := range overrides {
		o(s)
	}
	return s
}

func setupRouter(handler *SpecialistCreateHTTPHandler) *gin.Engine {
	router := gin.New()
	api := router.Group("/api/v1")
	handler.RegisterRoutes(api)
	return router
}

func TestSpecialistCreateHTTPHandler_CreateSpecialist(t *testing.T) {
	tests := []struct {
		name           string
		body           any
		setupMocks     func(*mocks.MockSpecialistCreateUseCaseInterface)
		expectedStatus int
		validateBody   func(*testing.T, map[string]any)
	}{
		{
			name: "success - creates specialist and returns 201 with specialist data",
			body: createRequestFactory(),
			setupMocks: func(mockUseCase *mocks.MockSpecialistCreateUseCaseInterface) {
				mockUseCase.EXPECT().
					Execute(gomock.Any(), application.CreateSpecialistDTO{
						Name:          "Dr. João Silva",
						Email:         "joao@exemplo.com",
						Phone:         "+5511999999999",
						Specialty:     "Cardiology",
						LicenseNumber: "CRM-123456",
						Description:   "Experienced cardiologist",
						Keywords:      []string{"heart", "cardiology"},
						AgreedToShare: true,
					}).
					Return(specialistFactory(), nil).
					Times(1)
			},
			expectedStatus: http.StatusCreated,
			validateBody: func(t *testing.T, body map[string]any) {
				assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", body["id"])
				assert.Equal(t, "Dr. João Silva", body["name"])
				assert.Equal(t, "joao@exemplo.com", body["email"])
				assert.Equal(t, "+5511999999999", body["phone"])
				assert.Equal(t, "Cardiology", body["specialty"])
				assert.Equal(t, "CRM-123456", body["license_number"])
				assert.Equal(t, "Experienced cardiologist", body["description"])
				assert.Equal(t, true, body["agreed_to_share"])
				assert.Equal(t, "pending", body["status"])
				assert.NotEmpty(t, body["created_at"])
				assert.NotEmpty(t, body["updated_at"])
			},
		},
		{
			name: "failure - returns 400 when request body is invalid JSON",
			body: "invalid-json{{{",
			setupMocks: func(mockUseCase *mocks.MockSpecialistCreateUseCaseInterface) {
				mockUseCase.EXPECT().Execute(gomock.Any(), gomock.Any()).Times(0)
			},
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body map[string]any) {
				assert.NotEmpty(t, body["error"])
			},
		},
		{
			name: "failure - returns 422 when command returns domain validation error",
			body: createRequestFactory(func(r *CreateSpecialistRequest) {
				r.Name = ""
			}),
			setupMocks: func(mockUseCase *mocks.MockSpecialistCreateUseCaseInterface) {
				mockUseCase.EXPECT().
					Execute(gomock.Any(), gomock.Any()).
					Return(nil, create.ErrInvalidName).
					Times(1)
			},
			expectedStatus: http.StatusUnprocessableEntity,
			validateBody: func(t *testing.T, body map[string]any) {
				assert.Equal(t, create.ErrInvalidName.Error(), body["error"])
			},
		},
		{
			name: "failure - returns 422 when command returns ErrSaveSpecialist",
			body: createRequestFactory(),
			setupMocks: func(mockUseCase *mocks.MockSpecialistCreateUseCaseInterface) {
				mockUseCase.EXPECT().
					Execute(gomock.Any(), gomock.Any()).
					Return(nil, application.ErrSaveSpecialist).
					Times(1)
			},
			expectedStatus: http.StatusUnprocessableEntity,
			validateBody: func(t *testing.T, body map[string]any) {
				assert.Equal(t, application.ErrSaveSpecialist.Error(), body["error"])
			},
		},
		{
			name: "failure - returns 422 when command returns context.Canceled",
			body: createRequestFactory(),
			setupMocks: func(mockUseCase *mocks.MockSpecialistCreateUseCaseInterface) {
				mockUseCase.EXPECT().
					Execute(gomock.Any(), gomock.Any()).
					Return(nil, context.Canceled).
					Times(1)
			},
			expectedStatus: http.StatusUnprocessableEntity,
			validateBody: func(t *testing.T, body map[string]any) {
				assert.Equal(t, context.Canceled.Error(), body["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUseCase := mocks.NewMockSpecialistCreateUseCaseInterface(ctrl)
			tt.setupMocks(mockUseCase)

			handler := NewSpecialistCreateHTTPHandler(mockUseCase)
			router := setupRouter(handler)

			var bodyBytes []byte
			switch v := tt.body.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				bodyBytes, _ = json.Marshal(v)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/specialists", bytes.NewReader(bodyBytes))
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
