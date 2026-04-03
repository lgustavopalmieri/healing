package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/create"
	"github.com/lib/pq"
)

type SpecialistCreateRepository struct {
	db *sql.DB
}

func (r *SpecialistCreateRepository) SaveWithValidation(ctx context.Context, specialist *domain.Specialist) (*domain.Specialist, error) {
	result, err := r.save(ctx, specialist)
	if err != nil {
		return nil, r.handleSaveError(err)
	}
	return result, nil
}

func (r *SpecialistCreateRepository) save(ctx context.Context, specialist *domain.Specialist) (*domain.Specialist, error) {
	query := `
		INSERT INTO specialists (
			id, name, email, phone, specialty, license_number, 
			description, keywords, agreed_to_share, rating, status, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		) RETURNING id, name, email, phone, specialty, license_number, 
		           description, keywords, agreed_to_share, rating, status, created_at, updated_at`

	var savedSpecialist domain.Specialist
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
		specialist.Rating,
		specialist.Status,
		specialist.CreatedAt,
		specialist.UpdatedAt,
	).Scan(
		&savedSpecialist.ID,
		&savedSpecialist.Name,
		&savedSpecialist.Email,
		&savedSpecialist.Phone,
		&savedSpecialist.Specialty,
		&savedSpecialist.LicenseNumber,
		&savedSpecialist.Description,
		&keywords,
		&savedSpecialist.AgreedToShare,
		&savedSpecialist.Rating,
		&savedSpecialist.Status,
		&savedSpecialist.CreatedAt,
		&savedSpecialist.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	savedSpecialist.Keywords = []string(keywords)
	return &savedSpecialist, nil
}

func (r *SpecialistCreateRepository) handleSaveError(err error) error {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		switch pqErr.Constraint {
		case "specialists_pkey":
			return create.ErrDuplicateID
		case "specialists_email_key":
			return create.ErrDuplicateEmail
		case "specialists_license_number_key":
			return create.ErrDuplicateLicense
		}
	}
	return fmt.Errorf(FailedToSaveErr, err)
}
