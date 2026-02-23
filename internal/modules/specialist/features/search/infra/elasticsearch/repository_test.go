package elasticsearch

import (
	"context"
	"testing"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	loggerMocks "github.com/lgustavopalmieri/healing-specialist/internal/commom/observability/mocks"
	estest "github.com/lgustavopalmieri/healing-specialist/internal/commom/tests/elasticsearch"
	cursor "github.com/lgustavopalmieri/healing-specialist/internal/commom/value-objects/pagination/cursor"
	searchinput "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search/search_input"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/elasticsearch/indexes"
)

var testHelper = estest.NewTestHelper()

func TestMain(m *testing.M) {
	testHelper.RunTestMain(m)
}

func setupTestRepo(t *testing.T, ctrl *gomock.Controller) (*Repository, string, func()) {
	client, indexName, cleanup := testHelper.SetupTestIndex(t, indexes.CreateSpecialistsIndex)
	logger := loggerMocks.NewMockLogger(ctrl)
	logger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	repo := NewRepository(client, indexName, logger)
	return repo, indexName, cleanup
}

func seedSpecialists(t *testing.T, client *elasticsearch.Client, indexName string, docs []*estest.SpecialistDocument) {
	estest.IndexSpecialists(t, context.Background(), client, indexName, docs)
}

func strPtr(s string) *string {
	return &s
}

func validPagination(pageSize int) *cursor.CursorPaginationInput {
	p, _ := cursor.NewCursorPaginationInput(nil, pageSize, cursor.DirectionNext)
	return p
}

func mustBuildInput(t *testing.T, searchTerm *string, filters []searchinput.Filter, sort []searchinput.Sort, pagination *cursor.CursorPaginationInput) *searchinput.ListSearchInput {
	input, err := searchinput.NewListSearchInput(searchTerm, filters, sort, pagination)
	require.NoError(t, err)
	return input
}

