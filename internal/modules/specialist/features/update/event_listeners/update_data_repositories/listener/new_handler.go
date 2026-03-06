package listener

import (
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
)

type UpdateDataRepositoriesHandler struct {
	sourceRepository SourceRepository
	dataRepositories []DataRepository
	tracer           observability.Tracer
	logger           observability.Logger
	retryConfig      event.RetryConfig
}

func NewUpdateDataRepositoriesHandler(
	sourceRepository SourceRepository,
	dataRepositories []DataRepository,
	tracer observability.Tracer,
	logger observability.Logger,
) *UpdateDataRepositoriesHandler {
	return &UpdateDataRepositoriesHandler{
		sourceRepository: sourceRepository,
		dataRepositories: dataRepositories,
		tracer:           tracer,
		logger:           logger,
		retryConfig:      event.DefaultRetryConfig(),
	}
}

func (h *UpdateDataRepositoriesHandler) WithRetryConfig(cfg event.RetryConfig) *UpdateDataRepositoriesHandler {
	h.retryConfig = cfg
	return h
}
