package elasticsearch

import (
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	searchinput "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search/search_input"
)

func (r *Repository) mapSortFieldToES(field searchinput.SearchableField) string {
	switch field {
	case searchinput.FieldCreatedAt:
		return "created_at"
	case searchinput.FieldUpdatedAt:
		return "updated_at"
	case searchinput.FieldRating:
		return "rating"
	case searchinput.FieldName:
		return "name.keyword"
	case searchinput.FieldSpecialty:
		return "specialty.keyword"
	default:
		return string(field)
	}
}

func (r *Repository) mapToSpecialist(source elasticsearchSource) *domain.Specialist {
	return &domain.Specialist{
		ID:            source.ID,
		Name:          source.Name,
		Email:         source.Email,
		Phone:         source.Phone,
		Specialty:     source.Specialty,
		LicenseNumber: source.LicenseNumber,
		Description:   source.Description,
		Keywords:      source.Keywords,
		AgreedToShare: source.AgreedToShare,
		Rating:        source.Rating,
		CreatedAt:     source.CreatedAt,
		UpdatedAt:     source.UpdatedAt,
	}
}
