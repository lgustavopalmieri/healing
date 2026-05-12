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

func httpKey(method, path string) string {
	return strings.ToUpper(method) + " " + path
}
