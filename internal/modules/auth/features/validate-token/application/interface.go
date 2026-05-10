package application

import "context"

//go:generate mockgen -source=interface.go -destination=mocks/repository_mock.go -package=mocks
type BlacklistRepository interface {
	IsBlacklisted(ctx context.Context, jti string) (bool, error)
}