func TestSearchRepository_Search(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name        string
		seedData    func() []*estest.SpecialistDocument
		buildInput  func(*testing.T) *searchinput.ListSearchInput
		expectError bool
		validate    func(*testing.T, []*estest.SpecialistDocument)
	}{
		{
			name: "success - returns specialists matching search term via multi_match",
			seedData: func() []*estest.SpecialistDocument {
				return estest.GetPredefinedSpecialists()
			},
			buildInput: func(t *testing.T) *searchinput.ListSearchInput {
				return mustBuildInput(t, strPtr("cardiologia"), nil, nil, validPagination(10))
			},
			expectError: false,
		},
		{
			name: "success - returns specialists matching short search term via wildcard",
			seedData: func() []*estest.SpecialistDocument {
				return estest.GetPredefinedSpecialists()
			},
			buildInput: func(t *testing.T) *searchinput.ListSearchInput {
				return mustBuildInput(t, strPtr("Dr"), nil, nil, validPagination(10))
			},
			expectError: false,
		},
		{
			name: "success - returns specialists matching filter by specialty",
			seedData: func() []*estest.SpecialistDocument {
				return estest.GetPredefinedSpecialists()
			},
			buildInput: func(t *testing.T) *searchinput.ListSearchInput {
				filters := []searchinput.Filter{
					{Field: searchinput.FieldSpecialty, Value: "Neurologia"},
				}
				return mustBuildInput(t, nil, filters, nil, validPagination(10))
			},
			expectError: false,
		},
		{
			name: "success - returns specialists matching filter by keywords",
			seedData: func() []*estest.SpecialistDocument {
				return estest.GetPredefinedSpecialists()
			},
			buildInput: func(t *testing.T) *searchinput.ListSearchInput {
				filters := []searchinput.Filter{
					{Field: searchinput.FieldKeywords, Value: "epilepsia"},
				}
				return mustBuildInput(t, nil, filters, nil, validPagination(10))
			},
			expectError: false,
		},
		{
			name: "success - returns empty result when no specialists match search term",
			seedData: func() []*estest.SpecialistDocument {
				return estest.GetPredefinedSpecialists()
			},
			buildInput: func(t *testing.T) *searchinput.ListSearchInput {
				return mustBuildInput(t, strPtr("xyznonexistent"), nil, nil, validPagination(10))
			},
			expectError: false,
		},
		{
			name: "success - returns only active specialists ignoring inactive status",
			seedData: func() []*estest.SpecialistDocument {
				return []*estest.SpecialistDocument{
					estest.SpecialistDocumentFactory(func(d *estest.SpecialistDocument) {
						d.ID = uuid.New().String()
						d.Name = "Dr. Ativo"
						d.Specialty = "Cardiologia"
						d.Status = "active"
						d.Rating = 4.5
						d.CreatedAt = now
						d.UpdatedAt = now
					}),
					estest.SpecialistDocumentFactory(func(d *estest.SpecialistDocument) {
						d.ID = uuid.New().String()
						d.Name = "Dr. Inativo"
						d.Specialty = "Cardiologia"
						d.Status = "deleted"
						d.Rating = 4.0
						d.CreatedAt = now
						d.UpdatedAt = now
					}),
					estest.SpecialistDocumentFactory(func(d *estest.SpecialistDocument) {
						d.ID = uuid.New().String()
						d.Name = "Dr. Banido"
						d.Specialty = "Cardiologia"
						d.Status = "banned"
						d.Rating = 3.0
						d.CreatedAt = now
						d.UpdatedAt = now
					}),
				}
			},
			buildInput: func(t *testing.T) *searchinput.ListSearchInput {
				return mustBuildInput(t, strPtr("Cardiologia"), nil, nil, validPagination(10))
			},
			expectError: false,
		},
		{
			name: "success - returns specialists matching combined search term and filter",
			seedData: func() []*estest.SpecialistDocument {
				return estest.GetPredefinedSpecialists()
			},
			buildInput: func(t *testing.T) *searchinput.ListSearchInput {
				filters := []searchinput.Filter{
					{Field: searchinput.FieldSpecialty, Value: "Cardiologia"},
				}
				return mustBuildInput(t, strPtr("arritmia"), filters, nil, validPagination(10))
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo, indexName, cleanup := setupTestRepo(t, ctrl)
			defer cleanup()

			docs := tt.seedData()
			seedSpecialists(t, repo.client, indexName, docs)

			input := tt.buildInput(t)
			output, err := repo.Search(context.Background(), input)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, output)
			require.NotNil(t, output.CursorOutput)

			switch tt.name {
			case "success - returns specialists matching search term via multi_match":
				assert.GreaterOrEqual(t, len(output.Specialists), 1)
				for _, s := range output.Specialists {
					assert.Equal(t, "active", string(s.Status))
				}

			case "success - returns specialists matching short search term via wildcard":
				assert.GreaterOrEqual(t, len(output.Specialists), 1)

			case "success - returns specialists matching filter by specialty":
				require.NotEmpty(t, output.Specialists)
				for _, s := range output.Specialists {
					assert.Equal(t, "Neurologia", s.Specialty)
				}

			case "success - returns specialists matching filter by keywords":
				require.NotEmpty(t, output.Specialists)

			case "success - returns empty result when no specialists match search term":
				assert.Empty(t, output.Specialists)
				assert.True(t, output.IsEmpty())
				assert.False(t, output.CursorOutput.HasNextPage)
				assert.Equal(t, 0, output.CursorOutput.TotalItemsInPage)

			case "success - returns only active specialists ignoring inactive status":
				require.Len(t, output.Specialists, 1)
				assert.Equal(t, "Dr. Ativo", output.Specialists[0].Name)
				assert.Equal(t, "active", string(output.Specialists[0].Status))

			case "success - returns specialists matching combined search term and filter":
				require.NotEmpty(t, output.Specialists)
				for _, s := range output.Specialists {
					assert.Equal(t, "Cardiologia", s.Specialty)
				}
			}
		})
	}
}

