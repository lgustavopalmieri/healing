package policy_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/policy"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

func TestRoutePolicy_HTTP(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(rp *policy.RoutePolicy)
		lookupMethod string
		lookupPath   string
		expectFound  bool
		validateRule func(t *testing.T, rule policy.HTTPRule)
	}{
		{
			name: "LookupHTTP retorna rule correta para key registrada via HTTPPublic",
			setup: func(rp *policy.RoutePolicy) {
				rp.HTTPPublic("GET", "/api/v1/specialists")
			},
			lookupMethod: "GET",
			lookupPath:   "/api/v1/specialists",
			expectFound:  true,
			validateRule: func(t *testing.T, rule policy.HTTPRule) {
				assert.True(t, rule.Policy.AllowPublic)
				assert.Empty(t, rule.OwnerIDParam)
			},
		},
		{
			name: "LookupHTTP retorna rule correta para key registrada via HTTPAuthenticated",
			setup: func(rp *policy.RoutePolicy) {
				rp.HTTPAuthenticated("POST", "/api/v1/appointments", role.Specialist, role.Patient)
			},
			lookupMethod: "POST",
			lookupPath:   "/api/v1/appointments",
			expectFound:  true,
			validateRule: func(t *testing.T, rule policy.HTTPRule) {
				assert.False(t, rule.Policy.AllowPublic)
				assert.False(t, rule.Policy.RequireOwnership)
				assert.Equal(t, []role.Role{role.Specialist, role.Patient}, rule.Policy.AllowedRoles)
			},
		},
		{
			name: "LookupHTTP retorna rule correta para key registrada via HTTPOwner",
			setup: func(rp *policy.RoutePolicy) {
				rp.HTTPOwner("PATCH", "/api/v1/specialists/:id", "id", role.Specialist)
			},
			lookupMethod: "PATCH",
			lookupPath:   "/api/v1/specialists/:id",
			expectFound:  true,
			validateRule: func(t *testing.T, rule policy.HTTPRule) {
				assert.True(t, rule.Policy.RequireOwnership)
				assert.Equal(t, "id", rule.OwnerIDParam)
				assert.Equal(t, []role.Role{role.Specialist}, rule.Policy.AllowedRoles)
			},
		},
		{
			name: "LookupHTTP retorna rule correta para key registrada via HTTPAdminReadOnly",
			setup: func(rp *policy.RoutePolicy) {
				rp.HTTPAdminReadOnly("GET", "/api/v1/admin/specialists")
			},
			lookupMethod: "GET",
			lookupPath:   "/api/v1/admin/specialists",
			expectFound:  true,
			validateRule: func(t *testing.T, rule policy.HTTPRule) {
				assert.False(t, rule.Policy.AllowPublic)
				assert.False(t, rule.Policy.RequireOwnership)
				assert.Equal(t, []role.Role{role.Admin}, rule.Policy.AllowedRoles)
			},
		},
		{
			name:         "LookupHTTP retorna ok=false para rota nao registrada",
			setup:        func(rp *policy.RoutePolicy) {},
			lookupMethod: "DELETE",
			lookupPath:   "/api/v1/nao-existe",
			expectFound:  false,
		},
		{
			name: "LookupHTTP e case-insensitive no method (get == GET)",
			setup: func(rp *policy.RoutePolicy) {
				rp.HTTPPublic("GET", "/api/v1/health")
			},
			lookupMethod: "get",
			lookupPath:   "/api/v1/health",
			expectFound:  true,
			validateRule: func(t *testing.T, rule policy.HTTPRule) {
				assert.True(t, rule.Policy.AllowPublic)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rp := policy.New()
			tt.setup(rp)

			rule, ok := rp.LookupHTTP(tt.lookupMethod, tt.lookupPath)
			assert.Equal(t, tt.expectFound, ok)
			if tt.expectFound {
				require.True(t, ok)
				if tt.validateRule != nil {
					tt.validateRule(t, rule)
				}
			}
		})
	}
}

