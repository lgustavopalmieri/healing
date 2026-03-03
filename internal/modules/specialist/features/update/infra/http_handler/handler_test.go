package httphandler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/update/application"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/update/infra/http_handler/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }

func updateRequestFactory(overrides ...func(*UpdateSpecialistRequest)) UpdateSpecialistRequest {
	req := UpdateSpecialistRequest{
		Name:          strPtr("Dr. Maria Santos"),
		Email:         strPtr("maria@example.com"),
		Phone:         strPtr("+5511888888888"),
		Specialty:     strPtr("Neurology"),
		LicenseNumber: strPtr("CRM-654321"),
		Description:   strPtr("Experienced neurologist"),
		Keywords:      []string{"neurology", "brain"},
		AgreedToShare: boolPtr(true),
		Status:        strPtr("active"),
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
		Name:          "Dr. Maria Santos",
		Email:         "maria@example.com",
		Phone:         "+5511888888888",
		Specialty:     "Neurology",
		LicenseNumber: "CRM-654321",
		Description:   "Experienced neurologist",
		Keywords:      []string{"neurology", "brain"},
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

func setupRouter(handler *SpecialistUpdateHTTPHandler) *gin.Engine {
	router := gin.New()
	api := router.Group("/api/v1")
	handler.RegisterRoutes(api)
	return router
}

func TestSpecialistUpdateHTTPHandler_UpdateSpecialist(t *testing.T) {
	tests := []struct {
		name           string
		specialistID   string
		body           any
		setupMocks     func(*mocks.MockSpecialistUpdateCommandInterface)
		expectedStatus int
		validateBody   func(*testing.T, map[string]any)
	}{
		{
			name:         "success - updates specialist and returns 200 with updated data",
			specialistID: "550e8400-e29b-41d4-a716-446655440000",
			body:         updateRequestFactory(),
			setupMocks: func(mockCommand *mocks.MockSpecialistUpdateCommandInterface) {
				mockCommand.EXPECT().
					Execute(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ any, dto application.UpdateSpecialistDTO) (*domain.Specialist, error) {
						assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", dto.ID)
						assert.Equal(t, "Dr. Maria Santos", *dto.Name)
						assert.Equal(t, "maria@example.com", *dto.Email)
						assert.Equal(t, domain.StatusActive, *dto.Status)
						return specialistFactory(), nil
					}).
					Times(1)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body map[string]any) {
				specialist := body["specialist"].(map[string]any)
				assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", specialist["id"])
				assert.Equal(t, "Dr. Maria Santos", specialist["name"])
				assert.Equal(t, "maria@example.com", specialist["email"])
				assert.Equal(t, "Neurology", specialist["specialty"])
				assert.Equal(t, "active", specialist["status"])
				assert.Equal(t, 4.8, specialist["rating"])
				assert.NotEmpty(t, specialist["created_at"])
				assert.NotEmpty(t, specialist["updated_at"])
			},
		},
		{
			name:         "success - updates specialist with partial fields (only name)",
			specialistID: "550e8400-e29b-41d4-a716-446655440000",
			body: updateRequestFactory(func(r *UpdateSpecialistRequest) {
				r.Email = nil
				r.Phone = nil
				r.Specialty = nil
				r.LicenseNumber = nil
				r.Description = nil
				r.Keywords = nil
				r.AgreedToShare = nil
				r.Status = nil
			}),
			setupMocks: func(mockCommand *mocks.MockSpecialistUpdateCommandInterface) {
				mockCommand.EXPECT().
					Execute(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ any, dto application.UpdateSpecialistDTO) (*domain.Specialist, error) {
						assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", dto.ID)
						assert.NotNil(t, dto.Name)
						assert.Equal(t, "Dr. Maria Santos", *dto.Name)
						assert.Nil(t, dto.Email)
						assert.Nil(t, dto.Phone)
						assert.Nil(t, dto.Specialty)
						assert.Nil(t, dto.LicenseNumber)
						assert.Nil(t, dto.Description)
						assert.Nil(t, dto.Keywords)
						assert.Nil(t, dto.AgreedToShare)
						assert.Nil(t, dto.Status)
						return specialistFactory(), nil
					}).
					Times(1)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body map[string]any) {
				specialist := body["specialist"].(map[string]any)
				assert.Equal(t, "Dr. Maria Santos", specialist["name"])
			},
		},
		{
			name:         "failure - returns 400 when request body is invalid JSON",
			specialistID: "550e8400-e29b-41d4-a716-446655440000",
			body:         "invalid{{{json",
			setupMocks: func(mockCommand *mocks.MockSpecialistUpdateCommandInterface) {
				mockCommand.EXPECT().Execute(gomock.Any(), gomock.Any()).Times(0)
			},
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body map[string]any) {
				assert.NotEmpty(t, body["error"])
			},
		},
		{
			name:         "failure - returns 422 when command returns ErrSpecialistNotFound",
			specialistID: "nonexistent-id",
			body:         updateRequestFactory(),
			setupMocks: func(mockCommand *mocks.MockSpecialistUpdateCommandInterface) {
				mockCommand.EXPECT().
					Execute(gomock.Any(), gomock.Any()).
					Return(nil, application.ErrSpecialistNotFound).
					Times(1)
			},
			expectedStatus: http.StatusUnprocessableEntity,
			validateBody: func(t *testing.T, body map[string]any) {
				assert.Equal(t, application.ErrSpecialistNotFound.Error(), body["error"])
			},
		},
		{
			name:         "failure - returns 422 when command returns domain ErrInvalidName",
			specialistID: "550e8400-e29b-41d4-a716-446655440000",
			body: updateRequestFactory(func(r *UpdateSpecialistRequest) {
				r.Name = strPtr("")
			}),
			setupMocks: func(mockCommand *mocks.MockSpecialistUpdateCommandInterface) {
				mockCommand.EXPECT().
					Execute(gomock.Any(), gomock.Any()).
					Return(nil, domain.ErrInvalidName).
					Times(1)
			},
			expectedStatus: http.StatusUnprocessableEntity,
			validateBody: func(t *testing.T, body map[string]any) {
				assert.Equal(t, domain.ErrInvalidName.Error(), body["error"])
			},
		},
		{
			name:         "failure - returns 422 when command returns ErrUpdateSpecialist",
			specialistID: "550e8400-e29b-41d4-a716-446655440000",
			body:         updateRequestFactory(),
			setupMocks: func(mockCommand *mocks.MockSpecialistUpdateCommandInterface) {
				mockCommand.EXPECT().
					Execute(gomock.Any(), gomock.Any()).
					Return(nil, application.ErrUpdateSpecialist).
					Times(1)
			},
			expectedStatus: http.StatusUnprocessableEntity,
			validateBody: func(t *testing.T, body map[string]any) {
				assert.Equal(t, application.ErrUpdateSpecialist.Error(), body["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockCommand := mocks.NewMockSpecialistUpdateCommandInterface(ctrl)
			tt.setupMocks(mockCommand)

			handler := NewSpecialistUpdateHTTPHandler(mockCommand)
			router := setupRouter(handler)

			var bodyBytes []byte
			switch v := tt.body.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				bodyBytes, _ = json.Marshal(v)
			}

			url := "/api/v1/specialists/" + tt.specialistID
			req := httptest.NewRequest(http.MethodPatch, url, bytes.NewReader(bodyBytes))
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
