package elasticsearch

import (
	"context"
	"fmt"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
)

func (r *Repository) PublishDLQ(ctx context.Context, specialist *domain.Specialist, reason error) error {
	dlqEvent := event.NewEvent(ElasticsearchUpdateDLQEventName, map[string]any{
		"id":     specialist.ID,
		"reason": reason.Error(),
		"source": "elasticsearch",
	})

	if err := r.EventDispatcher.Dispatch(ctx, dlqEvent); err != nil {
		return fmt.Errorf(FailedToPublishDLQErr, err)
	}

	return nil
}
