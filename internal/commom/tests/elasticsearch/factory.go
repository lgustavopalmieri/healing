package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
)

type SpecialistDocument struct {
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
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func SpecialistDocumentFactory(overrides ...func(*SpecialistDocument)) *SpecialistDocument {
	now := time.Now()
	doc := &SpecialistDocument{
		ID:            uuid.New().String(),
		Name:          "Dr. João Silva",
		Email:         "joao.silva@example.com",
		Phone:         "+5511999999999",
		Specialty:     "Cardiologia",
		LicenseNumber: "CRM-SP-123456",
		Description:   "Cardiologista com 15 anos de experiência em doenças cardiovasculares",
		Keywords:      []string{"cardiologia", "coração", "hipertensão"},
		AgreedToShare: true,
		Rating:        4.5,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	for _, override := range overrides {
		override(doc)
	}

	return doc
}

func GetPredefinedSpecialists() []*SpecialistDocument {
	now := time.Now()

	return []*SpecialistDocument{
		{
			ID:            uuid.New().String(),
			Name:          "Dr. João Silva",
			Email:         "joao.silva@example.com",
			Phone:         "+5511999999999",
			Specialty:     "Cardiologia",
			LicenseNumber: "CRM-SP-123456",
			Description:   "Cardiologista com 15 anos de experiência em doenças cardiovasculares e arritmias",
			Keywords:      []string{"cardiologia", "coração", "hipertensão", "arritmia"},
			AgreedToShare: true,
			Rating:        4.8,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
		{
			ID:            uuid.New().String(),
			Name:          "Dra. Maria Santos",
			Email:         "maria.santos@example.com",
			Phone:         "+5511988888888",
			Specialty:     "Neurologia",
			LicenseNumber: "CRM-RJ-654321",
			Description:   "Neurologista especializada em doenças neurodegenerativas e epilepsia",
			Keywords:      []string{"neurologia", "cérebro", "alzheimer", "parkinson", "epilepsia"},
			AgreedToShare: true,
			Rating:        4.9,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
		{
			ID:            uuid.New().String(),
			Name:          "Dr. Pedro Oliveira",
			Email:         "pedro.oliveira@example.com",
			Phone:         "+5511977777777",
			Specialty:     "Ortopedia",
			LicenseNumber: "CRM-MG-789012",
			Description:   "Ortopedista especializado em cirurgia de joelho e quadril",
			Keywords:      []string{"ortopedia", "joelho", "quadril", "cirurgia", "artroscopia"},
			AgreedToShare: true,
			Rating:        4.6,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
		{
			ID:            uuid.New().String(),
			Name:          "Dra. Ana Costa",
			Email:         "ana.costa@example.com",
			Phone:         "+5511966666666",
			Specialty:     "Pediatria",
			LicenseNumber: "CRM-SP-345678",
			Description:   "Pediatra com foco em desenvolvimento infantil e vacinação",
			Keywords:      []string{"pediatria", "criança", "bebê", "vacinação", "desenvolvimento"},
			AgreedToShare: true,
			Rating:        5.0,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
		{
			ID:            uuid.New().String(),
			Name:          "Dr. Carlos Ferreira",
			Email:         "carlos.ferreira@example.com",
			Phone:         "+5511955555555",
			Specialty:     "Dermatologia",
			LicenseNumber: "CRM-RS-901234",
			Description:   "Dermatologista especializado em tratamento de acne e envelhecimento cutâneo",
			Keywords:      []string{"dermatologia", "pele", "acne", "estética", "laser"},
			AgreedToShare: true,
			Rating:        4.3,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
	}
}

func IndexSpecialists(t *testing.T, ctx context.Context, client *elasticsearch.Client, indexName string, specialists []*SpecialistDocument) {
	for _, specialist := range specialists {
		var buf bytes.Buffer
		err := json.NewEncoder(&buf).Encode(specialist)
		require.NoError(t, err)

		res, err := client.Index(
			indexName,
			&buf,
			client.Index.WithContext(ctx),
			client.Index.WithDocumentID(specialist.ID),
			client.Index.WithRefresh("true"),
		)
		require.NoError(t, err)
		defer res.Body.Close()

		require.False(t, res.IsError(), fmt.Sprintf("failed to index document: %s", res.Status()))
	}
}

func ToSpecialistEntity(doc *SpecialistDocument) *domain.Specialist {
	return &domain.Specialist{
		ID:            doc.ID,
		Name:          doc.Name,
		Email:         doc.Email,
		Phone:         doc.Phone,
		Specialty:     doc.Specialty,
		LicenseNumber: doc.LicenseNumber,
		Description:   doc.Description,
		Keywords:      doc.Keywords,
		AgreedToShare: doc.AgreedToShare,
		Rating:        doc.Rating,
		CreatedAt:     doc.CreatedAt,
		UpdatedAt:     doc.UpdatedAt,
	}
}
