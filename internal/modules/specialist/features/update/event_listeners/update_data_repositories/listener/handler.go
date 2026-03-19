package listener

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
)

func (h *UpdateDataRepositoriesHandler) Handle(ctx context.Context, evt event.Event) error {
	payload := UpdateDataRepositoriesEventPayload{}
	err := json.Unmarshal(evt.Payload.([]byte), &payload)
	if err != nil {
		return fmt.Errorf("%s: %w", ErrUnmarshalEventPayloadMessage, err)
	}

	specialist, err := h.sourceRepository.FindByID(ctx, payload.ID)
	if err != nil {
		return ErrSpecialistNotFound
	}

	hasFailure := false

	for _, repo := range h.dataRepositories {
		retryErr := event.WithRetry(ctx, h.retryConfig, func(ctx context.Context) error {
			return repo.Update(ctx, specialist)
		})

		if retryErr != nil {
			hasFailure = true

			if dlqErr := repo.PublishDLQ(ctx, specialist, retryErr); dlqErr != nil {
				continue
			}

			continue
		}
	}

	if hasFailure {
		return ErrUpdateDataRepositories
	}

	return nil
}
