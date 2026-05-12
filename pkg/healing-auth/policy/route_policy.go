package policy

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
