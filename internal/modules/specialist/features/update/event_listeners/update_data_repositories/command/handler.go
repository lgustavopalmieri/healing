package command

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
)

func (h *UpdateDataRepositoriesHandler) Handle(ctx context.Context, evt event.Event) error {
	ctx, span := h.tracer.Start(ctx, UpdateDataRepositoriesSpanName)
	defer span.End()

	payload := UpdateDataRepositoriesEventPayload{}
	err := json.Unmarshal(evt.Payload.([]byte), &payload)
	if err != nil {
		return fmt.Errorf("%s: %w", ErrUnmarshalEventPayloadMessage, err)
	}

	specialist, err := h.sourceRepository.FindByID(ctx, payload.ID)
	if err != nil {
		span.RecordError(err)
		h.logger.Error(ctx, ErrSpecialistNotFoundMessage,
			observability.Field{Key: "id", Value: payload.ID},
			observability.Field{Key: "error", Value: err.Error()})
		return ErrSpecialistNotFound
	}

	h.logger.Info(ctx, StartingDataRepositoriesUpdateMessage,
		observability.Field{Key: "id", Value: specialist.ID})

	hasFailure := false

	for _, repo := range h.dataRepositories {
		retryErr := event.WithRetry(ctx, h.retryConfig, func(ctx context.Context) error {
			return repo.Update(ctx, specialist)
		})

		if retryErr != nil {
			hasFailure = true
			span.RecordError(retryErr)
			h.logger.Error(ctx, RepositoryUpdateFailedMessage,
				observability.Field{Key: "id", Value: specialist.ID},
				observability.Field{Key: "error", Value: retryErr.Error()})

			if dlqErr := repo.PublishDLQ(ctx, specialist, retryErr); dlqErr != nil {
				span.RecordError(dlqErr)
				h.logger.Error(ctx, DLQPublishFailedMessage,
					observability.Field{Key: "id", Value: specialist.ID},
					observability.Field{Key: "error", Value: dlqErr.Error()})
			}

			continue
		}

		h.logger.Info(ctx, RepositoryUpdateSucceededMessage,
			observability.Field{Key: "id", Value: specialist.ID})
	}

	if hasFailure {
		return ErrUpdateDataRepositories
	}

	h.logger.Info(ctx, DataRepositoriesUpdatedSuccessMessage,
		observability.Field{Key: "id", Value: specialist.ID})

	return nil
}