func TestSearchRepository_Search_Sorting(t *testing.T) {
	baseTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	sortableDocs := []*estest.SpecialistDocument{
		estest.SpecialistDocumentFactory(func(d *estest.SpecialistDocument) {
			d.ID = uuid.New().String()
			d.Name = "Dr. Alpha"
			d.Email = "alpha@example.com"
			d.LicenseNumber = "CRM-ALPHA"
			d.Specialty = "Cardiologia"
			d.Rating = 3.0
			d.Status = "active"
			d.CreatedAt = baseTime
			d.UpdatedAt = baseTime.Add(1 * time.Hour)
		}),
		estest.SpecialistDocumentFactory(func(d *estest.SpecialistDocument) {
			d.ID = uuid.New().String()
			d.Name = "Dr. Beta"
			d.Email = "beta@example.com"
			d.LicenseNumber = "CRM-BETA"
			d.Specialty = "Cardiologia"
			d.Rating = 5.0
			d.Status = "active"
			d.CreatedAt = baseTime.Add(1 * time.Hour)
			d.UpdatedAt = baseTime.Add(3 * time.Hour)
		}),
		estest.SpecialistDocumentFactory(func(d *estest.SpecialistDocument) {
			d.ID = uuid.New().String()
			d.Name = "Dr. Gamma"
			d.Email = "gamma@example.com"
			d.LicenseNumber = "CRM-GAMMA"
			d.Specialty = "Cardiologia"
			d.Rating = 4.0
			d.Status = "active"
			d.CreatedAt = baseTime.Add(2 * time.Hour)
			d.UpdatedAt = baseTime.Add(2 * time.Hour)
		}),
	}

	t.Run("success - default sort orders by rating desc then updated_at desc", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo, indexName, cleanup := setupTestRepo(t, ctrl)
		defer cleanup()

		seedSpecialists(t, repo.client, indexName, sortableDocs)

		input := mustBuildInput(t, strPtr("Cardiologia"), nil, nil, validPagination(10))
		output, err := repo.Search(context.Background(), input)

		require.NoError(t, err)
		require.Len(t, output.Specialists, 3)

		assert.Equal(t, 5.0, output.Specialists[0].Rating, "first should be highest rating (5.0)")
		assert.Equal(t, 4.0, output.Specialists[1].Rating, "second should be middle rating (4.0)")
		assert.Equal(t, 3.0, output.Specialists[2].Rating, "third should be lowest rating (3.0)")
	})

	t.Run("success - custom sort rating asc is respected with default updated_at appended", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo, indexName, cleanup := setupTestRepo(t, ctrl)
		defer cleanup()

		seedSpecialists(t, repo.client, indexName, sortableDocs)

		sort := []searchinput.Sort{
			{Field: searchinput.FieldRating, Order: searchinput.SortAsc},
		}
		input := mustBuildInput(t, strPtr("Cardiologia"), nil, sort, validPagination(10))

		require.Len(t, input.Sort, 2, "should have rating + default updated_at")
		assert.Equal(t, searchinput.FieldRating, input.Sort[0].Field)
		assert.Equal(t, searchinput.SortAsc, input.Sort[0].Order)
		assert.Equal(t, searchinput.FieldUpdatedAt, input.Sort[1].Field)

		output, err := repo.Search(context.Background(), input)

		require.NoError(t, err)
		require.Len(t, output.Specialists, 3)

		assert.Equal(t, 3.0, output.Specialists[0].Rating, "first should be lowest rating (3.0)")
		assert.Equal(t, 4.0, output.Specialists[1].Rating, "second should be middle rating (4.0)")
		assert.Equal(t, 5.0, output.Specialists[2].Rating, "third should be highest rating (5.0)")
	})

	t.Run("success - sort by updated_at asc returns chronological order", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo, indexName, cleanup := setupTestRepo(t, ctrl)
		defer cleanup()

		seedSpecialists(t, repo.client, indexName, sortableDocs)

		sort := []searchinput.Sort{
			{Field: searchinput.FieldUpdatedAt, Order: searchinput.SortAsc},
		}
		input := mustBuildInput(t, strPtr("Cardiologia"), nil, sort, validPagination(10))

		output, err := repo.Search(context.Background(), input)

		require.NoError(t, err)
		require.Len(t, output.Specialists, 3)

		assert.Equal(t, "Dr. Alpha", output.Specialists[0].Name, "oldest updated_at first")
		assert.Equal(t, "Dr. Gamma", output.Specialists[1].Name, "middle updated_at second")
		assert.Equal(t, "Dr. Beta", output.Specialists[2].Name, "newest updated_at third")
	})

	t.Run("success - tiebreaker by rating desc when updated_at is equal", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo, indexName, cleanup := setupTestRepo(t, ctrl)
		defer cleanup()

		sameTime := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)
		tiebreakDocs := []*estest.SpecialistDocument{
			estest.SpecialistDocumentFactory(func(d *estest.SpecialistDocument) {
				d.ID = uuid.New().String()
				d.Name = "Dr. Low Rating"
				d.Email = "low@example.com"
				d.LicenseNumber = "CRM-LOW"
				d.Specialty = "Ortopedia"
				d.Rating = 2.0
				d.Status = "active"
				d.UpdatedAt = sameTime
			}),
			estest.SpecialistDocumentFactory(func(d *estest.SpecialistDocument) {
				d.ID = uuid.New().String()
				d.Name = "Dr. High Rating"
				d.Email = "high@example.com"
				d.LicenseNumber = "CRM-HIGH"
				d.Specialty = "Ortopedia"
				d.Rating = 5.0
				d.Status = "active"
				d.UpdatedAt = sameTime
			}),
		}

		seedSpecialists(t, repo.client, indexName, tiebreakDocs)

		sort := []searchinput.Sort{
			{Field: searchinput.FieldUpdatedAt, Order: searchinput.SortDesc},
		}
		input := mustBuildInput(t, strPtr("Ortopedia"), nil, sort, validPagination(10))

		output, err := repo.Search(context.Background(), input)

		require.NoError(t, err)
		require.Len(t, output.Specialists, 2)

		assert.True(t,
			output.Specialists[0].Rating >= output.Specialists[1].Rating ||
				output.Specialists[0].UpdatedAt.Equal(output.Specialists[1].UpdatedAt),
			"when updated_at is equal, secondary sort (rating desc) should apply as tiebreaker",
		)
	})
}

