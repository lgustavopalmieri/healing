package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	observabilitymocks "github.com/lgustavopalmieri/healing-specialist/internal/commom/observability/mocks"
	elasticsearchtest "github.com/lgustavopalmieri/healing-specialist/internal/commom/tests/elasticsearch"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/value-objects/pagination/cursor"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	searchinput "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search/search_input"
	elasticsearch "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/search/infra/elasticsearch"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/elasticsearch/indexes"
)

func TestRepository_Search_BasicSearches(t *testing.T) {
	tests := []struct {
		name           string
		setupData      func() []*elasticsearchtest.SpecialistDocument
		input          func() *searchinput.ListSearchInput
		expectError    bool
		validateResult func(*testing.T, []*domain.Specialist)
	}{
		{
			name: "returns only specialists with dr/dra prefix using wildcard",
			setupData: func() []*elasticsearchtest.SpecialistDocument {
				return []*elasticsearchtest.SpecialistDocument{
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Dr. João Silva"
						d.Specialty = "Cardiologia"
					}),
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Dra. Maria Santos"
						d.Specialty = "Neurologia"
					}),
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Carlos Ferreira"
						d.Specialty = "Dermatologia"
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
			validateResult: func(t *testing.T, result []*domain.Specialist) {
				require.Len(t, result, 2)
				for _, specialist := range result {
					assert.NotEqual(t, "Carlos Ferreira", specialist.Name)
				}
			},
		},
		{
			name: "returns specialists matching specialty using multi-match",
			setupData: func() []*elasticsearchtest.SpecialistDocument {
				return []*elasticsearchtest.SpecialistDocument{
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Dr. João Silva"
						d.Specialty = "Cardiologia"
						d.Description = "Especialista em doenças cardiovasculares"
					}),
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Dra. Maria Santos"
						d.Specialty = "Neurologia"
						d.Description = "Especialista em doenças neurológicas"
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
			validateResult: func(t *testing.T, result []*domain.Specialist) {
				require.Len(t, result, 1)
				assert.Equal(t, "Dr. João Silva", result[0].Name)
				assert.Equal(t, "Cardiologia", result[0].Specialty)
			},
		},
		{
			name: "returns specialists matching description text",
			setupData: func() []*elasticsearchtest.SpecialistDocument {
				return []*elasticsearchtest.SpecialistDocument{
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Dr. João Silva"
						d.Specialty = "Cardiologia"
						d.Description = "Especialista em cirurgia cardíaca minimamente invasiva com técnicas robóticas"
					}),
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Dra. Maria Santos"
						d.Specialty = "Neurologia"
						d.Description = "Especialista em tratamento de epilepsia refratária"
					}),
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Dr. Pedro Oliveira"
						d.Specialty = "Ortopedia"
						d.Description = "Especialista em cirurgia de joelho e quadril"
					}),
				}
			},
			input: func() *searchinput.ListSearchInput {
				searchTerm := "robóticas"
				pagination, _ := cursor.NewCursorPaginationInput(nil, 10, cursor.DirectionNext)
				input, _ := searchinput.NewListSearchInput(&searchTerm, nil, nil, pagination)
				return input
			},
			expectError: false,
			validateResult: func(t *testing.T, result []*domain.Specialist) {
				require.Len(t, result, 1)
				assert.Equal(t, "Dr. João Silva", result[0].Name)
				assert.Contains(t, result[0].Description, "robóticas")
			},
		},
		{
			name: "returns specialists matching keywords",
			setupData: func() []*elasticsearchtest.SpecialistDocument {
				return []*elasticsearchtest.SpecialistDocument{
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Dr. João Silva"
						d.Specialty = "Cardiologia"
						d.Keywords = []string{"arritmia", "marcapasso", "ablação"}
					}),
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Dra. Maria Santos"
						d.Specialty = "Neurologia"
						d.Keywords = []string{"alzheimer", "parkinson", "demência"}
					}),
				}
			},
			input: func() *searchinput.ListSearchInput {
				searchTerm := "marcapasso"
				pagination, _ := cursor.NewCursorPaginationInput(nil, 10, cursor.DirectionNext)
				input, _ := searchinput.NewListSearchInput(&searchTerm, nil, nil, pagination)
				return input
			},
			expectError: false,
			validateResult: func(t *testing.T, result []*domain.Specialist) {
				require.Len(t, result, 1)
				assert.Equal(t, "Dr. João Silva", result[0].Name)
				assert.Contains(t, result[0].Keywords, "marcapasso")
			},
		},
		{
			name: "returns empty result when no match found",
			setupData: func() []*elasticsearchtest.SpecialistDocument {
				return []*elasticsearchtest.SpecialistDocument{
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Dr. João Silva"
						d.Specialty = "Cardiologia"
					}),
				}
			},
			input: func() *searchinput.ListSearchInput {
				searchTerm := "Oftalmologia"
				pagination, _ := cursor.NewCursorPaginationInput(nil, 10, cursor.DirectionNext)
				input, _ := searchinput.NewListSearchInput(&searchTerm, nil, nil, pagination)
				return input
			},
			expectError: false,
			validateResult: func(t *testing.T, result []*domain.Specialist) {
				require.Len(t, result, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := context.Background()

			client, indexName, cleanup := testHelper.SetupTestIndex(t, indexes.CreateSpecialistsIndex)
			defer cleanup()

			indexedDocs := tt.setupData()
			elasticsearchtest.IndexSpecialists(t, ctx, client, indexName, indexedDocs)

			time.Sleep(1 * time.Second)

			mockLogger := observabilitymocks.NewMockLogger(ctrl)
			repository := elasticsearch.NewRepository(client, indexName, mockLogger)

			input := tt.input()
			result, err := repository.Search(ctx, input)

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			tt.validateResult(t, result.Specialists)
		})
	}
}
