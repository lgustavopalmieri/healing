package domain

import (
	"time"
)

// Specialist represents the domain entity for a medical specialist.
//
// The Specialist is a licensed healthcare professional who provides specialized medical care through the platform.
// Key characteristics in this domain:
// - Must hold a valid professional license (e.g., CRM in Brazil).
// - Agrees to share consultation reports/medical records directly with patients, empowering patients with full ownership of their health data.
// - Maintains a public-facing profile that helps patients discover and choose the right provider.
// - Operates independently, managing their own availability (via the separate Agenda service) and pricing.
// - May optionally provide an AI agent for assisted consultations in future versions.
//
// This entity focuses exclusively on immutable identity, professional credentials, and discoverability data.
// Dynamic aspects like availability, pricing, and scheduling are deliberately decoupled into dedicated services.
type Specialist struct {
	ID            string           // Unique identifier (UUID) for the specialist
	Name          string           // Full name of the specialist
	Email         string           // Email address used for authentication and notifications
	Phone         string           // Optional phone number for contact or notifications
	Specialty     string           // Primary medical specialty (e.g., "Cardiology", "Neurology")
	LicenseNumber string           // Official professional license number (required for verification)
	Description   string           // Professional bio or summary shown to patients on the profile
	Keywords      []string         // Searchable keywords/tags (e.g., "heart arrhythmia", "pediatric cardiology", "hypertension")
	AgreedToShare bool             // Explicit agreement to share medical reports with patients (required on onboarding)
	Rating        float64          // Average rating from patient reviews (0.0 to 5.0)
	Status        SpecialistStatus // Platform status (pending, active, unavailable, deleted, banned)
	CreatedAt     time.Time        // Timestamp when the specialist account was created
	UpdatedAt     time.Time        // Timestamp of the last profile update
}
