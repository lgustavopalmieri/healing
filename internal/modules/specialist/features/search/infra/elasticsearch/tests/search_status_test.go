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
	searchinput "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search/search_input"
	elasticsearch "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/search/infra/elasticsearch"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/elasticsearch/indexes"
)

func TestRepository_Search_StatusFiltering(t *testing.T) {
	tests := []struct {
		name           string
		setupData      func() []*elasticsearchtest.SpecialistDocument
		input          func() *searchinput.ListSearchInput
		expectError    bool
		validateResult func(*testing.T, int)
	}{
		{
			name: "returns only active specialists and excludes pending",
			setupData: func() []*elasticsearchtest.SpecialistDocument {
				return []*elasticsearchtest.SpecialistDocument{
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Dr. João Silva"
						d.Status = "active"
					}),
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Dra. Maria Santos"
						d.Status = "pending"
					}),
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Dr. Pedro Oliveira"
						d.Status = "active"
					}),
				}
			},
			input: func() *searchinput.ListSearchInput {
				pagination, _ := cursor.NewCursorPaginationInput(nil, 10, cursor.DirectionNext)
				input, _ := searchinput.NewListSearchInput(nil, nil, nil, pagination)
				return input
			},
			expectError: false,
			validateResult: func(t *testing.T, count int) {
				assert.Equal(t, 2, count)
			},
		},
		{
			name: "excludes unavailable specialists from search results",
			setupData: func() []*elasticsearchtest.SpecialistDocument {
				return []*elasticsearchtest.SpecialistDocument{
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Dr. João Silva"
						d.Status = "active"
					}),
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Dra. Maria Santos"
						d.Status = "unavailable"
					}),
				}
			},
			input: func() *searchinput.ListSearchInput {
				pagination, _ := cursor.NewCursorPaginationInput(nil, 10, cursor.DirectionNext)
				input, _ := searchinput.NewListSearchInput(nil, nil, nil, pagination)
				return input
			},
			expectError: false,
			validateResult: func(t *testing.T, count int) {
				assert.Equal(t, 1, count)
			},
		},
		{
			name: "excludes deleted specialists from search results",
			setupData: func() []*elasticsearchtest.SpecialistDocument {
				return []*elasticsearchtest.SpecialistDocument{
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Dr. João Silva"
						d.Status = "active"
					}),
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Dra. Maria Santos"
						d.Status = "deleted"
					}),
				}
			},
			input: func() *searchinput.ListSearchInput {
				pagination, _ := cursor.NewCursorPaginationInput(nil, 10, cursor.DirectionNext)
				input, _ := searchinput.NewListSearchInput(nil, nil, nil, pagination)
				return input
			},
			expectError: false,
			validateResult: func(t *testing.T, count int) {
				assert.Equal(t, 1, count)
			},
		},
		{
			name: "excludes banned specialists from search results",
			setupData: func() []*elasticsearchtest.SpecialistDocument {
				return []*elasticsearchtest.SpecialistDocument{
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Dr. João Silva"
						d.Status = "active"
					}),
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Dra. Maria Santos"
						d.Status = "banned"
					}),
				}
			},
			input: func() *searchinput.ListSearchInput {
				pagination, _ := cursor.NewCursorPaginationInput(nil, 10, cursor.DirectionNext)
				input, _ := searchinput.NewListSearchInput(nil, nil, nil, pagination)
				return input
			},
			expectError: false,
			validateResult: func(t *testing.T, count int) {
				assert.Equal(t, 1, count)
			},
		},
		{
			name: "returns only active specialists when searching with term",
			setupData: func() []*elasticsearchtest.SpecialistDocument {
				return []*elasticsearchtest.SpecialistDocument{
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Dr. João Silva"
						d.Specialty = "Cardiologia"
						d.Status = "active"
					}),
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Dra. Maria Santos"
						d.Specialty = "Cardiologia"
						d.Status = "pending"
					}),
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Name = "Dr. Pedro Oliveira"
						d.Specialty = "Cardiologia"
						d.Status = "unavailable"
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
			validateResult: func(t *testing.T, count int) {
				assert.Equal(t, 1, count)
			},
		},
		{
			name: "returns empty when all specialists are non-active",
			setupData: func() []*elasticsearchtest.SpecialistDocument {
				return []*elasticsearchtest.SpecialistDocument{
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Status = "pending"
					}),
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Status = "unavailable"
					}),
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Status = "deleted"
					}),
					elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
						d.Status = "banned"
					}),
				}
			},
			input: func() *searchinput.ListSearchInput {
				pagination, _ := cursor.NewCursorPaginationInput(nil, 10, cursor.DirectionNext)
				input, _ := searchinput.NewListSearchInput(nil, nil, nil, pagination)
				return input
			},
			expectError: false,
			validateResult: func(t *testing.T, count int) {
				assert.Equal(t, 0, count)
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
			tt.validateResult(t, len(result.Specialists))
		})
	}
}
