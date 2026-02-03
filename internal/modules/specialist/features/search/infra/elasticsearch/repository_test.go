package elasticsearch

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
	elasticsearchtest "github.com/lgustavopalmieri/healing-specialist/internal/commom/tests/elasticsearch"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/value-objects/pagination/cursor"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	searchinput "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search/search_input"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/elasticsearch/indexes"
)

var testHelper *elasticsearchtest.TestHelper

func TestMain(m *testing.M) {
	testHelper = elasticsearchtest.NewTestHelper()
	testHelper.RunTestMain(m)
}

func TestRepository_Search(t *testing.T) {
	tests := []struct {
		name           string
		setupData      func() []*elasticsearchtest.SpecialistDocument
		input          func() *searchinput.ListSearchInput
		expectError    bool
		validateResult func(*testing.T, []*domain.Specialist, []*elasticsearchtest.SpecialistDocument)
	}{
		{
			name: "happy path - returns only specialists with dr/dra prefix",
			setupData: func() []*elasticsearchtest.SpecialistDocument {
				return []*elasticsearchtest.SpecialistDocument{
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Dr. João Silva"
						d.Email = "joao.silva@example.com"
						d.Specialty = "Cardiologia"
						d.LicenseNumber = "CRM-SP-123456"
					}),
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Dra. Maria Santos"
						d.Email = "maria.santos@example.com"
						d.Specialty = "Neurologia"
						d.LicenseNumber = "CRM-RJ-654321"
					}),
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Dr. Pedro Oliveira"
						d.Email = "pedro.oliveira@example.com"
						d.Specialty = "Ortopedia"
						d.LicenseNumber = "CRM-MG-789012"
					}),
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Dra. Ana Costa"
						d.Email = "ana.costa@example.com"
						d.Specialty = "Pediatria"
						d.LicenseNumber = "CRM-SP-345678"
					}),
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Carlos Ferreira"
						d.Email = "carlos.ferreira@example.com"
						d.Specialty = "Dermatologia"
						d.LicenseNumber = "CRM-RS-901234"
					}),
				}
			},
			input: func() *searchinput.ListSearchInput {
				searchTerm := "dr"
				pagination, _ := cursor.NewCursorPaginationInput(nil, 10, cursor.DirectionNext)
				input, _ := searchinput.NewListSearchInput(&searchTerm, nil, nil, pagination)
				return input
			},
			expectError: false,
			validateResult: func(t *testing.T, result []*domain.Specialist, indexed []*elasticsearchtest.SpecialistDocument) {
				require.Len(t, result, 4)

				expectedNames := map[string]bool{
					"Dr. João Silva":     false,
					"Dra. Maria Santos":  false,
					"Dr. Pedro Oliveira": false,
					"Dra. Ana Costa":     false,
				}

				for _, specialist := range result {
					assert.NotEmpty(t, specialist.ID)
					assert.NotEmpty(t, specialist.Name)
					assert.NotEmpty(t, specialist.Email)
					assert.NotEmpty(t, specialist.Specialty)
					assert.NotEmpty(t, specialist.LicenseNumber)
					assert.True(t, specialist.AgreedToShare)

					if _, exists := expectedNames[specialist.Name]; exists {
						expectedNames[specialist.Name] = true
					}
				}

				for name, found := range expectedNames {
					assert.True(t, found, "Expected specialist %s not found in results", name)
				}

				for _, specialist := range result {
					assert.NotEqual(t, "Carlos Ferreira", specialist.Name, "Carlos Ferreira should not be in results")
				}
			},
		},
		{
			name: "happy path - returns only specialists with specific specialty",
			setupData: func() []*elasticsearchtest.SpecialistDocument {
				return []*elasticsearchtest.SpecialistDocument{
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Dr. João Silva"
						d.Email = "joao.silva@example.com"
						d.Specialty = "Cardiologia"
						d.LicenseNumber = "CRM-SP-123456"
						d.Description = "Especialista em doenças cardiovasculares"
						d.Keywords = []string{"coração", "hipertensão"}
					}),
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Dra. Maria Santos"
						d.Email = "maria.santos@example.com"
						d.Specialty = "Neurologia"
						d.LicenseNumber = "CRM-RJ-654321"
						d.Description = "Especialista em doenças neurológicas"
						d.Keywords = []string{"cérebro", "alzheimer"}
					}),
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Dr. Pedro Oliveira"
						d.Email = "pedro.oliveira@example.com"
						d.Specialty = "Ortopedia"
						d.LicenseNumber = "CRM-MG-789012"
						d.Description = "Especialista em cirurgia ortopédica"
						d.Keywords = []string{"joelho", "quadril"}
					}),
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Dra. Ana Costa"
						d.Email = "ana.costa@example.com"
						d.Specialty = "Pediatria"
						d.LicenseNumber = "CRM-SP-345678"
						d.Description = "Especialista em saúde infantil"
						d.Keywords = []string{"criança", "bebê"}
					}),
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Dr. Carlos Ferreira"
						d.Email = "carlos.ferreira@example.com"
						d.Specialty = "Dermatologia"
						d.LicenseNumber = "CRM-RS-901234"
						d.Description = "Especialista em doenças de pele"
						d.Keywords = []string{"pele", "acne"}
					}),
				}
			},
			input: func() *searchinput.ListSearchInput {
				searchTerm := "Cardiologia"
				pagination, _ := cursor.NewCursorPaginationInput(nil, 10, cursor.DirectionNext)
				input, _ := searchinput.NewListSearchInput(&searchTerm, nil, nil, pagination)
				return input
			},
			expectError: false,
			validateResult: func(t *testing.T, result []*domain.Specialist, indexed []*elasticsearchtest.SpecialistDocument) {
				require.Len(t, result, 1)

				specialist := result[0]
				assert.Equal(t, "Dr. João Silva", specialist.Name)
				assert.Equal(t, "Cardiologia", specialist.Specialty)
				assert.NotEmpty(t, specialist.ID)
				assert.NotEmpty(t, specialist.Email)
				assert.NotEmpty(t, specialist.LicenseNumber)
				assert.True(t, specialist.AgreedToShare)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			client, indexName, cleanup := testHelper.SetupTestIndex(t, indexes.CreateSpecialistsIndex)
			defer cleanup()

			indexedDocs := tt.setupData()
			elasticsearchtest.IndexSpecialists(t, ctx, client, indexName, indexedDocs)

			time.Sleep(1 * time.Second)

			logger := &mockLogger{}
			repository := NewRepository(client, indexName, logger)

			input := tt.input()

			result, err := repository.Search(ctx, input)

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			require.NotNil(t, result.Specialists)
			require.NotNil(t, result.CursorOutput)

			tt.validateResult(t, result.Specialists, indexedDocs)
		})
	}
}

type mockLogger struct{}

func (m *mockLogger) Info(ctx context.Context, msg string, fields ...observability.Field)  {}
func (m *mockLogger) Error(ctx context.Context, msg string, fields ...observability.Field) {}
func (m *mockLogger) Debug(ctx context.Context, msg string, fields ...observability.Field) {}
func (m *mockLogger) Warn(ctx context.Context, msg string, fields ...observability.Field)  {}