func TestRoutePolicy_GRPC(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(rp *policy.RoutePolicy)
		lookupMethod string
		expectFound  bool
		validateRule func(t *testing.T, rule policy.GRPCRule)
	}{
		{
			name: "LookupGRPC retorna rule correta para fullMethod registrado via GRPCPublic",
			setup: func(rp *policy.RoutePolicy) {
				rp.GRPCPublic("/pb.SpecialistService/SearchSpecialists")
			},
			lookupMethod: "/pb.SpecialistService/SearchSpecialists",
			expectFound:  true,
			validateRule: func(t *testing.T, rule policy.GRPCRule) {
				assert.True(t, rule.Policy.AllowPublic)
				assert.False(t, rule.OwnershipInPayload)
			},
		},
		{
			name: "LookupGRPC retorna rule correta para fullMethod registrado via GRPCAuthenticated",
			setup: func(rp *policy.RoutePolicy) {
				rp.GRPCAuthenticated("/pb.AppointmentService/CreateAppointment", role.Patient)
			},
			lookupMethod: "/pb.AppointmentService/CreateAppointment",
			expectFound:  true,
			validateRule: func(t *testing.T, rule policy.GRPCRule) {
				assert.False(t, rule.Policy.AllowPublic)
				assert.False(t, rule.OwnershipInPayload)
				assert.Equal(t, []role.Role{role.Patient}, rule.Policy.AllowedRoles)
			},
		},
		{
			name: "LookupGRPC retorna rule correta para fullMethod registrado via GRPCOwnerInPayload",
			setup: func(rp *policy.RoutePolicy) {
				rp.GRPCOwnerInPayload("/pb.UpdateSpecialistService/UpdateSpecialist", role.Specialist)
			},
			lookupMethod: "/pb.UpdateSpecialistService/UpdateSpecialist",
			expectFound:  true,
			validateRule: func(t *testing.T, rule policy.GRPCRule) {
				assert.True(t, rule.Policy.RequireOwnership)
				assert.True(t, rule.OwnershipInPayload)
				assert.Equal(t, []role.Role{role.Specialist}, rule.Policy.AllowedRoles)
			},
		},
		{
			name:         "LookupGRPC retorna ok=false para fullMethod nao registrado",
			setup:        func(rp *policy.RoutePolicy) {},
			lookupMethod: "/pb.SomeService/SomeMethod",
			expectFound:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rp := policy.New()
			tt.setup(rp)

			rule, ok := rp.LookupGRPC(tt.lookupMethod)
			assert.Equal(t, tt.expectFound, ok)
			if tt.expectFound {
				require.True(t, ok)
				if tt.validateRule != nil {
					tt.validateRule(t, rule)
				}
			}
		})
	}
}

func TestRoutePolicy_DuplicateRegistration_HTTP(t *testing.T) {
	tests := []struct {
		name  string
		setup func(rp *policy.RoutePolicy)
	}{
		{
			name: "HTTPPublic duplicado panica",
			setup: func(rp *policy.RoutePolicy) {
				rp.HTTPPublic("POST", "/api/v1/login")
				rp.HTTPPublic("POST", "/api/v1/login")
			},
		},
		{
			name: "HTTPPublic seguido de HTTPAuthenticated no mesmo path panica",
			setup: func(rp *policy.RoutePolicy) {
				rp.HTTPPublic("POST", "/api/v1/login")
				rp.HTTPAuthenticated("POST", "/api/v1/login")
			},
		},
		{
			name: "HTTPOwner duplicado panica",
			setup: func(rp *policy.RoutePolicy) {
				rp.HTTPOwner("PATCH", "/api/v1/specialists/:id", "id", role.Specialist)
				rp.HTTPOwner("PATCH", "/api/v1/specialists/:id", "id", role.Specialist)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rp := policy.New()
			assert.Panics(t, func() {
				tt.setup(rp)
			})
		})
	}
}

func TestRoutePolicy_DuplicateRegistration_GRPC(t *testing.T) {
	t.Run("registro gRPC duplicado panica", func(t *testing.T) {
		rp := policy.New()
		rp.GRPCPublic("/pb.Svc/Method")

		assert.Panics(t, func() {
			rp.GRPCAuthenticated("/pb.Svc/Method", role.Specialist)
		})
	})
}

func TestRoutePolicy_InvalidRegistration(t *testing.T) {
	tests := []struct {
		name  string
		setup func(rp *policy.RoutePolicy)
	}{
		{
			name: "HTTP com method vazio panica",
			setup: func(rp *policy.RoutePolicy) {
				rp.HTTPPublic("", "/api/v1/login")
			},
		},
		{
			name: "HTTP com path vazio panica",
			setup: func(rp *policy.RoutePolicy) {
				rp.HTTPPublic("POST", "")
			},
		},
		{
			name: "gRPC com fullMethod vazio panica",
			setup: func(rp *policy.RoutePolicy) {
				rp.GRPCPublic("")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rp := policy.New()
			assert.Panics(t, func() {
				tt.setup(rp)
			})
		})
	}
}

func TestRoutePolicy_HTTP_CaseInsensitiveMethodVariants(t *testing.T) {
	rp := policy.New()
	rp.HTTPPublic("POST", "/api/v1/login")

	tests := []string{"post", "POST", "Post", "pOsT"}
	for _, m := range tests {
		t.Run("method variant: "+m, func(t *testing.T) {
			rule, ok := rp.LookupHTTP(m, "/api/v1/login")
			require.True(t, ok)
			assert.True(t, rule.Policy.AllowPublic)
		})
	}
}

func TestRoutePolicy_GRPC_FullMethodIsCaseSensitive(t *testing.T) {
	rp := policy.New()
	rp.GRPCPublic("/pb.SpecialistService/CreateSpecialist")

	_, ok := rp.LookupGRPC("/pb.specialistservice/createspecialist")

	assert.False(t, ok)
}
