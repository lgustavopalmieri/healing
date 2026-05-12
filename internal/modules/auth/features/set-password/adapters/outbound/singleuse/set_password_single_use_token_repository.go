package singleuse

import (
	"context"

	singleusetoken "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/shared/providers/single_use_token"
)

type SetPasswordSingleUseTokenRepository struct {
	store *singleusetoken.SingleUseTokenCacheRepository
}

func NewSetPasswordSingleUseTokenRepository(store *singleusetoken.SingleUseTokenCacheRepository) *SetPasswordSingleUseTokenRepository {
	return &SetPasswordSingleUseTokenRepository{store: store}
}

func (r *SetPasswordSingleUseTokenRepository) Consume(ctx context.Context, jti string) (bool, error) {
	return r.store.Consume(ctx, singleusetoken.PurposeSetPassword, jti)
}
