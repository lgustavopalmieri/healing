package listener

import (
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
)

type UpdateDataRepositoriesHandler struct {
	sourceRepository SourceRepository
	dataRepositories []DataRepository
	retryConfig      event.RetryConfig
}

func NewUpdateDataRepositoriesHandler(
	sourceRepository SourceRepository,
	dataRepositories []DataRepository,
) *UpdateDataRepositoriesHandler {
	return &UpdateDataRepositoriesHandler{
		sourceRepository: sourceRepository,
		dataRepositories: dataRepositories,
		retryConfig:      event.DefaultRetryConfig(),
	}
}

func (h *UpdateDataRepositoriesHandler) WithRetryConfig(cfg event.RetryConfig) *UpdateDataRepositoriesHandler {
	h.retryConfig = cfg
	return h
}
