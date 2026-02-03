package elasticsearch

import "time"

type elasticsearchResponse struct {
	Hits struct {
		Hits []elasticsearchHit `json:"hits"`
	} `json:"hits"`
}

type elasticsearchHit struct {
	Source elasticsearchSource `json:"_source"`
	Sort   []interface{}       `json:"sort"`
}

type elasticsearchSource struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	Phone         string    `json:"phone"`
	Specialty     string    `json:"specialty"`
	LicenseNumber string    `json:"license_number"`
	Description   string    `json:"description"`
	Keywords      []string  `json:"keywords"`
	AgreedToShare bool      `json:"agreed_to_share"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
