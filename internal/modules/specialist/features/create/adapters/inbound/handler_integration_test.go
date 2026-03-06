package integration

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
	postgrestest "github.com/lgustavopalmieri/healing-specialist/internal/commom/tests/database/postgresql"
	kafkatest "github.com/lgustavopalmieri/healing-specialist/internal/commom/tests/event/kafka"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	createdb "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/adapters/outbound/database"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/application"
	vldb "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/event_listeners/validate_license/adapters/outbound/database"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/event_listeners/validate_license/adapters/outbound/external"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/event_listeners/validate_license/listener"
	platformkafka "github.com/lgustavopalmieri/healing-specialist/internal/platform/kafka"
)

var (
	pgHelper    = postgrestest.NewTestHelper()
	kafkaHelper = kafkatest.NewTestHelper()
)

func TestMain(m *testing.M) {
	pgHelper.RunTestMainWithoutExit(m)
	kafkaHelper.SharedContainer = kafkatest.SetupKafkaContainer(&testing.T{})
	err := kafkaHelper.SharedContainer.CreateTopics(
		application.SpecialistCreatedEventName,
		"specialist.updated",
	)
	if err != nil {
		panic("failed to create kafka topics: " + err.Error())
	}
	code := m.Run()
	pgHelper.TerminateContainer()
	if kafkaHelper.SharedContainer != nil {
		kafkaHelper.SharedContainer.Terminate(&testing.T{})
	}
	os.Exit(code)
}

type noopSpan struct{}

func (s *noopSpan) End()                                                     {}
func (s *noopSpan) RecordError(err error)                                    {}
func (s *noopSpan) SetAttribute(key string, attr ...observability.Attribute) {}

type noopTracer struct{}

func (t *noopTracer) Start(ctx context.Context, name string) (context.Context, observability.Span) {
	return ctx, &noopSpan{}
}

type noopLogger struct{}

func (l *noopLogger) Debug(ctx context.Context, msg string, fields ...observability.Field) {}
func (l *noopLogger) Info(ctx context.Context, msg string, fields ...observability.Field)  {}
func (l *noopLogger) Warn(ctx context.Context, msg string, fields ...observability.Field)  {}
func (l *noopLogger) Error(ctx context.Context, msg string, fields ...observability.Field) {}

func specialistIntegrationFactory(overrides ...func(*domain.Specialist)) *domain.Specialist {
	now := time.Now().UTC()
	uniqueID := uuid.New().String()
	s := &domain.Specialist{
		ID:            uniqueID,
		Name:          "Dr. Integration Test",
		Email:         "integration+" + uniqueID[:8] + "@example.com",
		Phone:         "+5511999999999",
		Specialty:     "Cardiologia",
		LicenseNumber: "CRM" + uniqueID[:6],
		Description:   "Integration test specialist",
		Keywords:      []string{"test", "integration"},
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

func seedSpecialist(t *testing.T, db *sql.DB, s *domain.Specialist) {
	createRepo := createdb.NewSpecialistCreateRepository(db)
	_, err := createRepo.Save(context.Background(), s)
	require.NoError(t, err)
}

func setupMockLicenseServer(valid bool) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"valid": valid})
	}))
}

func setupHandler(db *sql.DB, licenseServerURL string, eventPublisher event.EventDispatcher) *listener.ValidateLicenseHandler {
	repository := vldb.NewValidateLicenseRepository(db)
	gateway := external.NewLicenseGateway(licenseServerURL, &http.Client{})
	tracer := &noopTracer{}
	logger := &noopLogger{}
	return listener.NewValidateLicenseHandler(repository, gateway, eventPublisher, tracer, logger)
}

func TestValidateLicenseHandler_Integration(t *testing.T) {
	tests := []struct {
		name            string
		licenseValid    bool
		specialistSetup func() *domain.Specialist
		expectStatus    domain.SpecialistStatus
		expectEvent     bool
	}{
		{
			name:         "success - consumes specialist.created, validates license, updates status to authorized_license and publishes specialist.updated",
			licenseValid: true,
			specialistSetup: func() *domain.Specialist {
				return specialistIntegrationFactory()
			},
			expectStatus: domain.StatusAuthorizedLicense,
			expectEvent:  true,
		},
		// TODO: re-enable when external license validation API is integrated
		// Currently the gateway is hardcoded to return true (see external/gateway.go)
		// {
		// 	name:         "failure - does not update status when license is invalid",
		// 	licenseValid: false,
		// 	specialistSetup: func() *domain.Specialist {
		// 		return specialistIntegrationFactory()
		// 	},
		// 	expectStatus: domain.StatusPending,
		// 	expectEvent:  false,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, dbCleanup := pgHelper.SetupTestDB(t)
			defer dbCleanup()

			specialist := tt.specialistSetup()
			seedSpecialist(t, db, specialist)

			licenseServer := setupMockLicenseServer(tt.licenseValid)
			defer licenseServer.Close()

			brokers := kafkaHelper.SharedContainer.Brokers

			producer, err := platformkafka.NewKafkaProducer(brokers)
			require.NoError(t, err)

			handler := setupHandler(db, licenseServer.URL, producer)

			manager := event.NewListenerManager()
			manager.Register(application.SpecialistCreatedEventName, handler)

			consumerGroupID := "test-validate-license-" + uuid.New().String()[:8]

			consumer, err := platformkafka.NewKafkaConsumer(brokers, consumerGroupID, manager)
			require.NoError(t, err)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			go consumer.Start(ctx)

			time.Sleep(3 * time.Second)

			payload := listener.ValidateLicenseEventPayload{
				ID:            specialist.ID,
				Email:         specialist.Email,
				LicenseNumber: specialist.LicenseNumber,
				Specialty:     specialist.Specialty,
			}
			kafkaHelper.SharedContainer.ProduceEvent(t, application.SpecialistCreatedEventName, payload)

			time.Sleep(5 * time.Second)

			repo := vldb.NewValidateLicenseRepository(db)
			updated, err := repo.FindByID(context.Background(), specialist.ID)
			require.NoError(t, err)
			assert.Equal(t, tt.expectStatus, updated.Status)

			if tt.expectEvent {
				eventData := kafkaHelper.SharedContainer.ConsumeEvent(
					t,
					"specialist.updated",
					"test-verify-event-"+uuid.New().String()[:8],
					10*time.Second,
				)
				assert.NotNil(t, eventData)

				var eventPayload map[string]any
				err = json.Unmarshal(eventData, &eventPayload)
				require.NoError(t, err)
				assert.Equal(t, specialist.ID, eventPayload["id"])
			}

			cancel()
		})
	}
}
