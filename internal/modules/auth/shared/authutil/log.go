package authutil

import (
	"context"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
)

func LogError(ctx context.Context, logger observability.Logger, message string, err error, subjectID string) {
	if logger == nil {
		return
	}
	logger.Error(ctx, message,
		observability.Field{Key: "subject_id", Value: subjectID},
		observability.Field{Key: "error", Value: err.Error()},
	)
}
