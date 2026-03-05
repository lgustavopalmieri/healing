package bootstrap

import (
	"log"

	createhttp "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/adapters/inbound/http_handler"
	searchhttp "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/search/adapters/inbound/http_handler"
	updatehttp "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/update/adapters/inbound/http_handler"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/server"
)

func RegisterHTTPServices(httpServer *server.HTTPServer, deps ServiceDependencies) {
	log.Println("🔧 Registering HTTP services...")

	api := httpServer.Engine.Group("/api/v1")

	createHandler := createhttp.NewSpecialistCreateHandler(createhttp.Dependencies{
		DB:             deps.DB,
		EventPublisher: deps.EventPublisher,
		Tracer:         deps.Tracer,
		Logger:         deps.Logger,
	})
	createHandler.RegisterRoutes(api)

	searchHandler := searchhttp.NewSpecialistSearchHandler(searchhttp.Dependencies{
		ESClient:           deps.ESClient,
		ESIndexSpecialists: deps.ESIndexSpecialists,
		Logger:             deps.Logger,
	})
	searchHandler.RegisterRoutes(api)

	updateHandler := updatehttp.NewSpecialistUpdateHandler(updatehttp.Dependencies{
		DB:             deps.DB,
		EventPublisher: deps.EventPublisher,
		Tracer:         deps.Tracer,
		Logger:         deps.Logger,
	})
	updateHandler.RegisterRoutes(api)

	log.Println("✅ HTTP services registered successfully")
}
