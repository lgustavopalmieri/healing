package httphandler

import (
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"

	osrepo "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/search/adapters/outbound/opensearch"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/search/application"
)

type Dependencies struct {
	OSClient  *opensearchapi.Client
	IndexName string
}

func NewSpecialistSearchHandler(deps Dependencies) *SpecialistSearchHTTPHandler {
	repository := osrepo.NewRepository(deps.OSClient, deps.IndexName)

	useCase := application.NewSearchSpecialistsUseCase(repository)

	return NewSpecialistSearchHTTPHandler(useCase)
}
