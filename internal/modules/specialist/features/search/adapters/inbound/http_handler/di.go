package httphandler

import (
	"github.com/elastic/go-elasticsearch/v8"
	esrepo "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/search/adapters/outbound/elasticsearch"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/search/application"
)

type Dependencies struct {
	ESClient           *elasticsearch.Client
	ESIndexSpecialists string
}

func NewSpecialistSearchHandler(deps Dependencies) *SpecialistSearchHTTPHandler {
	repository := esrepo.NewRepository(deps.ESClient, deps.ESIndexSpecialists)

	useCase := application.NewSearchSpecialistsUseCase(repository)

	return NewSpecialistSearchHTTPHandler(useCase)
}
