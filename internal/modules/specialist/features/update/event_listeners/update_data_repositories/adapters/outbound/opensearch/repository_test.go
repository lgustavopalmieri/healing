package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ostest "github.com/lgustavopalmieri/healing-specialist/internal/commom/tests/opensearch"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/opensearch/indexes"
)

var testHelper = ostest.NewTestHelper()

func TestMain(m *testing.M) {
	testHelper.RunTestMain(m)
}

func specialistFactory(overrides ...func(*domain.Specialist)) *domain.Specialist {
	now := time.Now().UTC()
	uniqueID := uuid.New().String()
	s := &domain.Specialist{
		ID:            uniqueID,
		Name:          "Dr. João Silva",
		Email:         "joao.silva+" + uniqueID[:8] + "@example.com",
		Phone:         "+5511999999999",
		Specialty:     "Cardiologia",
		LicenseNumber: "CRM" + uniqueID[:6],
		Description:   "Cardiologista especializado em arritmias",
		Keywords:      []string{"cardiologia", "arritmia"},
		AgreedToShare: true,
		Rating:        4.5,
		Status:        domain.StatusActive,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	for _, o := range overrides {
		o(s)
	}
	return s
}

func getDocument(t *testing.T, client *opensearchapi.Client, indexName string, id string) map[string]any {
	resp, err := client.Document.Get(context.Background(), opensearchapi.DocumentGetReq{
		Index:      indexName,
		DocumentID: id,
	})
	if err != nil {
		return nil
	}

	if !resp.Found {
		return nil
	}

	var source map[string]any
	if err := json.Unmarshal(resp.Source, &source); err != nil {
		require.NoError(t, err)
	}
	return source
}

func TestOpenSearchRepository_Update(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(*testing.T, *opensearchapi.Client, string)
		specialist     func() *domain.Specialist
		expectError    bool
		validateResult func(*testing.T, *opensearchapi.Client, string, *domain.Specialist)
	}{
		{
			name:  "success - indexes specialist document and can be retrieved",
			setup: func(t *testing.T, client *opensearchapi.Client, indexName string) {},
			specialist: func() *domain.Specialist {
				return specialistFactory()
			},
			expectError: false,
			validateResult: func(t *testing.T, client *opensearchapi.Client, indexName string, specialist *domain.Specialist) {
				_, err := client.Indices.Refresh(context.Background(), &opensearchapi.IndicesRefreshReq{
					Indices: []string{indexName},
				})
				require.NoError(t, err)

				doc := getDocument(t, client, indexName, specialist.ID)
				require.NotNil(t, doc)
				assert.Equal(t, specialist.ID, doc["id"])
				assert.Equal(t, specialist.Name, doc["name"])
				assert.Equal(t, specialist.Email, doc["email"])
				assert.Equal(t, specialist.Specialty, doc["specialty"])
				assert.Equal(t, string(specialist.Status), doc["status"])
			},
		},
		{
			name:  "success - updates existing specialist document with new data",
			setup: func(t *testing.T, client *opensearchapi.Client, indexName string) {},
			specialist: func() *domain.Specialist {
				return specialistFactory()
			},
			expectError: false,
			validateResult: func(t *testing.T, client *opensearchapi.Client, indexName string, specialist *domain.Specialist) {
				specialist.Name = "Dr. Updated Name"
				specialist.Specialty = "Neurologia"

				repo := NewRepository(client, indexName)
				err := repo.Update(context.Background(), specialist)
				require.NoError(t, err)

				_, err = client.Indices.Refresh(context.Background(), &opensearchapi.IndicesRefreshReq{
					Indices: []string{indexName},
				})
				require.NoError(t, err)

				doc := getDocument(t, client, indexName, specialist.ID)
				require.NotNil(t, doc)
				assert.Equal(t, "Dr. Updated Name", doc["name"])
				assert.Equal(t, "Neurologia", doc["specialty"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, indexName, cleanup := testHelper.SetupTestIndex(t, indexes.CreateSpecialistsIndex)
			defer cleanup()

			tt.setup(t, client, indexName)

			specialist := tt.specialist()
			repo := NewRepository(client, indexName)

			err := repo.Update(context.Background(), specialist)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			tt.validateResult(t, client, indexName, specialist)
		})
	}
}

func seedOSDocument(t *testing.T, client *opensearchapi.Client, indexName string, specialist *domain.Specialist) {
	doc := map[string]any{
		"id":              specialist.ID,
		"name":            specialist.Name,
		"email":           specialist.Email,
		"phone":           specialist.Phone,
		"specialty":       specialist.Specialty,
		"license_number":  specialist.LicenseNumber,
		"description":     specialist.Description,
		"keywords":        specialist.Keywords,
		"agreed_to_share": specialist.AgreedToShare,
		"rating":          specialist.Rating,
		"status":          string(specialist.Status),
		"created_at":      specialist.CreatedAt,
		"updated_at":      specialist.UpdatedAt,
	}

	body, err := json.Marshal(doc)
	require.NoError(t, err)

	_, err = client.Index(
		context.Background(),
		opensearchapi.IndexReq{
			Index:      indexName,
			DocumentID: specialist.ID,
			Body:       bytes.NewReader(body),
			Params:     opensearchapi.IndexParams{Refresh: "true"},
		},
	)
	require.NoError(t, err)
}
