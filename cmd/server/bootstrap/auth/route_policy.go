package auth

import (
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/policy"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

func BuildRoutePolicy() *policy.RoutePolicy {
	p := policy.New()

	// Health + Swagger
	p.HTTPPublic("GET", "/health")
	p.HTTPPublic("GET", "/swagger/*any")

	// Specialist — publicas
	p.HTTPPublic("POST", "/api/v1/specialists")
	p.HTTPPublic("POST", "/api/v1/specialists/search")

	// Specialist — ownership (specialist so edita o proprio perfil)
	p.HTTPOwner("PATCH", "/api/v1/specialists/:id", "id", role.Specialist)

	// Auth — publicas
	p.HTTPPublic("POST", "/api/v1/auth/specialist/login")
	p.HTTPPublic("POST", "/api/v1/auth/patient/login")
	p.HTTPPublic("POST", "/api/v1/auth/admin/login")
	p.HTTPPublic("POST", "/api/v1/auth/refresh")
	p.HTTPPublic("POST", "/api/v1/auth/set-password")
	p.HTTPPublic("POST", "/api/v1/auth/reset-password/request")
	p.HTTPPublic("POST", "/api/v1/auth/reset-password")

	// Auth — autenticadas (qualquer role logada)
	p.HTTPAuthenticated("POST", "/api/v1/auth/logout", role.Specialist, role.Patient, role.Admin)
	p.HTTPAuthenticated("POST", "/api/v1/auth/change-password", role.Specialist, role.Patient, role.Admin)
	p.HTTPAuthenticated("POST", "/api/v1/auth/revoke-all-sessions", role.Specialist, role.Patient, role.Admin)

	// gRPC — publicas
	p.GRPCPublic("/pb.SpecialistService/CreateSpecialist")
	p.GRPCPublic("/pb.SearchSpecialistService/SearchSpecialists")

	// gRPC — ownership (specialist so edita o proprio)
	p.GRPCOwnerInPayload("/pb.UpdateSpecialistService/UpdateSpecialist", role.Specialist)

	return p
}
