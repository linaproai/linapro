// This file implements Redis-backed short-lived key-value operations, including
// conditional writes, compare-and-delete, counters, expiration, and TTL reads.

package redis

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	"lina-core/internal/service/coordination/internal/core"
	"lina-core/pkg/bizerr"
)

// Get returns a string value for key when it exists.
func (r *redisBackend) Get(ctx context.Context, key string) (string, bool, error) {
	normalizedKey, err := core.RequireLogicalKey(key)
	if err != nil {
		return "", false, err
	}
	value, err := r.client.Get(ctx, normalizedKey).Result()
	if errors.Is(err, redis.Nil) {
		return "", false, nil
	}
	if err != nil {
		r.health.recordFailure(err)
		return "", false, bizerr.WrapCode(err, core.CodeCoordinationKVOperationFailed)
	}
	r.health.recordSuccess()
	return value, true, nil
}

// Set stores a string value with optional TTL. ttl=0 means no expiration.
func (r *redisBackend) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	normalizedKey, _, err := core.NormalizeKVWrite(key, ttl)
	if err != nil {
		return err
	}
	if err = r.client.Set(ctx, normalizedKey, value, ttl).Err(); err != nil {
		r.health.recordFailure(err)
		return bizerr.WrapCode(err, core.CodeCoordinationKVOperationFailed)
	}
	r.health.recordSuccess()
	return nil
}

// SetNX stores a string value only when the key is absent.
func (r *redisBackend) SetNX(ctx context.Context, key string, value string, ttl time.Duration) (bool, error) {
	normalizedKey, _, err := core.NormalizeKVWrite(key, ttl)
	if err != nil {
		return false, err
	}
	ok, err := r.client.SetNX(ctx, normalizedKey, value, ttl).Result()
	if err != nil {
		r.health.recordFailure(err)
		return false, bizerr.WrapCode(err, core.CodeCoordinationKVOperationFailed)
	}
	r.health.recordSuccess()
	return ok, nil
}

// Delete removes a key and treats missing keys as success.
func (r *redisBackend) Delete(ctx context.Context, key string) error {
	normalizedKey, err := core.RequireLogicalKey(key)
	if err != nil {
		return err
	}
	if err = r.client.Del(ctx, normalizedKey).Err(); err != nil {
		r.health.recordFailure(err)
		return bizerr.WrapCode(err, core.CodeCoordinationKVOperationFailed)
	}
	r.health.recordSuccess()
	return nil
}

// CompareAndDelete removes a key only when its current value matches expected.
func (r *redisBackend) CompareAndDelete(ctx context.Context, key string, expected string) (bool, error) {
	normalizedKey, err := core.RequireLogicalKey(key)
	if err != nil {
		return false, err
	}
	result, err := r.client.Eval(ctx, redisReleaseScript, []string{normalizedKey}, expected).Int64()
	if err != nil {
		r.health.recordFailure(err)
		return false, bizerr.WrapCode(err, core.CodeCoordinationKVOperationFailed)
	}
	r.health.recordSuccess()
	return result > 0, nil
}

// IncrBy increments an integer key by delta and returns the new value.
func (r *redisBackend) IncrBy(ctx context.Context, key string, delta int64, ttl time.Duration) (int64, error) {
	normalizedKey, _, err := core.NormalizeKVWrite(key, ttl)
	if err != nil {
		return 0, err
	}
	value, err := r.client.IncrBy(ctx, normalizedKey, delta).Result()
	if err != nil {
		r.health.recordFailure(err)
		return 0, bizerr.WrapCode(err, core.CodeCoordinationKVOperationFailed)
	}
	if ttl > 0 {
		if err = r.client.Expire(ctx, normalizedKey, ttl).Err(); err != nil {
			r.health.recordFailure(err)
			return 0, bizerr.WrapCode(err, core.CodeCoordinationKVOperationFailed)
		}
	}
	r.health.recordSuccess()
	return value, nil
}

// Expire updates a key expiration. ttl=0 clears expiration.
func (r *redisBackend) Expire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	normalizedKey, _, err := core.NormalizeKVWrite(key, ttl)
	if err != nil {
		return false, err
	}
	var updated bool
	if ttl == 0 {
		updated, err = r.client.Persist(ctx, normalizedKey).Result()
		if !updated && err == nil {
			exists, existsErr := r.client.Exists(ctx, normalizedKey).Result()
			if existsErr != nil {
				r.health.recordFailure(existsErr)
				return false, bizerr.WrapCode(existsErr, core.CodeCoordinationKVOperationFailed)
			}
			updated = exists > 0
		}
	} else {
		updated, err = r.client.Expire(ctx, normalizedKey, ttl).Result()
	}
	if err != nil {
		r.health.recordFailure(err)
		return false, bizerr.WrapCode(err, core.CodeCoordinationKVOperationFailed)
	}
	r.health.recordSuccess()
	return updated, nil
}

// TTL returns the remaining key lifetime. A negative duration means no TTL or missing key.
func (r *redisBackend) TTL(ctx context.Context, key string) (time.Duration, error) {
	normalizedKey, err := core.RequireLogicalKey(key)
	if err != nil {
		return 0, err
	}
	ttl, err := r.client.TTL(ctx, normalizedKey).Result()
	if err != nil {
		r.health.recordFailure(err)
		return 0, bizerr.WrapCode(err, core.CodeCoordinationKVOperationFailed)
	}
	r.health.recordSuccess()
	return ttl, nil
}
