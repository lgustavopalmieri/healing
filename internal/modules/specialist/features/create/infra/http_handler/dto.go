package httphandler

import (
	"time"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/application"
)

type CreateSpecialistRequest struct {
	Name          string   `json:"name"`
	Email         string   `json:"email"`
	Phone         string   `json:"phone"`
	Specialty     string   `json:"specialty"`
	LicenseNumber string   `json:"license_number"`
	Description   string   `json:"description"`
	Keywords      []string `json:"keywords"`
	AgreedToShare bool     `json:"agreed_to_share"`
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

func ToCreateSpecialistDTO(req CreateSpecialistRequest) application.CreateSpecialistDTO {
	return application.CreateSpecialistDTO{
		Name:          req.Name,
		Email:         req.Email,
		Phone:         req.Phone,
		Specialty:     req.Specialty,
		LicenseNumber: req.LicenseNumber,
		Description:   req.Description,
		Keywords:      append([]string(nil), req.Keywords...),
		AgreedToShare: req.AgreedToShare,
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
