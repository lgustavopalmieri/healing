package grpcservice

import (
	cursor "github.com/lgustavopalmieri/healing-specialist/internal/commom/value-objects/pagination/cursor"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	searchinput "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search/search_input"
	searchoutput "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search/search_output"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/search/application"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/search/infra/grpc_service/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToSearchDTO(req *pb.SearchSpecialistsRequest) (*application.SearchSpecialistsDTO, error) {
	var searchTerm *string
	if req.GetSearchTerm() != "" {
		term := req.GetSearchTerm()
		searchTerm = &term
	}

	filters := make([]searchinput.Filter, 0, len(req.GetFilters()))
	for _, f := range req.GetFilters() {
		filters = append(filters, searchinput.Filter{
			Field: searchinput.SearchableField(f.GetField()),
			Value: f.GetValue(),
		})
	}

	sorts := make([]searchinput.Sort, 0, len(req.GetSort()))
	for _, s := range req.GetSort() {
		sorts = append(sorts, searchinput.Sort{
			Field: searchinput.SearchableField(s.GetField()),
			Order: searchinput.SortOrder(s.GetOrder()),
		})
	}

	pageSize := int(req.GetPageSize())
	if pageSize <= 0 {
		pageSize = 20
	}

	direction := cursor.DirectionNext
	if req.GetDirection() == string(cursor.DirectionPrevious) {
		direction = cursor.DirectionPrevious
	}

	var encodedCursor *string
	if req.GetCursor() != "" {
		c := req.GetCursor()
		encodedCursor = &c
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

func ToSearchSpecialistsResponse(output *searchoutput.ListSearchOutput) *pb.SearchSpecialistsResponse {
	specialists := make([]*pb.Specialist, 0, len(output.Specialists))
	for _, s := range output.Specialists {
		specialists = append(specialists, ToProtoSpecialist(s))
	}

	return &pb.SearchSpecialistsResponse{
		Specialists: specialists,
		Pagination:  ToPaginationInfo(output.CursorOutput),
	}
}

func ToProtoSpecialist(s *domain.Specialist) *pb.Specialist {
	if s == nil {
		return nil
	}

	return &pb.Specialist{
		Id:            s.ID,
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
		CreatedAt:     timestamppb.New(s.CreatedAt),
		UpdatedAt:     timestamppb.New(s.UpdatedAt),
	}
}

func ToPaginationInfo(c *cursor.CursorPaginationOutput) *pb.PaginationInfo {
	if c == nil {
		return nil
	}

	info := &pb.PaginationInfo{
		HasNextPage:      c.HasNextPage,
		HasPreviousPage:  c.HasPreviousPage,
		TotalItemsInPage: int32(c.TotalItemsInPage),
	}

	if c.NextCursor != nil {
		info.NextCursor = *c.NextCursor
	}

	if c.PreviousCursor != nil {
		info.PreviousCursor = *c.PreviousCursor
	}

	return info
}