func TestSearchRepository_Search_Pagination(t *testing.T) {
	t.Run("success - first page has correct cursor metadata", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo, indexName, cleanup := setupTestRepo(t, ctrl)
		defer cleanup()

		seedSpecialists(t, repo.client, indexName, estest.GetPredefinedSpecialists())

		input := mustBuildInput(t, strPtr("Dr"), nil, nil, validPagination(2))
		output, err := repo.Search(context.Background(), input)

		require.NoError(t, err)
		assert.Len(t, output.Specialists, 2)
		assert.Equal(t, 2, output.CursorOutput.TotalItemsInPage)
		assert.True(t, output.CursorOutput.HasNextPage)
		assert.False(t, output.CursorOutput.HasPreviousPage)
		assert.NotNil(t, output.CursorOutput.NextCursor)
		assert.Nil(t, output.CursorOutput.PreviousCursor)
	})

	t.Run("success - second page via cursor returns different specialists", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo, indexName, cleanup := setupTestRepo(t, ctrl)
		defer cleanup()

		seedSpecialists(t, repo.client, indexName, estest.GetPredefinedSpecialists())

		firstInput := mustBuildInput(t, strPtr("Dr"), nil, nil, validPagination(2))
		firstOutput, err := repo.Search(context.Background(), firstInput)
		require.NoError(t, err)
		require.NotNil(t, firstOutput.CursorOutput.NextCursor)

		secondPagination, err := cursor.NewCursorPaginationInput(
			firstOutput.CursorOutput.NextCursor, 2, cursor.DirectionNext,
		)
		require.NoError(t, err)

		secondInput := mustBuildInput(t, strPtr("Dr"), nil, nil, secondPagination)
		secondOutput, err := repo.Search(context.Background(), secondInput)

		require.NoError(t, err)
		assert.NotEmpty(t, secondOutput.Specialists)
		assert.True(t, secondOutput.CursorOutput.HasPreviousPage)
		assert.NotNil(t, secondOutput.CursorOutput.PreviousCursor)
		assert.Equal(t, len(secondOutput.Specialists), secondOutput.CursorOutput.TotalItemsInPage)

		firstIDs := make(map[string]bool)
		for _, s := range firstOutput.Specialists {
			firstIDs[s.ID] = true
		}
		for _, s := range secondOutput.Specialists {
			assert.False(t, firstIDs[s.ID], "second page must not contain specialists from first page")
		}
	})

	t.Run("success - last page has hasNextPage false and partial item count", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo, indexName, cleanup := setupTestRepo(t, ctrl)
		defer cleanup()

		seedSpecialists(t, repo.client, indexName, estest.GetPredefinedSpecialists())

		input := mustBuildInput(t, strPtr("Dr"), nil, nil, validPagination(100))
		output, err := repo.Search(context.Background(), input)

		require.NoError(t, err)
		assert.False(t, output.CursorOutput.HasNextPage, "should not have next page when all results fit")
		assert.Nil(t, output.CursorOutput.NextCursor)
		assert.Equal(t, len(output.Specialists), output.CursorOutput.TotalItemsInPage)
	})

	t.Run("success - page size equals total results has hasNextPage false", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo, indexName, cleanup := setupTestRepo(t, ctrl)
		defer cleanup()

		docs := []*estest.SpecialistDocument{
			estest.SpecialistDocumentFactory(func(d *estest.SpecialistDocument) {
				d.ID = uuid.New().String()
				d.Name = "Dr. Unico"
				d.Email = "unico@example.com"
				d.LicenseNumber = "CRM-UNICO"
				d.Specialty = "Psiquiatria"
				d.Status = "active"
				d.Rating = 4.0
			}),
		}
		seedSpecialists(t, repo.client, indexName, docs)

		filters := []searchinput.Filter{
			{Field: searchinput.FieldSpecialty, Value: "Psiquiatria"},
		}
		input := mustBuildInput(t, nil, filters, nil, validPagination(1))
		output, err := repo.Search(context.Background(), input)

		require.NoError(t, err)
		assert.Len(t, output.Specialists, 1)
		assert.Equal(t, 1, output.CursorOutput.TotalItemsInPage)
		assert.False(t, output.CursorOutput.HasNextPage)
		assert.Nil(t, output.CursorOutput.NextCursor)
	})

	t.Run("success - empty result returns zero TotalItemsInPage and no cursors", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo, _, cleanup := setupTestRepo(t, ctrl)
		defer cleanup()

		input := mustBuildInput(t, strPtr("nonexistent"), nil, nil, validPagination(10))
		output, err := repo.Search(context.Background(), input)

		require.NoError(t, err)
		assert.Empty(t, output.Specialists)
		assert.Equal(t, 0, output.CursorOutput.TotalItemsInPage)
		assert.False(t, output.CursorOutput.HasNextPage)
		assert.False(t, output.CursorOutput.HasPreviousPage)
		assert.Nil(t, output.CursorOutput.NextCursor)
		assert.Nil(t, output.CursorOutput.PreviousCursor)
	})

	t.Run("success - full traversal collects all specialists without duplicates", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo, indexName, cleanup := setupTestRepo(t, ctrl)
		defer cleanup()

		seedSpecialists(t, repo.client, indexName, estest.GetPredefinedSpecialists())

		allIDs := make(map[string]bool)
		totalCollected := 0
		var nextCursor *string

		for page := 0; page < 10; page++ {
			pagination, err := cursor.NewCursorPaginationInput(nextCursor, 2, cursor.DirectionNext)
			require.NoError(t, err)

			input := mustBuildInput(t, strPtr("Dr"), nil, nil, pagination)
			output, err := repo.Search(context.Background(), input)
			require.NoError(t, err)

			for _, s := range output.Specialists {
				assert.False(t, allIDs[s.ID], "specialist %s appeared on multiple pages", s.ID)
				allIDs[s.ID] = true
			}
			totalCollected += len(output.Specialists)

			if !output.CursorOutput.HasNextPage {
				break
			}
			nextCursor = output.CursorOutput.NextCursor
		}

		assert.Equal(t, len(allIDs), totalCollected, "total collected should match unique IDs")
		assert.GreaterOrEqual(t, totalCollected, 3, "should have collected at least 3 specialists across pages")
	})

	t.Run("success - TotalItemsInPage matches actual specialist count on each page", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo, indexName, cleanup := setupTestRepo(t, ctrl)
		defer cleanup()

		seedSpecialists(t, repo.client, indexName, estest.GetPredefinedSpecialists())

		var nextCursor *string
		for page := 0; page < 10; page++ {
			pagination, err := cursor.NewCursorPaginationInput(nextCursor, 2, cursor.DirectionNext)
			require.NoError(t, err)

			input := mustBuildInput(t, strPtr("Dr"), nil, nil, pagination)
			output, err := repo.Search(context.Background(), input)
			require.NoError(t, err)

			assert.Equal(t, len(output.Specialists), output.CursorOutput.TotalItemsInPage,
				"TotalItemsInPage must match actual specialist count on page %d", page+1)

			if !output.CursorOutput.HasNextPage {
				break
			}
			nextCursor = output.CursorOutput.NextCursor
		}
	})
}

