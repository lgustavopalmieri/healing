package bootstrap

import (
	"database/sql"
	"log"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
	creategrpc "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/infra/grpc_service"
	createpb "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/infra/grpc_service/pb"
	searchgrpc "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/search/infra/grpc_service"
	searchpb "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/search/infra/grpc_service/pb"
	updategrpc "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/update/infra/grpc_service"
	updatepb "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/update/infra/grpc_service/pb"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/server"
)

type ServiceDependencies struct {
	DB                 *sql.DB
	ESClient           *elasticsearch.Client
	EventPublisher     event.EventDispatcher
	Tracer             observability.Tracer
	Logger             observability.Logger
	ESIndexSpecialists string
}

func RegisterServices(grpcServer *server.GRPCServer, deps ServiceDependencies) {
	log.Println("🔧 Registering gRPC services...")
	specialistCreateService := creategrpc.NewSpecialistCreateService(creategrpc.Dependencies{
		DB:             deps.DB,
		EventPublisher: deps.EventPublisher,
		Tracer:         deps.Tracer,
		Logger:         deps.Logger,
	})
	createpb.RegisterSpecialistServiceServer(grpcServer.GetServer(), specialistCreateService)

	specialistSearchService := searchgrpc.NewSpecialistSearchService(searchgrpc.Dependencies{
		ESClient:           deps.ESClient,
		ESIndexSpecialists: deps.ESIndexSpecialists,
		Logger:             deps.Logger,
	})
	searchpb.RegisterSearchSpecialistServiceServer(grpcServer.GetServer(), specialistSearchService)

	specialistUpdateService := updategrpc.NewSpecialistUpdateService(updategrpc.Dependencies{
		DB:             deps.DB,
		EventPublisher: deps.EventPublisher,
		Tracer:         deps.Tracer,
		Logger:         deps.Logger,
	})
	updatepb.RegisterUpdateSpecialistServiceServer(grpcServer.GetServer(), specialistUpdateService)

	log.Println("✅ Services registered successfully")
}
