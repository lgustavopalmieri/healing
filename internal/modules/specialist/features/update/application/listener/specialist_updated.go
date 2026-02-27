package listener

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
)

type SpecialistUpdatedListener struct {
	repository  SpecialistFindByIDRepositoryInterface
	projections []SpecialistReadProjectionInterface
	tracer      observability.Tracer
	logger      observability.Logger
}

func NewSpecialistUpdatedListener(
	repository SpecialistFindByIDRepositoryInterface,
	projections []SpecialistReadProjectionInterface,
	tracer observability.Tracer,
	logger observability.Logger,
) *SpecialistUpdatedListener {
	return &SpecialistUpdatedListener{
		repository:  repository,
		projections: projections,
		tracer:      tracer,
		logger:      logger,
	}
}

func (l *SpecialistUpdatedListener) Handle(ctx context.Context, evt event.Event) error {
	ctx, span := l.tracer.Start(ctx, SpecialistUpdatedListenerSpanName)
	defer span.End()

	l.logger.Info(ctx, StartingSpecialistUpdatedMessage)

	var payload SpecialistUpdatedPayload
	err := json.Unmarshal(evt.Payload.([]byte), &payload)
	if err != nil {
		span.RecordError(err)
		l.logger.Error(ctx, ErrUnmarshalSpecialistUpdatedEventMessage,
			observability.Field{Key: "error", Value: err.Error()})
		return fmt.Errorf("%w: %w", ErrUnmarshalPayload, err)
	}

	specialist, err := l.repository.FindByID(ctx, payload.ID)
	if err != nil {
		span.RecordError(err)
		l.logger.Error(ctx, ErrFindSpecialistByIDMessage,
			observability.Field{Key: "id", Value: payload.ID},
			observability.Field{Key: "error", Value: err.Error()})
		return fmt.Errorf("%w: %w", ErrFindSpecialistByID, err)
	}

	for _, projection := range l.projections {
		if err := projection.Update(ctx, specialist); err != nil {
			span.RecordError(err)
			l.logger.Error(ctx, ErrUpdateProjectionMessage,
				observability.Field{Key: "id", Value: specialist.ID},
				observability.Field{Key: "error", Value: err.Error()})
			return fmt.Errorf("%w: %w", ErrUpdateProjection, err)
		}
	}

	l.logger.Info(ctx, SpecialistUpdatedProcessedSuccessMessage,
		observability.Field{Key: "id", Value: specialist.ID},
		observability.Field{Key: "status", Value: string(specialist.Status)})

	return nil
}
