package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	"github.com/lib/pq"
)

type SpecialistUpdateRepository struct {
	db *sql.DB
}

func (r *SpecialistUpdateRepository) FindByID(ctx context.Context, id string) (*domain.Specialist, error) {
	query := `
		SELECT id, name, email, phone, specialty, license_number,
		       description, keywords, agreed_to_share, rating, status, created_at, updated_at
		FROM specialists
		WHERE id = $1`

	var specialist domain.Specialist
	var keywords pq.StringArray

	err := r.db.QueryRowContext(ctx, query, id).Scan(
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

func (r *SpecialistUpdateRepository) Update(ctx context.Context, specialist *domain.Specialist) (*domain.Specialist, error) {
	query := `
		UPDATE specialists SET
			name = $2, email = $3, phone = $4, specialty = $5, license_number = $6,
			description = $7, keywords = $8, agreed_to_share = $9, status = $10, updated_at = $11
		WHERE id = $1
		RETURNING id, name, email, phone, specialty, license_number,
		          description, keywords, agreed_to_share, rating, status, created_at, updated_at`

	var updated domain.Specialist
	var keywords pq.StringArray

	err := r.db.QueryRowContext(
		ctx,
		query,
		specialist.ID,
		specialist.Name,
		specialist.Email,
		specialist.Phone,
		specialist.Specialty,
		specialist.LicenseNumber,
		specialist.Description,
		pq.Array(specialist.Keywords),
		specialist.AgreedToShare,
		specialist.Status,
		specialist.UpdatedAt,
	).Scan(
		&updated.ID,
		&updated.Name,
		&updated.Email,
		&updated.Phone,
		&updated.Specialty,
		&updated.LicenseNumber,
		&updated.Description,
		&keywords,
		&updated.AgreedToShare,
		&updated.Rating,
		&updated.Status,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf(UpdateNotFoundErr, specialist.ID)
		}
		return nil, fmt.Errorf(FailedToUpdateErr, err)
	}

	updated.Keywords = []string(keywords)
	return &updated, nil
}
