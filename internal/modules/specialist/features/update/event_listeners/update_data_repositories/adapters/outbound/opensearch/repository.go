package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
)

func (r *Repository) Update(ctx context.Context, specialist *domain.Specialist) error {
	doc := map[string]any{
		"id":              specialist.ID,
		"name":            specialist.Name,
		"email":           specialist.Email,
		"phone":           specialist.Phone,
		"specialty":       specialist.Specialty,
		"license_number":  specialist.LicenseNumber,
		"description":     specialist.Description,
		"keywords":        specialist.Keywords,
		"agreed_to_share": specialist.AgreedToShare,
		"rating":          specialist.Rating,
		"status":          string(specialist.Status),
		"created_at":      specialist.CreatedAt,
		"updated_at":      specialist.UpdatedAt,
	}

	body, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf(FailedToSerializeErr, err)
	}

	_, err = r.Client.Index(
		ctx,
		opensearchapi.IndexReq{
			Index:      r.IndexName,
			DocumentID: specialist.ID,
			Body:       bytes.NewReader(body),
		},
	)
	if err != nil {
		return fmt.Errorf(FailedToIndexErr, err)
	}

	return nil
}
