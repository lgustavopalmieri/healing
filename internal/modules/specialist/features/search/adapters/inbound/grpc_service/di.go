package grpcservice

import (
	"github.com/elastic/go-elasticsearch/v8"
	esrepo "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/search/adapters/outbound/elasticsearch"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/search/application"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/elasticsearch/indexes"
)

type Dependencies struct {
	ESClient *elasticsearch.Client
}

func NewSpecialistSearchService(deps Dependencies) *SpecialistSearchGRPCService {
	repository := esrepo.NewRepository(deps.ESClient, indexes.SpecialistsIndex)

	useCase := application.NewSearchSpecialistsUseCase(repository)

	return NewSpecialistSearchGRPCService(useCase)
}
