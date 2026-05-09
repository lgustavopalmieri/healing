package policy

import (
	"fmt"
	"strings"

	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

type HTTPRule struct {
	Policy       AccessPolicy
	OwnerIDParam string
}

type GRPCRule struct {
	Policy             AccessPolicy
	OwnershipInPayload bool
}

type RoutePolicy struct {
	httpRules map[string]HTTPRule
	grpcRules map[string]GRPCRule
}

func New() *RoutePolicy {
	return &RoutePolicy{
		httpRules: make(map[string]HTTPRule),
		grpcRules: make(map[string]GRPCRule),
	}
}

func (rp *RoutePolicy) HTTPPublic(method, path string) {
	rp.registerHTTP(method, path, HTTPRule{Policy: PublicAccess()})
}

func (rp *RoutePolicy) HTTPAuthenticated(method, path string, roles ...role.Role) {
	rp.registerHTTP(method, path, HTTPRule{Policy: AuthenticatedAccess(roles...)})
}

func (rp *RoutePolicy) HTTPOwner(method, path, ownerParam string, r role.Role) {
	rp.registerHTTP(method, path, HTTPRule{
		Policy:       OwnedAccess(r),
		OwnerIDParam: ownerParam,
	})
}

func (rp *RoutePolicy) HTTPAdminReadOnly(method, path string) {
	rp.registerHTTP(method, path, HTTPRule{Policy: AdminReadOnly()})
}

func (rp *RoutePolicy) LookupHTTP(method, path string) (HTTPRule, bool) {
	rule, ok := rp.httpRules[httpKey(method, path)]
	return rule, ok
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

func (rp *RoutePolicy) registerHTTP(method, path string, rule HTTPRule) {
	if method == "" || path == "" {
		panic(fmt.Sprintf("route_policy: invalid HTTP registration method=%q path=%q", method, path))
	}
	key := httpKey(method, path)
	if _, exists := rp.httpRules[key]; exists {
		panic(fmt.Sprintf("route_policy: duplicate HTTP registration for %q", key))
	}
	rp.httpRules[key] = rule
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

func httpKey(method, path string) string {
	return strings.ToUpper(method) + " " + path
}
