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

func TestRepository_Search_Sorting(t *testing.T) {
	t.Run("sorts by created_at descending by default", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		client, indexName, cleanup := testHelper.SetupTestIndex(t, indexes.CreateSpecialistsIndex)
		defer cleanup()

		now := time.Now()
		docs := []*elasticsearchtest.SpecialistDocument{
			elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
				d.Name = "Dr. First"
				d.CreatedAt = now.Add(-3 * time.Hour)
			}),
			elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
				d.Name = "Dr. Third"
				d.CreatedAt = now.Add(-1 * time.Hour)
			}),
			elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
				d.Name = "Dr. Second"
				d.CreatedAt = now.Add(-2 * time.Hour)
			}),
		}
		elasticsearchtest.IndexSpecialists(t, ctx, client, indexName, docs)
		time.Sleep(1 * time.Second)

		mockLogger := observabilitymocks.NewMockLogger(ctrl)
		repository := elasticsearch.NewRepository(client, indexName, mockLogger)

		searchTerm := "Dr"
		pagination, _ := cursor.NewCursorPaginationInput(nil, 10, cursor.DirectionNext)
		input, _ := searchinput.NewListSearchInput(&searchTerm, nil, nil, pagination)

		result, err := repository.Search(ctx, input)

		require.NoError(t, err)
		require.Len(t, result.Specialists, 3)
		assert.Equal(t, "Dr. Third", result.Specialists[0].Name)
		assert.Equal(t, "Dr. Second", result.Specialists[1].Name)
		assert.Equal(t, "Dr. First", result.Specialists[2].Name)
	})

	t.Run("sorts by created_at ascending", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		client, indexName, cleanup := testHelper.SetupTestIndex(t, indexes.CreateSpecialistsIndex)
		defer cleanup()

		now := time.Now()
		docs := []*elasticsearchtest.SpecialistDocument{
			elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
				d.Name = "Dr. Third"
				d.CreatedAt = now.Add(-1 * time.Hour)
			}),
			elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
				d.Name = "Dr. First"
				d.CreatedAt = now.Add(-3 * time.Hour)
			}),
			elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
				d.Name = "Dr. Second"
				d.CreatedAt = now.Add(-2 * time.Hour)
			}),
		}
		elasticsearchtest.IndexSpecialists(t, ctx, client, indexName, docs)
		time.Sleep(1 * time.Second)

		mockLogger := observabilitymocks.NewMockLogger(ctrl)
		repository := elasticsearch.NewRepository(client, indexName, mockLogger)

		searchTerm := "Dr"
		sort := []searchinput.Sort{
			{Field: searchinput.FieldCreatedAt, Order: searchinput.SortAsc},
		}
		pagination, _ := cursor.NewCursorPaginationInput(nil, 10, cursor.DirectionNext)
		input, _ := searchinput.NewListSearchInput(&searchTerm, nil, sort, pagination)

		result, err := repository.Search(ctx, input)

		require.NoError(t, err)
		require.Len(t, result.Specialists, 3)
		assert.Equal(t, "Dr. First", result.Specialists[0].Name)
		assert.Equal(t, "Dr. Second", result.Specialists[1].Name)
		assert.Equal(t, "Dr. Third", result.Specialists[2].Name)
	})

	t.Run("sorts by name ascending", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		client, indexName, cleanup := testHelper.SetupTestIndex(t, indexes.CreateSpecialistsIndex)
		defer cleanup()

		docs := []*elasticsearchtest.SpecialistDocument{
			elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
				d.Name = "Dr. Carlos"
			}),
			elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
				d.Name = "Dr. Ana"
			}),
			elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
				d.Name = "Dr. Bruno"
			}),
		}
		elasticsearchtest.IndexSpecialists(t, ctx, client, indexName, docs)
		time.Sleep(1 * time.Second)

		mockLogger := observabilitymocks.NewMockLogger(ctrl)
		repository := elasticsearch.NewRepository(client, indexName, mockLogger)

		searchTerm := "Dr"
		sort := []searchinput.Sort{
			{Field: searchinput.FieldName, Order: searchinput.SortAsc},
		}
		pagination, _ := cursor.NewCursorPaginationInput(nil, 10, cursor.DirectionNext)
		input, _ := searchinput.NewListSearchInput(&searchTerm, nil, sort, pagination)

		result, err := repository.Search(ctx, input)

		require.NoError(t, err)
		require.Len(t, result.Specialists, 3)
		assert.Equal(t, "Dr. Ana", result.Specialists[0].Name)
		assert.Equal(t, "Dr. Bruno", result.Specialists[1].Name)
		assert.Equal(t, "Dr. Carlos", result.Specialists[2].Name)
	})

	t.Run("sorts by name descending", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		client, indexName, cleanup := testHelper.SetupTestIndex(t, indexes.CreateSpecialistsIndex)
		defer cleanup()

		docs := []*elasticsearchtest.SpecialistDocument{
			elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
				d.Name = "Dr. Ana"
			}),
			elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
				d.Name = "Dr. Carlos"
			}),
			elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
				d.Name = "Dr. Bruno"
			}),
		}
		elasticsearchtest.IndexSpecialists(t, ctx, client, indexName, docs)
		time.Sleep(1 * time.Second)

		mockLogger := observabilitymocks.NewMockLogger(ctrl)
		repository := elasticsearch.NewRepository(client, indexName, mockLogger)

		searchTerm := "Dr"
		sort := []searchinput.Sort{
			{Field: searchinput.FieldName, Order: searchinput.SortDesc},
		}
		pagination, _ := cursor.NewCursorPaginationInput(nil, 10, cursor.DirectionNext)
		input, _ := searchinput.NewListSearchInput(&searchTerm, nil, sort, pagination)

		result, err := repository.Search(ctx, input)

		require.NoError(t, err)
		require.Len(t, result.Specialists, 3)
		assert.Equal(t, "Dr. Carlos", result.Specialists[0].Name)
		assert.Equal(t, "Dr. Bruno", result.Specialists[1].Name)
		assert.Equal(t, "Dr. Ana", result.Specialists[2].Name)
	})

	t.Run("sorts by specialty ascending", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		client, indexName, cleanup := testHelper.SetupTestIndex(t, indexes.CreateSpecialistsIndex)
		defer cleanup()

		docs := []*elasticsearchtest.SpecialistDocument{
			elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
				d.Name = "Dr. João"
				d.Specialty = "Ortopedia"
			}),
			elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
				d.Name = "Dr. Maria"
				d.Specialty = "Cardiologia"
			}),
			elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
				d.Name = "Dr. Pedro"
				d.Specialty = "Neurologia"
			}),
		}
		elasticsearchtest.IndexSpecialists(t, ctx, client, indexName, docs)
		time.Sleep(1 * time.Second)

		mockLogger := observabilitymocks.NewMockLogger(ctrl)
		repository := elasticsearch.NewRepository(client, indexName, mockLogger)

		searchTerm := "Dr"
		sort := []searchinput.Sort{
			{Field: searchinput.FieldSpecialty, Order: searchinput.SortAsc},
		}
		pagination, _ := cursor.NewCursorPaginationInput(nil, 10, cursor.DirectionNext)
		input, _ := searchinput.NewListSearchInput(&searchTerm, nil, sort, pagination)

		result, err := repository.Search(ctx, input)

		require.NoError(t, err)
		require.Len(t, result.Specialists, 3)
		assert.Equal(t, "Cardiologia", result.Specialists[0].Specialty)
		assert.Equal(t, "Neurologia", result.Specialists[1].Specialty)
		assert.Equal(t, "Ortopedia", result.Specialists[2].Specialty)
	})

	t.Run("pagination with sorting maintains order across pages", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		client, indexName, cleanup := testHelper.SetupTestIndex(t, indexes.CreateSpecialistsIndex)
		defer cleanup()

		docs := []*elasticsearchtest.SpecialistDocument{
			elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
				d.Name = "Dr. Eduardo"
			}),
			elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
				d.Name = "Dr. Ana"
			}),
			elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
				d.Name = "Dr. Carlos"
			}),
			elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
				d.Name = "Dr. Bruno"
			}),
			elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
				d.Name = "Dr. Diana"
			}),
		}
		elasticsearchtest.IndexSpecialists(t, ctx, client, indexName, docs)
		time.Sleep(1 * time.Second)

		mockLogger := observabilitymocks.NewMockLogger(ctrl)
		repository := elasticsearch.NewRepository(client, indexName, mockLogger)

		searchTerm := "Dr"
		sort := []searchinput.Sort{
			{Field: searchinput.FieldName, Order: searchinput.SortAsc},
		}

		pagination1, _ := cursor.NewCursorPaginationInput(nil, 2, cursor.DirectionNext)
		input1, _ := searchinput.NewListSearchInput(&searchTerm, nil, sort, pagination1)
		result1, err := repository.Search(ctx, input1)
		require.NoError(t, err)

		pagination2, _ := cursor.NewCursorPaginationInput(result1.CursorOutput.NextCursor, 2, cursor.DirectionNext)
		input2, _ := searchinput.NewListSearchInput(&searchTerm, nil, sort, pagination2)
		result2, err := repository.Search(ctx, input2)
		require.NoError(t, err)

		allNames := []string{
			result1.Specialists[0].Name,
			result1.Specialists[1].Name,
			result2.Specialists[0].Name,
			result2.Specialists[1].Name,
		}

		assert.Equal(t, "Dr. Ana", allNames[0])
		assert.Equal(t, "Dr. Bruno", allNames[1])
		assert.Equal(t, "Dr. Carlos", allNames[2])
		assert.Equal(t, "Dr. Diana", allNames[3])
	})
}
