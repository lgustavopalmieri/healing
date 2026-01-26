package bootstrap

import (
	"database/sql"
	"log"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
	grpcservice "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/infra/grpc_service"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/infra/grpc_service/pb"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/server"
)

type ServiceDependencies struct {
	DB             *sql.DB
	EventPublisher event.EventDispatcher
	Tracer         observability.Tracer
	Logger         observability.Logger
}

func RegisterServices(grpcServer *server.GRPCServer, deps ServiceDependencies) {
	log.Println("🔧 Registering gRPC services...")
	specialistCreateService := grpcservice.NewSpecialistCreateService(grpcservice.Dependencies{
		DB:             deps.DB,
		EventPublisher: deps.EventPublisher,
		Tracer:         deps.Tracer,
		Logger:         deps.Logger,
	})

	pb.RegisterSpecialistServiceServer(grpcServer.GetServer(), specialistCreateService)
	log.Println("✅ Services registered successfully")
}
