package grpcservice

import (
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/search/application"
	esrepo "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/search/infra/elasticsearch"
)

type Dependencies struct {
	ESClient           *elasticsearch.Client
	ESIndexSpecialists string
	Logger             observability.Logger
}

func NewSpecialistSearchService(deps Dependencies) *SpecialistSearchGRPCService {
	repository := esrepo.NewRepository(deps.ESClient, deps.ESIndexSpecialists, deps.Logger)

	command := application.NewSearchSpecialistsCommand(repository, deps.Logger)

	return NewSpecialistSearchGRPCService(command)
}
