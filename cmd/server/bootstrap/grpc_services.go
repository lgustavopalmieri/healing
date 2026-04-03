package bootstrap

import (
	"database/sql"
	"log"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
	creategrpc "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/adapters/inbound/grpc_service"
	createpb "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/adapters/inbound/grpc_service/pb"
	searchgrpc "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/search/adapters/inbound/grpc_service"
	searchpb "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/search/adapters/inbound/grpc_service/pb"
	updategrpc "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/update/adapters/inbound/grpc_service"
	updatepb "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/update/adapters/inbound/grpc_service/pb"
	platformOS "github.com/lgustavopalmieri/healing-specialist/internal/platform/opensearch"
	opensearchindexes "github.com/lgustavopalmieri/healing-specialist/internal/platform/opensearch/indexes"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/server"
)

type ServiceDependencies struct {
	DB             *sql.DB
	OSFactory      *platformOS.Factory
	EventPublisher event.EventDispatcher
	Logger         observability.Logger
}

func RegisterServices(grpcServer *server.GRPCServer, deps ServiceDependencies) {
	log.Println("🔧 Registering gRPC services...")
	specialistCreateService := creategrpc.NewSpecialistCreateService(creategrpc.Dependencies{
		DB:             deps.DB,
		EventPublisher: deps.EventPublisher,
		Logger:         deps.Logger,
	})
	createpb.RegisterSpecialistServiceServer(grpcServer.GetServer(), specialistCreateService)

	specialistSearchService := searchgrpc.NewSpecialistSearchService(searchgrpc.Dependencies{
		OSClient:  deps.OSFactory.Client,
		IndexName: deps.OSFactory.IndexName(opensearchindexes.SpecialistsIndex),
	})
	searchpb.RegisterSearchSpecialistServiceServer(grpcServer.GetServer(), specialistSearchService)

	specialistUpdateService := updategrpc.NewSpecialistUpdateService(updategrpc.Dependencies{
		DB:             deps.DB,
		EventPublisher: deps.EventPublisher,
	})
	updatepb.RegisterUpdateSpecialistServiceServer(grpcServer.GetServer(), specialistUpdateService)

	log.Println("✅ gRPC services registered successfully")
}
