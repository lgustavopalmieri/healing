package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	"github.com/lib/pq"
)

type SourceRepository struct {
	DB *sql.DB
}

func (r *SourceRepository) FindByID(ctx context.Context, id string) (*domain.Specialist, error) {
	query := `
		SELECT id, name, email, phone, specialty, license_number,
		       description, keywords, agreed_to_share, rating, status, created_at, updated_at
		FROM specialists
		WHERE id = $1`

	var specialist domain.Specialist
	var keywords pq.StringArray

	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&specialist.ID,
		&specialist.Name,
		&specialist.Email,
		&specialist.Phone,
		&specialist.Specialty,
		&specialist.LicenseNumber,
		&specialist.Description,
		&keywords,
		&specialist.AgreedToShare,
		&specialist.Rating,
		&specialist.Status,
		&specialist.CreatedAt,
		&specialist.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf(SpecialistNotFoundErr, id)
		}
		return nil, fmt.Errorf(FailedToFindByIDErr, err)
	}

	specialist.Keywords = []string(keywords)
	return &specialist, nil
}
