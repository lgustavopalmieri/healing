package opensearch

import "time"

type opensearchSource struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	Phone         string    `json:"phone"`
	Specialty     string    `json:"specialty"`
	LicenseNumber string    `json:"license_number"`
	Description   string    `json:"description"`
	Keywords      []string  `json:"keywords"`
	AgreedToShare bool      `json:"agreed_to_share"`
	Rating        float64   `json:"rating"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
