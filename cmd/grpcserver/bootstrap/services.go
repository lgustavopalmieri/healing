package bootstrap

import (
	"database/sql"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
	grpchandler "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/infra/grpc_handler"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/infra/grpc_handler/pb"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/server"
)

type ServiceDependencies struct {
	DB             *sql.DB
	EventPublisher event.EventDispatcher
	Tracer         observability.Tracer
	Logger         observability.Logger
}

func RegisterServices(grpcServer *server.GRPCServer, deps ServiceDependencies) {
	specialistCreateHandler := grpchandler.NewSpecialistCreateService(grpchandler.Dependencies{
		DB:             deps.DB,
		EventPublisher: deps.EventPublisher,
		Tracer:         deps.Tracer,
		Logger:         deps.Logger,
	})

	pb.RegisterSpecialistServiceServer(grpcServer.GetServer(), specialistCreateHandler)
}
