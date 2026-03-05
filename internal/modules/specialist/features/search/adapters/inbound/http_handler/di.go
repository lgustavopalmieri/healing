package httphandler

import (
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/search/application"
	esrepo "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/search/adapters/outbound/elasticsearch"
)

type Dependencies struct {
	ESClient           *elasticsearch.Client
	ESIndexSpecialists string
	Logger             observability.Logger
}

func NewSpecialistSearchHandler(deps Dependencies) *SpecialistSearchHTTPHandler {
	repository := esrepo.NewRepository(deps.ESClient, deps.ESIndexSpecialists, deps.Logger)

	useCase := application.NewSearchSpecialistsUseCase(repository, deps.Logger)

	return NewSpecialistSearchHTTPHandler(useCase)
}
