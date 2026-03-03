package httphandler

import (
	"time"

	cursor "github.com/lgustavopalmieri/healing-specialist/internal/commom/value-objects/pagination/cursor"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	searchinput "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search/search_input"
	searchoutput "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search/search_output"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/search/application"
)

type SearchSpecialistsRequest struct {
	SearchTerm string          `json:"search_term"`
	Filters    []FilterRequest `json:"filters"`
	Sort       []SortRequest   `json:"sort"`
	PageSize   int             `json:"page_size"`
	Cursor     string          `json:"cursor"`
	Direction  string          `json:"direction"`
}

type FilterRequest struct {
	Field string `json:"field"`
	Value string `json:"value"`
}

type SortRequest struct {
	Field string `json:"field"`
	Order string `json:"order"`
}

type SpecialistResponse struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Email         string   `json:"email"`
	Phone         string   `json:"phone"`
	Specialty     string   `json:"specialty"`
	LicenseNumber string   `json:"license_number"`
	Description   string   `json:"description"`
	Keywords      []string `json:"keywords"`
	AgreedToShare bool     `json:"agreed_to_share"`
	Rating        float64  `json:"rating"`
	Status        string   `json:"status"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
}

type PaginationResponse struct {
	NextCursor      string `json:"next_cursor"`
	PreviousCursor  string `json:"previous_cursor"`
	HasNextPage     bool   `json:"has_next_page"`
	HasPreviousPage bool   `json:"has_previous_page"`
	TotalItems      int    `json:"total_items_in_page"`
}

type SearchSpecialistsResponse struct {
	Specialists []SpecialistResponse `json:"specialists"`
	Pagination  PaginationResponse   `json:"pagination"`
}

func ToSearchDTO(req SearchSpecialistsRequest) (*application.SearchSpecialistsDTO, error) {
	var searchTerm *string
	if req.SearchTerm != "" {
		searchTerm = &req.SearchTerm
	}

	filters := make([]searchinput.Filter, 0, len(req.Filters))
	for _, f := range req.Filters {
		filters = append(filters, searchinput.Filter{
			Field: searchinput.SearchableField(f.Field),
			Value: f.Value,
		})
	}

	sorts := make([]searchinput.Sort, 0, len(req.Sort))
	for _, s := range req.Sort {
		sorts = append(sorts, searchinput.Sort{
			Field: searchinput.SearchableField(s.Field),
			Order: searchinput.SortOrder(s.Order),
		})
	}

	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	direction := cursor.DirectionNext
	if req.Direction == string(cursor.DirectionPrevious) {
		direction = cursor.DirectionPrevious
	}

	var encodedCursor *string
	if req.Cursor != "" {
		encodedCursor = &req.Cursor
	}

	pagination, err := cursor.NewCursorPaginationInput(encodedCursor, pageSize, direction)
	if err != nil {
		return nil, err
	}

	return &application.SearchSpecialistsDTO{
		SearchTerm: searchTerm,
		Filters:    filters,
		Sort:       sorts,
		Pagination: pagination,
	}, nil
}

func ToSearchResponse(output *searchoutput.ListSearchOutput) SearchSpecialistsResponse {
	specialists := make([]SpecialistResponse, 0, len(output.Specialists))
	for _, s := range output.Specialists {
		specialists = append(specialists, ToSpecialistResponse(s))
	}

	pagination := PaginationResponse{
		HasNextPage:     output.CursorOutput.HasNextPage,
		HasPreviousPage: output.CursorOutput.HasPreviousPage,
		TotalItems:      output.CursorOutput.TotalItemsInPage,
	}

	if output.CursorOutput.NextCursor != nil {
		pagination.NextCursor = *output.CursorOutput.NextCursor
	}

	if output.CursorOutput.PreviousCursor != nil {
		pagination.PreviousCursor = *output.CursorOutput.PreviousCursor
	}

	return SearchSpecialistsResponse{
		Specialists: specialists,
		Pagination:  pagination,
	}
}

func ToSpecialistResponse(s *domain.Specialist) SpecialistResponse {
	return SpecialistResponse{
		ID:            s.ID,
		Name:          s.Name,
		Email:         s.Email,
		Phone:         s.Phone,
		Specialty:     s.Specialty,
		LicenseNumber: s.LicenseNumber,
		Description:   s.Description,
		Keywords:      append([]string(nil), s.Keywords...),
		AgreedToShare: s.AgreedToShare,
		Rating:        s.Rating,
		Status:        string(s.Status),
		CreatedAt:     s.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     s.UpdatedAt.Format(time.RFC3339),
	}
}
