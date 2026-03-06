package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
	estest "github.com/lgustavopalmieri/healing-specialist/internal/commom/tests/elasticsearch"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/elasticsearch/indexes"
)

var testHelper = estest.NewTestHelper()

func TestMain(m *testing.M) {
	testHelper.RunTestMain(m)
}

type noopLogger struct{}

func (l *noopLogger) Debug(ctx context.Context, msg string, fields ...observability.Field) {}
func (l *noopLogger) Info(ctx context.Context, msg string, fields ...observability.Field)  {}
func (l *noopLogger) Warn(ctx context.Context, msg string, fields ...observability.Field)  {}
func (l *noopLogger) Error(ctx context.Context, msg string, fields ...observability.Field) {}

type captureDispatcher struct {
	Events []event.Event
}

func (d *captureDispatcher) Dispatch(ctx context.Context, evt event.Event) error {
	d.Events = append(d.Events, evt)
	return nil
}

type failingDispatcher struct{}

func (d *failingDispatcher) Dispatch(ctx context.Context, evt event.Event) error {
	return errors.New("kafka unavailable")
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

func getDocument(t *testing.T, client *elasticsearch.Client, indexName string, id string) map[string]any {
	res, err := client.Get(indexName, id)
	require.NoError(t, err)
	defer res.Body.Close()

	if res.IsError() {
		return nil
	}

	var result map[string]any
	err = json.NewDecoder(res.Body).Decode(&result)
	require.NoError(t, err)

	source, ok := result["_source"].(map[string]any)
	if !ok {
		return nil
	}
	return source
}

func TestElasticsearchRepository_Update(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(*testing.T, *elasticsearch.Client, string)
		specialist     func() *domain.Specialist
		expectError    bool
		validateResult func(*testing.T, *elasticsearch.Client, string, *domain.Specialist)
	}{
		{
			name:  "success - indexes specialist document and can be retrieved",
			setup: func(t *testing.T, client *elasticsearch.Client, indexName string) {},
			specialist: func() *domain.Specialist {
				return specialistFactory()
			},
			expectError: false,
			validateResult: func(t *testing.T, client *elasticsearch.Client, indexName string, specialist *domain.Specialist) {
				res, err := client.Indices.Refresh(client.Indices.Refresh.WithIndex(indexName))
				require.NoError(t, err)
				res.Body.Close()

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
			setup: func(t *testing.T, client *elasticsearch.Client, indexName string) {},
			specialist: func() *domain.Specialist {
				return specialistFactory()
			},
			expectError: false,
			validateResult: func(t *testing.T, client *elasticsearch.Client, indexName string, specialist *domain.Specialist) {
				specialist.Name = "Dr. Updated Name"
				specialist.Specialty = "Neurologia"

				logger := &noopLogger{}
				dispatcher := &captureDispatcher{}
				repo := NewRepository(client, indexName, logger, dispatcher)
				err := repo.Update(context.Background(), specialist)
				require.NoError(t, err)

				res, err := client.Indices.Refresh(client.Indices.Refresh.WithIndex(indexName))
				require.NoError(t, err)
				res.Body.Close()

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
			logger := &noopLogger{}
			dispatcher := &captureDispatcher{}
			repo := NewRepository(client, indexName, logger, dispatcher)

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

func TestElasticsearchRepository_PublishDLQ(t *testing.T) {
	tests := []struct {
		name           string
		dispatcher     event.EventDispatcher
		expectError    bool
		validateResult func(*testing.T, *captureDispatcher)
	}{
		{
			name:        "success - publishes DLQ event with specialist ID and reason",
			dispatcher:  &captureDispatcher{},
			expectError: false,
			validateResult: func(t *testing.T, dispatcher *captureDispatcher) {
				require.Len(t, dispatcher.Events, 1)
				evt := dispatcher.Events[0]
				assert.Equal(t, ElasticsearchUpdateDLQEventName, evt.Name)

				payload, ok := evt.Payload.(map[string]any)
				require.True(t, ok)
				assert.Equal(t, "specialist-dlq-123", payload["id"])
				assert.Equal(t, "es unavailable", payload["reason"])
				assert.Equal(t, "elasticsearch", payload["source"])
			},
		},
		{
			name:           "failure - returns error when event dispatcher fails",
			dispatcher:     &failingDispatcher{},
			expectError:    true,
			validateResult: func(t *testing.T, dispatcher *captureDispatcher) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := &noopLogger{}
			repo := &Repository{
				Logger:          logger,
				EventDispatcher: tt.dispatcher,
			}

			specialist := &domain.Specialist{ID: "specialist-dlq-123"}
			reason := errors.New("es unavailable")

			err := repo.PublishDLQ(context.Background(), specialist, reason)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if capture, ok := tt.dispatcher.(*captureDispatcher); ok {
				tt.validateResult(t, capture)
			}
		})
	}
}

func seedESDocument(t *testing.T, client *elasticsearch.Client, indexName string, specialist *domain.Specialist) {
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

	res, err := client.Index(
		indexName,
		bytes.NewReader(body),
		client.Index.WithContext(context.Background()),
		client.Index.WithDocumentID(specialist.ID),
		client.Index.WithRefresh("true"),
	)
	require.NoError(t, err)
	defer res.Body.Close()
	require.False(t, res.IsError())
}
