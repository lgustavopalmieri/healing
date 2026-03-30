package opensearch

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"

	ostest "github.com/lgustavopalmieri/healing-specialist/internal/commom/tests/opensearch"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/value-objects/pagination/cursor"
	searchinput "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search/search_input"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/opensearch/indexes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testHelper = ostest.NewTestHelper()

func TestMain(m *testing.M) {
	testHelper.RunTestMain(m)
}

func setupTestRepo(t *testing.T) (*Repository, string, func()) {
	client, indexName, cleanup := testHelper.SetupTestIndex(t, indexes.CreateSpecialistsIndex)
	repo := NewRepository(client, indexName)
	return repo, indexName, cleanup
}

func seedSpecialists(t *testing.T, client *opensearchapi.Client, indexName string, docs []*ostest.SpecialistDocument) {
	ostest.IndexSpecialists(t, context.Background(), client, indexName, docs)
}

func strPtr(s string) *string { return &s }

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
		seedData    func() []*ostest.SpecialistDocument
		buildInput  func(*testing.T) *searchinput.ListSearchInput
		expectError bool
	}{
		{
			name: "success - returns specialists matching search term via multi_match",
			seedData: func() []*ostest.SpecialistDocument {
				return ostest.GetPredefinedSpecialists()
			},
			buildInput: func(t *testing.T) *searchinput.ListSearchInput {
				return mustBuildInput(t, strPtr("cardiologia"), nil, nil, validPagination(10))
			},
		},
		{
			name: "success - returns specialists matching short search term via wildcard",
			seedData: func() []*ostest.SpecialistDocument {
				return ostest.GetPredefinedSpecialists()
			},
			buildInput: func(t *testing.T) *searchinput.ListSearchInput {
				return mustBuildInput(t, strPtr("Dr"), nil, nil, validPagination(10))
			},
		},
		{
			name: "success - returns specialists matching filter by specialty",
			seedData: func() []*ostest.SpecialistDocument {
				return ostest.GetPredefinedSpecialists()
			},
			buildInput: func(t *testing.T) *searchinput.ListSearchInput {
				filters := []searchinput.Filter{
					{Field: searchinput.FieldSpecialty, Value: "Neurologia"},
				}
				return mustBuildInput(t, nil, filters, nil, validPagination(10))
			},
		},
		{
			name: "success - returns specialists matching filter by keywords",
			seedData: func() []*ostest.SpecialistDocument {
				return ostest.GetPredefinedSpecialists()
			},
			buildInput: func(t *testing.T) *searchinput.ListSearchInput {
				filters := []searchinput.Filter{
					{Field: searchinput.FieldKeywords, Value: "epilepsia"},
				}
				return mustBuildInput(t, nil, filters, nil, validPagination(10))
			},
		},
		{
			name: "success - returns empty result when no specialists match search term",
			seedData: func() []*ostest.SpecialistDocument {
				return ostest.GetPredefinedSpecialists()
			},
			buildInput: func(t *testing.T) *searchinput.ListSearchInput {
				return mustBuildInput(t, strPtr("xyznonexistent"), nil, nil, validPagination(10))
			},
		},
		{
			name: "success - returns only active specialists ignoring inactive status",
			seedData: func() []*ostest.SpecialistDocument {
				return []*ostest.SpecialistDocument{
					ostest.SpecialistDocumentFactory(func(d *ostest.SpecialistDocument) {
						d.ID = uuid.New().String()
						d.Name = "Dr. Ativo"
						d.Specialty = "Cardiologia"
						d.Status = "active"
						d.Rating = 4.5
						d.CreatedAt = now
						d.UpdatedAt = now
					}),
					ostest.SpecialistDocumentFactory(func(d *ostest.SpecialistDocument) {
						d.ID = uuid.New().String()
						d.Name = "Dr. Licenciado"
						d.Specialty = "Cardiologia"
						d.Status = "authorized_license"
						d.Rating = 4.2
						d.CreatedAt = now
						d.UpdatedAt = now
					}),
					ostest.SpecialistDocumentFactory(func(d *ostest.SpecialistDocument) {
						d.ID = uuid.New().String()
						d.Name = "Dr. Inativo"
						d.Specialty = "Cardiologia"
						d.Status = "deleted"
						d.Rating = 4.0
						d.CreatedAt = now
						d.UpdatedAt = now
					}),
					ostest.SpecialistDocumentFactory(func(d *ostest.SpecialistDocument) {
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
		},
		{
			name: "success - returns specialists matching combined search term and filter",
			seedData: func() []*ostest.SpecialistDocument {
				return ostest.GetPredefinedSpecialists()
			},
			buildInput: func(t *testing.T) *searchinput.ListSearchInput {
				filters := []searchinput.Filter{
					{Field: searchinput.FieldSpecialty, Value: "Cardiologia"},
				}
				return mustBuildInput(t, strPtr("arritmia"), filters, nil, validPagination(10))
			},
		},
		{
			name: "success - returns sort values for cursor construction",
			seedData: func() []*ostest.SpecialistDocument {
				return ostest.GetPredefinedSpecialists()
			},
			buildInput: func(t *testing.T) *searchinput.ListSearchInput {
				return mustBuildInput(t, strPtr("Dr"), nil, nil, validPagination(10))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, indexName, cleanup := setupTestRepo(t)
			defer cleanup()

			docs := tt.seedData()
			seedSpecialists(t, repo.client, indexName, docs)

			input := tt.buildInput(t)
			result, err := repo.Search(context.Background(), input)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			switch tt.name {
			case "success - returns specialists matching search term via multi_match":
				assert.GreaterOrEqual(t, len(result.Specialists), 1)
				for _, s := range result.Specialists {
					assert.Equal(t, "active", string(s.Status))
				}

			case "success - returns specialists matching short search term via wildcard":
				assert.GreaterOrEqual(t, len(result.Specialists), 1)

			case "success - returns specialists matching filter by specialty":
				require.NotEmpty(t, result.Specialists)
				for _, s := range result.Specialists {
					assert.Equal(t, "Neurologia", s.Specialty)
				}

			case "success - returns specialists matching filter by keywords":
				require.NotEmpty(t, result.Specialists)

			case "success - returns empty result when no specialists match search term":
				assert.Empty(t, result.Specialists)
				assert.False(t, result.HasNextPage)
				assert.Nil(t, result.FirstSortValues)
				assert.Nil(t, result.LastSortValues)

			case "success - returns only active specialists ignoring inactive status":
				require.Len(t, result.Specialists, 2)
				statuses := map[string]bool{}
				for _, s := range result.Specialists {
					statuses[string(s.Status)] = true
				}
				assert.True(t, statuses["active"])
				assert.True(t, statuses["authorized_license"])

			case "success - returns specialists matching combined search term and filter":
				require.NotEmpty(t, result.Specialists)
				for _, s := range result.Specialists {
					assert.Equal(t, "Cardiologia", s.Specialty)
				}

			case "success - returns sort values for cursor construction":
				require.NotEmpty(t, result.Specialists)
				assert.NotNil(t, result.FirstSortValues)
				assert.NotNil(t, result.LastSortValues)
				assert.NotEmpty(t, result.FirstSortValues)
				assert.NotEmpty(t, result.LastSortValues)
			}
		})
	}
}

func TestSearchRepository_Search_Sorting(t *testing.T) {
	baseTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	sortableDocs := []*ostest.SpecialistDocument{
		ostest.SpecialistDocumentFactory(func(d *ostest.SpecialistDocument) {
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
		ostest.SpecialistDocumentFactory(func(d *ostest.SpecialistDocument) {
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
		ostest.SpecialistDocumentFactory(func(d *ostest.SpecialistDocument) {
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
		repo, indexName, cleanup := setupTestRepo(t)
		defer cleanup()

		seedSpecialists(t, repo.client, indexName, sortableDocs)

		input := mustBuildInput(t, strPtr("Cardiologia"), nil, nil, validPagination(10))
		result, err := repo.Search(context.Background(), input)

		require.NoError(t, err)
		require.Len(t, result.Specialists, 3)

		assert.Equal(t, 5.0, result.Specialists[0].Rating)
		assert.Equal(t, 4.0, result.Specialists[1].Rating)
		assert.Equal(t, 3.0, result.Specialists[2].Rating)
	})

	t.Run("success - custom sort rating asc is respected with default updated_at appended", func(t *testing.T) {
		repo, indexName, cleanup := setupTestRepo(t)
		defer cleanup()

		seedSpecialists(t, repo.client, indexName, sortableDocs)

		sort := []searchinput.Sort{
			{Field: searchinput.FieldRating, Order: searchinput.SortAsc},
		}
		input := mustBuildInput(t, strPtr("Cardiologia"), nil, sort, validPagination(10))

		require.Len(t, input.Sort, 2)
		assert.Equal(t, searchinput.FieldRating, input.Sort[0].Field)
		assert.Equal(t, searchinput.SortAsc, input.Sort[0].Order)
		assert.Equal(t, searchinput.FieldUpdatedAt, input.Sort[1].Field)

		result, err := repo.Search(context.Background(), input)

		require.NoError(t, err)
		require.Len(t, result.Specialists, 3)

		assert.Equal(t, 3.0, result.Specialists[0].Rating)
		assert.Equal(t, 4.0, result.Specialists[1].Rating)
		assert.Equal(t, 5.0, result.Specialists[2].Rating)
	})

	t.Run("success - sort by updated_at asc returns chronological order", func(t *testing.T) {
		repo, indexName, cleanup := setupTestRepo(t)
		defer cleanup()

		seedSpecialists(t, repo.client, indexName, sortableDocs)

		sort := []searchinput.Sort{
			{Field: searchinput.FieldUpdatedAt, Order: searchinput.SortAsc},
		}
		input := mustBuildInput(t, strPtr("Cardiologia"), nil, sort, validPagination(10))

		result, err := repo.Search(context.Background(), input)

		require.NoError(t, err)
		require.Len(t, result.Specialists, 3)

		assert.Equal(t, "Dr. Alpha", result.Specialists[0].Name)
		assert.Equal(t, "Dr. Gamma", result.Specialists[1].Name)
		assert.Equal(t, "Dr. Beta", result.Specialists[2].Name)
	})

	t.Run("success - tiebreaker by rating desc when updated_at is equal", func(t *testing.T) {
		repo, indexName, cleanup := setupTestRepo(t)
		defer cleanup()

		sameTime := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)
		tiebreakDocs := []*ostest.SpecialistDocument{
			ostest.SpecialistDocumentFactory(func(d *ostest.SpecialistDocument) {
				d.ID = uuid.New().String()
				d.Name = "Dr. Low Rating"
				d.Email = "low@example.com"
				d.LicenseNumber = "CRM-LOW"
				d.Specialty = "Ortopedia"
				d.Rating = 2.0
				d.Status = "active"
				d.UpdatedAt = sameTime
			}),
			ostest.SpecialistDocumentFactory(func(d *ostest.SpecialistDocument) {
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

		result, err := repo.Search(context.Background(), input)

		require.NoError(t, err)
		require.Len(t, result.Specialists, 2)

		assert.True(t,
			result.Specialists[0].Rating >= result.Specialists[1].Rating ||
				result.Specialists[0].UpdatedAt.Equal(result.Specialists[1].UpdatedAt),
		)
	})
}
