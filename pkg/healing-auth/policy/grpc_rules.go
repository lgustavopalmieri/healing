package policy

import (
	"fmt"

	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

type GRPCRule struct {
	Policy             AccessPolicy
	OwnershipInPayload bool
}

func (rp *RoutePolicy) GRPCPublic(fullMethod string) {
	rp.registerGRPC(fullMethod, GRPCRule{Policy: PublicAccess()})
}

func (rp *RoutePolicy) GRPCAuthenticated(fullMethod string, roles ...role.Role) {
	rp.registerGRPC(fullMethod, GRPCRule{Policy: AuthenticatedAccess(roles...)})
}

func (rp *RoutePolicy) GRPCOwnerInPayload(fullMethod string, r role.Role) {
	rp.registerGRPC(fullMethod, GRPCRule{
		Policy:             OwnedAccess(r),
		OwnershipInPayload: true,
	})
}

func (rp *RoutePolicy) LookupGRPC(fullMethod string) (GRPCRule, bool) {
	rule, ok := rp.grpcRules[fullMethod]
	return rule, ok
}

func (rp *RoutePolicy) registerGRPC(fullMethod string, rule GRPCRule) {
	if fullMethod == "" {
		panic("route_policy: invalid gRPC registration with empty fullMethod")
	}
	if _, exists := rp.grpcRules[fullMethod]; exists {
		panic(fmt.Sprintf("route_policy: duplicate gRPC registration for %q", fullMethod))
	}
	rp.grpcRules[fullMethod] = rule
}
