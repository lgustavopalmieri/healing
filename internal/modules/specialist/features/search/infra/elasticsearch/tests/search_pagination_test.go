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

func TestRepository_Search_Pagination(t *testing.T) {
	t.Run("first page returns correct page size and has next", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		client, indexName, cleanup := testHelper.SetupTestIndex(t, indexes.CreateSpecialistsIndex)
		defer cleanup()

		docs := []*elasticsearchtest.SpecialistDocument{
			elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
				d.Name = "Dr. Specialist 1"
			}),
			elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
				d.Name = "Dr. Specialist 2"
			}),
			elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
				d.Name = "Dr. Specialist 3"
			}),
			elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
				d.Name = "Dr. Specialist 4"
			}),
			elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
				d.Name = "Dr. Specialist 5"
			}),
		}
		elasticsearchtest.IndexSpecialists(t, ctx, client, indexName, docs)
		time.Sleep(1 * time.Second)

		mockLogger := observabilitymocks.NewMockLogger(ctrl)
		repository := elasticsearch.NewRepository(client, indexName, mockLogger)

		searchTerm := "Specialist"
		pagination, _ := cursor.NewCursorPaginationInput(nil, 2, cursor.DirectionNext)
		input, _ := searchinput.NewListSearchInput(&searchTerm, nil, nil, pagination)

		result, err := repository.Search(ctx, input)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Len(t, result.Specialists, 2)
		assert.True(t, result.CursorOutput.HasNextPage)
		assert.False(t, result.CursorOutput.HasPreviousPage)
		assert.NotNil(t, result.CursorOutput.NextCursor)
		assert.Nil(t, result.CursorOutput.PreviousCursor)
	})

	t.Run("second page returns correct results using cursor", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		client, indexName, cleanup := testHelper.SetupTestIndex(t, indexes.CreateSpecialistsIndex)
		defer cleanup()

		docs := []*elasticsearchtest.SpecialistDocument{
			elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
				d.Name = "Dr. Specialist 1"
			}),
			elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
				d.Name = "Dr. Specialist 2"
			}),
			elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
				d.Name = "Dr. Specialist 3"
			}),
			elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
				d.Name = "Dr. Specialist 4"
			}),
			elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
				d.Name = "Dr. Specialist 5"
			}),
		}
		elasticsearchtest.IndexSpecialists(t, ctx, client, indexName, docs)
		time.Sleep(1 * time.Second)

		mockLogger := observabilitymocks.NewMockLogger(ctrl)
		repository := elasticsearch.NewRepository(client, indexName, mockLogger)

		searchTerm := "Specialist"
		pagination1, _ := cursor.NewCursorPaginationInput(nil, 2, cursor.DirectionNext)
		input1, _ := searchinput.NewListSearchInput(&searchTerm, nil, nil, pagination1)
		result1, err := repository.Search(ctx, input1)
		require.NoError(t, err)

		firstPageNames := make(map[string]bool)
		for _, s := range result1.Specialists {
			firstPageNames[s.Name] = true
		}

		pagination2, _ := cursor.NewCursorPaginationInput(result1.CursorOutput.NextCursor, 2, cursor.DirectionNext)
		input2, _ := searchinput.NewListSearchInput(&searchTerm, nil, nil, pagination2)
		result2, err := repository.Search(ctx, input2)

		require.NoError(t, err)
		require.NotNil(t, result2)
		assert.Len(t, result2.Specialists, 2)
		assert.True(t, result2.CursorOutput.HasNextPage)
		assert.True(t, result2.CursorOutput.HasPreviousPage)

		for _, specialist := range result2.Specialists {
			assert.False(t, firstPageNames[specialist.Name], "Specialist %s should not appear in second page", specialist.Name)
		}
	})

	t.Run("last page has no next cursor", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		client, indexName, cleanup := testHelper.SetupTestIndex(t, indexes.CreateSpecialistsIndex)
		defer cleanup()

		docs := []*elasticsearchtest.SpecialistDocument{
			elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
				d.Name = "Dr. Specialist 1"
			}),
			elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
				d.Name = "Dr. Specialist 2"
			}),
			elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
				d.Name = "Dr. Specialist 3"
			}),
		}
		elasticsearchtest.IndexSpecialists(t, ctx, client, indexName, docs)
		time.Sleep(1 * time.Second)

		mockLogger := observabilitymocks.NewMockLogger(ctrl)
		repository := elasticsearch.NewRepository(client, indexName, mockLogger)

		searchTerm := "Specialist"
		pagination, _ := cursor.NewCursorPaginationInput(nil, 5, cursor.DirectionNext)
		input, _ := searchinput.NewListSearchInput(&searchTerm, nil, nil, pagination)

		result, err := repository.Search(ctx, input)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Len(t, result.Specialists, 3)
		assert.False(t, result.CursorOutput.HasNextPage)
		assert.Nil(t, result.CursorOutput.NextCursor)
	})

	t.Run("navigates through all pages correctly", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		client, indexName, cleanup := testHelper.SetupTestIndex(t, indexes.CreateSpecialistsIndex)
		defer cleanup()

		totalDocs := 10
		docs := make([]*elasticsearchtest.SpecialistDocument, totalDocs)
		for i := 0; i < totalDocs; i++ {
			docs[i] = elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
				d.Name = "Dr. Specialist"
			})
		}
		elasticsearchtest.IndexSpecialists(t, ctx, client, indexName, docs)
		time.Sleep(1 * time.Second)

		mockLogger := observabilitymocks.NewMockLogger(ctrl)
		repository := elasticsearch.NewRepository(client, indexName, mockLogger)

		searchTerm := "Specialist"
		pageSize := 3
		allSpecialists := make([]*domain.Specialist, 0)
		var nextCursor *string

		for {
			pagination, _ := cursor.NewCursorPaginationInput(nextCursor, pageSize, cursor.DirectionNext)
			input, _ := searchinput.NewListSearchInput(&searchTerm, nil, nil, pagination)
			result, err := repository.Search(ctx, input)

			require.NoError(t, err)
			allSpecialists = append(allSpecialists, result.Specialists...)

			if !result.CursorOutput.HasNextPage {
				break
			}
			nextCursor = result.CursorOutput.NextCursor
		}

		assert.Len(t, allSpecialists, totalDocs)

		seenIDs := make(map[string]bool)
		for _, specialist := range allSpecialists {
			assert.False(t, seenIDs[specialist.ID], "Duplicate specialist ID found: %s", specialist.ID)
			seenIDs[specialist.ID] = true
		}
	})

	t.Run("sorts by rating desc and updated_at desc by default", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		client, indexName, cleanup := testHelper.SetupTestIndex(t, indexes.CreateSpecialistsIndex)
		defer cleanup()

		now := time.Now()
		docs := []*elasticsearchtest.SpecialistDocument{
			elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
				d.Name = "Dr. Low Rating Old"
				d.Rating = 3.0
				d.UpdatedAt = now.Add(-2 * time.Hour)
			}),
			elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
				d.Name = "Dr. High Rating Recent"
				d.Rating = 5.0
				d.UpdatedAt = now
			}),
			elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
				d.Name = "Dr. High Rating Old"
				d.Rating = 5.0
				d.UpdatedAt = now.Add(-1 * time.Hour)
			}),
			elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
				d.Name = "Dr. Mid Rating Recent"
				d.Rating = 4.0
				d.UpdatedAt = now.Add(-30 * time.Minute)
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
		require.Len(t, result.Specialists, 4)

		assert.Equal(t, "Dr. High Rating Recent", result.Specialists[0].Name)
		assert.Equal(t, 5.0, result.Specialists[0].Rating)

		assert.Equal(t, "Dr. High Rating Old", result.Specialists[1].Name)
		assert.Equal(t, 5.0, result.Specialists[1].Rating)

		assert.Equal(t, "Dr. Mid Rating Recent", result.Specialists[2].Name)
		assert.Equal(t, 4.0, result.Specialists[2].Rating)

		assert.Equal(t, "Dr. Low Rating Old", result.Specialists[3].Name)
		assert.Equal(t, 3.0, result.Specialists[3].Rating)
	})

	t.Run("pagination with same rating uses updated_at as tiebreaker", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		client, indexName, cleanup := testHelper.SetupTestIndex(t, indexes.CreateSpecialistsIndex)
		defer cleanup()

		now := time.Now()
		docs := []*elasticsearchtest.SpecialistDocument{
			elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
				d.Name = "Dr. Same Rating 1"
				d.Rating = 4.5
				d.UpdatedAt = now.Add(-3 * time.Hour)
			}),
			elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
				d.Name = "Dr. Same Rating 2"
				d.Rating = 4.5
				d.UpdatedAt = now.Add(-2 * time.Hour)
			}),
			elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
				d.Name = "Dr. Same Rating 3"
				d.Rating = 4.5
				d.UpdatedAt = now.Add(-1 * time.Hour)
			}),
		}
		elasticsearchtest.IndexSpecialists(t, ctx, client, indexName, docs)
		time.Sleep(1 * time.Second)

		mockLogger := observabilitymocks.NewMockLogger(ctrl)
		repository := elasticsearch.NewRepository(client, indexName, mockLogger)

		searchTerm := "Same Rating"
		pagination, _ := cursor.NewCursorPaginationInput(nil, 2, cursor.DirectionNext)
		input, _ := searchinput.NewListSearchInput(&searchTerm, nil, nil, pagination)

		result, err := repository.Search(ctx, input)

		require.NoError(t, err)
		require.Len(t, result.Specialists, 2)

		assert.Equal(t, "Dr. Same Rating 3", result.Specialists[0].Name)
		assert.Equal(t, "Dr. Same Rating 2", result.Specialists[1].Name)

		pagination2, _ := cursor.NewCursorPaginationInput(result.CursorOutput.NextCursor, 2, cursor.DirectionNext)
		input2, _ := searchinput.NewListSearchInput(&searchTerm, nil, nil, pagination2)
		result2, err := repository.Search(ctx, input2)

		require.NoError(t, err)
		require.Len(t, result2.Specialists, 1)
		assert.Equal(t, "Dr. Same Rating 1", result2.Specialists[0].Name)
	})

	t.Run("no duplicates across pages with identical rating and updated_at", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		client, indexName, cleanup := testHelper.SetupTestIndex(t, indexes.CreateSpecialistsIndex)
		defer cleanup()

		now := time.Now()
		docs := make([]*elasticsearchtest.SpecialistDocument, 5)
		for i := range docs {
			docs[i] = elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
				d.Name = "Dr. Identical"
				d.Rating = 4.5
				d.UpdatedAt = now
			})
		}
		elasticsearchtest.IndexSpecialists(t, ctx, client, indexName, docs)
		time.Sleep(1 * time.Second)

		mockLogger := observabilitymocks.NewMockLogger(ctrl)
		repository := elasticsearch.NewRepository(client, indexName, mockLogger)

		searchTerm := "Identical"
		allIDs := make(map[string]bool)
		var nextCursor *string

		for {
			pagination, _ := cursor.NewCursorPaginationInput(nextCursor, 2, cursor.DirectionNext)
			input, _ := searchinput.NewListSearchInput(&searchTerm, nil, nil, pagination)
			result, err := repository.Search(ctx, input)

			require.NoError(t, err)

			for _, specialist := range result.Specialists {
				assert.False(t, allIDs[specialist.ID], "Duplicate ID found: %s", specialist.ID)
				allIDs[specialist.ID] = true
			}

			if !result.CursorOutput.HasNextPage {
				break
			}
			nextCursor = result.CursorOutput.NextCursor
		}

		assert.Len(t, allIDs, 5)
	})

	t.Run("empty result set returns no cursors", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		client, indexName, cleanup := testHelper.SetupTestIndex(t, indexes.CreateSpecialistsIndex)
		defer cleanup()

		docs := []*elasticsearchtest.SpecialistDocument{
			elasticsearchtest.SpecialistDocumentFactory(func(d *elasticsearchtest.SpecialistDocument) {
				d.Name = "Dr. João Silva"
			}),
		}
		elasticsearchtest.IndexSpecialists(t, ctx, client, indexName, docs)
		time.Sleep(1 * time.Second)

		mockLogger := observabilitymocks.NewMockLogger(ctrl)
		repository := elasticsearch.NewRepository(client, indexName, mockLogger)

		searchTerm := "NonExistent"
		pagination, _ := cursor.NewCursorPaginationInput(nil, 10, cursor.DirectionNext)
		input, _ := searchinput.NewListSearchInput(&searchTerm, nil, nil, pagination)

		result, err := repository.Search(ctx, input)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Len(t, result.Specialists, 0)
		assert.False(t, result.CursorOutput.HasNextPage)
		assert.False(t, result.CursorOutput.HasPreviousPage)
		assert.Nil(t, result.CursorOutput.NextCursor)
		assert.Nil(t, result.CursorOutput.PreviousCursor)
	})
}
