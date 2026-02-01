package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	"github.com/lib/pq"
)

type SpecialistCreateRepository struct {
	db *sql.DB
}

func (r *SpecialistCreateRepository) Save(ctx context.Context, specialist *domain.Specialist) (*domain.Specialist, error) {
	query := `
		INSERT INTO specialists (
			id, name, email, phone, specialty, license_number, 
			description, keywords, agreed_to_share, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		) RETURNING id, name, email, phone, specialty, license_number, 
		           description, keywords, agreed_to_share, created_at, updated_at`

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
		&savedSpecialist.CreatedAt,
		&savedSpecialist.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf(FailedToSaveErr, err)
	}

	savedSpecialist.Keywords = []string(keywords)
	return &savedSpecialist, nil
}

func (r *SpecialistCreateRepository) ValidateUniqueness(ctx context.Context, id, email, licenseNumber string) error {
	var idExists bool
	err := r.db.QueryRowContext(
		ctx,
		"SELECT EXISTS(SELECT 1 FROM specialists WHERE id = $1)",
		id,
	).Scan(&idExists)

	if err != nil {
		return fmt.Errorf(FailedToCheckIdErr, err)
	}

	if idExists {
		return fmt.Errorf(IdAlreadyExistsErr, id)
	}

	var emailExists bool
	err = r.db.QueryRowContext(
		ctx,
		"SELECT EXISTS(SELECT 1 FROM specialists WHERE email = $1)",
		email,
	).Scan(&emailExists)

	if err != nil {
		return fmt.Errorf(FailedToCheckEmailErr, err)
	}

	if emailExists {
		return fmt.Errorf(EmailAlreadyExistsErr, email)
	}

	var licenseExists bool
	err = r.db.QueryRowContext(
		ctx,
		"SELECT EXISTS(SELECT 1 FROM specialists WHERE license_number = $1)",
		licenseNumber,
	).Scan(&licenseExists)

	if err != nil {
		return fmt.Errorf(FailedToCheckLicenseErr, err)
	}

	if licenseExists {
		return fmt.Errorf(LicenseAlreadyExistsErr, licenseNumber)
	}

	return nil
}
