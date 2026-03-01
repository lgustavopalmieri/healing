package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	"github.com/lib/pq"
)

type ValidateLicenseRepository struct {
	DB *sql.DB
}

func (r *ValidateLicenseRepository) FindByID(ctx context.Context, id string) (*domain.Specialist, error) {
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

func (r *ValidateLicenseRepository) UpdateStatus(ctx context.Context, id string, status domain.SpecialistStatus) error {
	query := `UPDATE specialists SET status = $2, updated_at = $3 WHERE id = $1`

	result, err := r.DB.ExecContext(ctx, query, id, status, time.Now().UTC())
	if err != nil {
		return fmt.Errorf(FailedToUpdateStatusErr, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(FailedToUpdateStatusErr, err)
	}

	if rows == 0 {
		return fmt.Errorf(UpdateStatusNotFoundErr, id)
	}

	return nil
}
