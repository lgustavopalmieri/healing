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

func TestRepository_Search_Filters(t *testing.T) {
	tests := []struct {
		name           string
		setupData      func() []*elasticsearchtest.SpecialistDocument
		input          func() *searchinput.ListSearchInput
		expectError    bool
		validateResult func(*testing.T, []*domain.Specialist)
	}{
		{
			name: "filters by specialty without search term",
			setupData: func() []*elasticsearchtest.SpecialistDocument {
				return []*elasticsearchtest.SpecialistDocument{
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Dr. João Silva"
						d.Specialty = "Cardiologia"
					}),
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Dra. Maria Santos"
						d.Specialty = "Cardiologia"
					}),
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Dr. Pedro Oliveira"
						d.Specialty = "Ortopedia"
					}),
				}
			},
			input: func() *searchinput.ListSearchInput {
				filters := []searchinput.Filter{
					{Field: searchinput.FieldSpecialty, Value: "Cardiologia"},
				}
				pagination, _ := cursor.NewCursorPaginationInput(nil, 10, cursor.DirectionNext)
				input, _ := searchinput.NewListSearchInput(nil, filters, nil, pagination)
				return input
			},
			expectError: false,
			validateResult: func(t *testing.T, result []*domain.Specialist) {
				require.Len(t, result, 2)
				for _, specialist := range result {
					assert.Equal(t, "Cardiologia", specialist.Specialty)
				}
			},
		},
		{
			name: "filters by name without search term",
			setupData: func() []*elasticsearchtest.SpecialistDocument {
				return []*elasticsearchtest.SpecialistDocument{
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Dr. João Silva"
						d.Specialty = "Cardiologia"
					}),
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Dr. João Santos"
						d.Specialty = "Neurologia"
					}),
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Dra. Maria Santos"
						d.Specialty = "Pediatria"
					}),
				}
			},
			input: func() *searchinput.ListSearchInput {
				filters := []searchinput.Filter{
					{Field: searchinput.FieldName, Value: "João"},
				}
				pagination, _ := cursor.NewCursorPaginationInput(nil, 10, cursor.DirectionNext)
				input, _ := searchinput.NewListSearchInput(nil, filters, nil, pagination)
				return input
			},
			expectError: false,
			validateResult: func(t *testing.T, result []*domain.Specialist) {
				require.Len(t, result, 2)
				for _, specialist := range result {
					assert.Contains(t, specialist.Name, "João")
				}
			},
		},
		{
			name: "filters by keywords without search term",
			setupData: func() []*elasticsearchtest.SpecialistDocument {
				return []*elasticsearchtest.SpecialistDocument{
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Dr. João Silva"
						d.Keywords = []string{"arritmia", "marcapasso"}
					}),
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Dra. Maria Santos"
						d.Keywords = []string{"hipertensão", "diabetes"}
					}),
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Dr. Pedro Oliveira"
						d.Keywords = []string{"arritmia", "ablação"}
					}),
				}
			},
			input: func() *searchinput.ListSearchInput {
				filters := []searchinput.Filter{
					{Field: searchinput.FieldKeywords, Value: "arritmia"},
				}
				pagination, _ := cursor.NewCursorPaginationInput(nil, 10, cursor.DirectionNext)
				input, _ := searchinput.NewListSearchInput(nil, filters, nil, pagination)
				return input
			},
			expectError: false,
			validateResult: func(t *testing.T, result []*domain.Specialist) {
				require.Len(t, result, 2)
				for _, specialist := range result {
					assert.Contains(t, specialist.Keywords, "arritmia")
				}
			},
		},
		{
			name: "filters by description without search term",
			setupData: func() []*elasticsearchtest.SpecialistDocument {
				return []*elasticsearchtest.SpecialistDocument{
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Dr. João Silva"
						d.Description = "Especialista em cirurgia cardíaca pediátrica"
					}),
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Dra. Maria Santos"
						d.Description = "Especialista em neurologia pediátrica"
					}),
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Dr. Pedro Oliveira"
						d.Description = "Especialista em ortopedia geral"
					}),
				}
			},
			input: func() *searchinput.ListSearchInput {
				filters := []searchinput.Filter{
					{Field: searchinput.FieldDescription, Value: "pediátrica"},
				}
				pagination, _ := cursor.NewCursorPaginationInput(nil, 10, cursor.DirectionNext)
				input, _ := searchinput.NewListSearchInput(nil, filters, nil, pagination)
				return input
			},
			expectError: false,
			validateResult: func(t *testing.T, result []*domain.Specialist) {
				require.Len(t, result, 2)
				for _, specialist := range result {
					assert.Contains(t, specialist.Description, "pediátrica")
				}
			},
		},
		{
			name: "combines search term with filter",
			setupData: func() []*elasticsearchtest.SpecialistDocument {
				return []*elasticsearchtest.SpecialistDocument{
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Dr. João Silva"
						d.Specialty = "Cardiologia"
						d.Description = "Especialista em arritmias"
					}),
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Dra. Maria Santos"
						d.Specialty = "Cardiologia"
						d.Description = "Especialista em insuficiência cardíaca"
					}),
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Dr. Pedro Oliveira"
						d.Specialty = "Neurologia"
						d.Description = "Especialista em arritmias cerebrais"
					}),
				}
			},
			input: func() *searchinput.ListSearchInput {
				searchTerm := "arritmias"
				filters := []searchinput.Filter{
					{Field: searchinput.FieldSpecialty, Value: "Cardiologia"},
				}
				pagination, _ := cursor.NewCursorPaginationInput(nil, 10, cursor.DirectionNext)
				input, _ := searchinput.NewListSearchInput(&searchTerm, filters, nil, pagination)
				return input
			},
			expectError: false,
			validateResult: func(t *testing.T, result []*domain.Specialist) {
				require.Len(t, result, 1)
				assert.Equal(t, "Dr. João Silva", result[0].Name)
				assert.Equal(t, "Cardiologia", result[0].Specialty)
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
