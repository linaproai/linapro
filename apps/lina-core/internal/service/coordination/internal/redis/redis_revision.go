// This file implements Redis-backed monotonic cache revision reads and bumps;
// callers rely on the returned values for scoped cache invalidation ordering.

package redis

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"

	"lina-core/internal/service/coordination/internal/core"
	"lina-core/pkg/bizerr"
)

// Bump increments and returns one revision.
func (r *redisBackend) Bump(ctx context.Context, key core.RevisionKey, reason string) (int64, error) {
	backendKey, err := r.keys.RevisionKey(key)
	if err != nil {
		return 0, err
	}
	revision, err := r.client.Incr(ctx, backendKey).Result()
	if err != nil {
		r.health.recordFailure(err)
		return 0, bizerr.WrapCode(err, core.CodeCoordinationRevisionUnavailable)
	}
	r.health.recordSuccess()
	return revision, nil
}

// Current returns the latest revision, initializing missing revisions as 0.
func (r *redisBackend) Current(ctx context.Context, key core.RevisionKey) (int64, error) {
	backendKey, err := r.keys.RevisionKey(key)
	if err != nil {
		return 0, err
	}
	value, err := r.client.Get(ctx, backendKey).Int64()
	if errors.Is(err, redis.Nil) {
		return 0, nil
	}
	if err != nil {
		r.health.recordFailure(err)
		return 0, bizerr.WrapCode(err, core.CodeCoordinationRevisionUnavailable)
	}
	r.health.recordSuccess()
	return value, nil
}
