package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
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

	res, err := r.Client.Index(
		r.IndexName,
		bytes.NewReader(body),
		r.Client.Index.WithContext(ctx),
		r.Client.Index.WithDocumentID(specialist.ID),
	)
	if err != nil {
		r.Logger.Error(ctx, "elasticsearch index request failed",
			observability.Field{Key: "id", Value: specialist.ID},
			observability.Field{Key: "error", Value: err.Error()})
		return fmt.Errorf(FailedToIndexErr, err)
	}
	defer res.Body.Close()

	if res.IsError() {
		respBody, _ := io.ReadAll(res.Body)
		r.Logger.Error(ctx, "elasticsearch returned error on index",
			observability.Field{Key: "id", Value: specialist.ID},
			observability.Field{Key: "status", Value: res.Status()},
			observability.Field{Key: "body", Value: string(respBody)})
		return fmt.Errorf(IndexErrorResponseErr, res.Status())
	}

	return nil
}
