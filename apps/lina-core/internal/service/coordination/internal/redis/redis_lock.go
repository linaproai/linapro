// This file implements Redis-backed distributed lock operations, including
// atomic owner checks, lease renewal, release, and fencing token allocation.

package redis

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	"lina-core/internal/service/coordination/internal/core"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
)

// Lua scripts keep Redis lock operations atomic.
const (
	redisRenewScript = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("PEXPIRE", KEYS[1], ARGV[2])
end
return 0
`
	redisReleaseScript = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
end
return 0
`
)

// Acquire obtains a lock when it is absent or expired.
func (r *redisBackend) Acquire(
	ctx context.Context,
	name string,
	owner string,
	reason string,
	ttl time.Duration,
) (*core.LockHandle, bool, error) {
	if ttl <= 0 {
		return nil, false, core.BizerrTTLInvalid()
	}
	lockKey, err := r.keys.LockKey(name)
	if err != nil {
		return nil, false, err
	}
	fenceKey, err := r.keys.LockFenceKey(name)
	if err != nil {
		return nil, false, err
	}
	token, err := core.NewOwnerToken(owner)
	if err != nil {
		return nil, false, err
	}
	ok, err := r.client.SetNX(ctx, lockKey, token, ttl).Result()
	if err != nil {
		r.health.recordFailure(err)
		return nil, false, bizerr.WrapCode(err, core.CodeCoordinationRedisUnavailable)
	}
	if !ok {
		return nil, false, nil
	}
	fencingToken, err := r.client.Incr(ctx, fenceKey).Result()
	if err != nil {
		r.health.recordFailure(err)
		if releaseErr := r.Release(ctx, &core.LockHandle{Name: name, Owner: owner, Token: token}); releaseErr != nil {
			logger.Warningf(ctx, "[coordination] release lock after fencing failure name=%s err=%v", name, releaseErr)
		}
		return nil, false, bizerr.WrapCode(err, core.CodeCoordinationRedisUnavailable)
	}
	r.health.recordSuccess()
	return &core.LockHandle{
		Name:         name,
		Owner:        owner,
		Token:        token,
		Reason:       reason,
		Lease:        ttl,
		FencingToken: fencingToken,
		AcquiredAt:   time.Now(),
	}, true, nil
}

// Renew extends a lock only when the caller still owns it.
func (r *redisBackend) Renew(ctx context.Context, handle *core.LockHandle, ttl time.Duration) error {
	if ttl <= 0 {
		return core.BizerrTTLInvalid()
	}
	if handle == nil {
		return bizerr.NewCode(core.CodeCoordinationLockNotHeld)
	}
	lockKey, err := r.keys.LockKey(core.HandleName(handle))
	if err != nil {
		return err
	}
	result, err := r.client.Eval(ctx, redisRenewScript, []string{lockKey}, handle.Token, ttl.Milliseconds()).Int64()
	if err != nil {
		r.health.recordFailure(err)
		return bizerr.WrapCode(err, core.CodeCoordinationRedisUnavailable)
	}
	if result <= 0 {
		return bizerr.NewCode(core.CodeCoordinationLockNotHeld)
	}
	r.health.recordSuccess()
	return nil
}

// Release releases a lock only when the caller still owns it.
func (r *redisBackend) Release(ctx context.Context, handle *core.LockHandle) error {
	if handle == nil {
		return bizerr.NewCode(core.CodeCoordinationLockNotHeld)
	}
	lockKey, err := r.keys.LockKey(core.HandleName(handle))
	if err != nil {
		return err
	}
	result, err := r.client.Eval(ctx, redisReleaseScript, []string{lockKey}, handle.Token).Int64()
	if err != nil {
		r.health.recordFailure(err)
		return bizerr.WrapCode(err, core.CodeCoordinationRedisUnavailable)
	}
	if result <= 0 {
		return bizerr.NewCode(core.CodeCoordinationLockNotHeld)
	}
	r.health.recordSuccess()
	return nil
}

// IsHeld reports whether the handle still owns the lock.
func (r *redisBackend) IsHeld(ctx context.Context, handle *core.LockHandle) (bool, error) {
	if handle == nil {
		return false, nil
	}
	lockKey, err := r.keys.LockKey(core.HandleName(handle))
	if err != nil {
		return false, err
	}
	value, err := r.client.Get(ctx, lockKey).Result()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		r.health.recordFailure(err)
		return false, bizerr.WrapCode(err, core.CodeCoordinationRedisUnavailable)
	}
	r.health.recordSuccess()
	return value == handle.Token, nil
}