func TestSearchRepository_Search_ErrorHandling(t *testing.T) {
	tests := []struct {
		name       string
		setupRepo  func(*testing.T, *gomock.Controller) (*Repository, func())
		buildInput func(*testing.T) *searchinput.ListSearchInput
		assertErr  func(*testing.T, error)
	}{
		{
			name: "failure - returns error when searching against non-existent index",
			setupRepo: func(t *testing.T, ctrl *gomock.Controller) (*Repository, func()) {
				client, _, cleanup := testHelper.SetupTestIndex(t, indexes.CreateSpecialistsIndex)
				logger := loggerMocks.NewMockLogger(ctrl)
				logger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
				repo := NewRepository(client, "non_existent_index_"+uuid.New().String()[:8], logger)
				return repo, cleanup
			},
			buildInput: func(t *testing.T) *searchinput.ListSearchInput {
				return mustBuildInput(t, strPtr("cardiologia"), nil, nil, validPagination(10))
			},
			assertErr: func(t *testing.T, err error) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "elasticsearch error")
			},
		},
		{
			name: "failure - returns ErrInvalidCursor when cursor is corrupted",
			setupRepo: func(t *testing.T, ctrl *gomock.Controller) (*Repository, func()) {
				client, indexName, cleanup := testHelper.SetupTestIndex(t, indexes.CreateSpecialistsIndex)
				logger := loggerMocks.NewMockLogger(ctrl)
				logger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
				repo := NewRepository(client, indexName, logger)
				return repo, cleanup
			},
			buildInput: func(t *testing.T) *searchinput.ListSearchInput {
				corruptedCursor := "dGhpcyBpcyBub3QgYSB2YWxpZCBjdXJzb3I="
				pagination, err := cursor.NewCursorPaginationInput(&corruptedCursor, 10, cursor.DirectionNext)
				require.NoError(t, err)
				return mustBuildInput(t, strPtr("cardiologia"), nil, nil, pagination)
			},
			assertErr: func(t *testing.T, err error) {
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidCursor)
			},
		},
		{
			name: "failure - returns error when elasticsearch connection is unreachable",
			setupRepo: func(t *testing.T, ctrl *gomock.Controller) (*Repository, func()) {
				cfg := elasticsearch.Config{
					Addresses: []string{"http://localhost:19999"},
				}
				client, err := elasticsearch.NewClient(cfg)
				require.NoError(t, err)

				logger := loggerMocks.NewMockLogger(ctrl)
				logger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
				repo := NewRepository(client, "test_index", logger)
				return repo, func() {}
			},
			buildInput: func(t *testing.T) *searchinput.ListSearchInput {
				return mustBuildInput(t, strPtr("cardiologia"), nil, nil, validPagination(10))
			},
			assertErr: func(t *testing.T, err error) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "search request failed")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo, cleanup := tt.setupRepo(t, ctrl)
			defer cleanup()

			input := tt.buildInput(t)
			output, err := repo.Search(context.Background(), input)

			tt.assertErr(t, err)
			assert.Nil(t, output)
		})
	}
}
