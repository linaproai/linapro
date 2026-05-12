// This file implements the Redis-backed coordination provider used when
// cluster.enabled=true and cluster.coordination=redis.

package redis

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"lina-core/internal/service/coordination/internal/core"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
)

// Options contains normalized Redis connection settings for coordination.
type Options struct {
	Address        string           // Address is the host:port endpoint for Redis.
	DB             int              // DB selects the Redis logical database.
	Password       string           // Password authenticates to Redis when configured.
	ConnectTimeout time.Duration    // ConnectTimeout bounds Redis connection establishment.
	ReadTimeout    time.Duration    // ReadTimeout bounds Redis read operations.
	WriteTimeout   time.Duration    // WriteTimeout bounds Redis write operations.
	KeyBuilder     *core.KeyBuilder // KeyBuilder scopes all Redis keys and channels.
}

// redisBackend implements all coordination stores through one Redis client.
type redisBackend struct {
	client  *redis.Client
	keys    *core.KeyBuilder
	health  *redisHealth
	closeMu sync.Mutex
	closed  bool
}

// redisHealth stores observable Redis coordination health state.
type redisHealth struct {
	mu                  sync.RWMutex
	lastSuccessAt       time.Time
	lastError           string
	subscriberRunning   bool
	lastEventReceivedAt time.Time
}

// redisSubscription represents one active Redis pub/sub consumer.
type redisSubscription struct {
	pubsub *redis.PubSub
	cancel context.CancelFunc
	done   chan struct{}
	once   sync.Once
}

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

// New creates a Redis coordination service and verifies connectivity.
func New(ctx context.Context, options Options) (core.Service, error) {
	keys := options.KeyBuilder
	if keys == nil {
		keys = core.DefaultKeyBuilder()
	}
	if options.ConnectTimeout <= 0 {
		options.ConnectTimeout = 3 * time.Second
	}
	if options.ReadTimeout <= 0 {
		options.ReadTimeout = 2 * time.Second
	}
	if options.WriteTimeout <= 0 {
		options.WriteTimeout = 2 * time.Second
	}
	client := redis.NewClient(&redis.Options{
		Addr:         options.Address,
		DB:           options.DB,
		Password:     options.Password,
		DialTimeout:  options.ConnectTimeout,
		ReadTimeout:  options.ReadTimeout,
		WriteTimeout: options.WriteTimeout,
	})
	backend := &redisBackend{
		client: client,
		keys:   keys,
		health: &redisHealth{},
	}
	if err := backend.Ping(ctx); err != nil {
		closeErr := client.Close()
		if closeErr != nil {
			logger.Warningf(ctx, "[coordination] close redis client after ping failure: %v", closeErr)
		}
		return nil, err
	}
	return core.NewService(
		core.BackendRedis,
		keys,
		backend,
		backend,
		backend,
		backend,
		backend,
		backend.Close,
	), nil
}

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

// Publish publishes one event to peer nodes.
func (r *redisBackend) Publish(ctx context.Context, event core.Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	if err = r.client.Publish(ctx, r.keys.EventChannel(), payload).Err(); err != nil {
		r.health.recordFailure(err)
		return bizerr.WrapCode(err, core.CodeCoordinationEventPublishFailed)
	}
	r.health.recordSuccess()
	return nil
}

// Subscribe starts consuming events until the subscription is closed.
func (r *redisBackend) Subscribe(ctx context.Context, handler core.EventHandler) (core.Subscription, error) {
	if handler == nil {
		return nil, bizerr.NewCode(core.CodeCoordinationKeyInvalid, bizerr.P("field", "handler"))
	}
	subscribeCtx, cancel := context.WithCancel(ctx)
	pubsub := r.client.Subscribe(subscribeCtx, r.keys.EventChannel())
	if _, err := pubsub.Receive(subscribeCtx); err != nil {
		cancel()
		closeErr := pubsub.Close()
		if closeErr != nil {
			logger.Warningf(ctx, "[coordination] close failed subscription: %v", closeErr)
		}
		r.health.recordFailure(err)
		return nil, bizerr.WrapCode(err, core.CodeCoordinationEventSubscribeFailed)
	}
	subscription := &redisSubscription{
		pubsub: pubsub,
		cancel: cancel,
		done:   make(chan struct{}),
	}
	r.health.setSubscriberRunning(true)
	go r.consumeEvents(subscribeCtx, pubsub, handler, subscription.done)
	return subscription, nil
}

// Ping verifies backend connectivity.
func (r *redisBackend) Ping(ctx context.Context) error {
	if err := r.client.Ping(ctx).Err(); err != nil {
		r.health.recordFailure(err)
		return bizerr.WrapCode(err, core.CodeCoordinationRedisUnavailable)
	}
	r.health.recordSuccess()
	return nil
}

// Snapshot returns the latest health state.
func (r *redisBackend) Snapshot(ctx context.Context) core.HealthSnapshot {
	if err := ctx.Err(); err != nil {
		r.health.recordFailure(err)
	}
	return r.health.snapshot()
}

// Close releases Redis resources.
func (r *redisBackend) Close(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.closeMu.Lock()
	defer r.closeMu.Unlock()

	if r.closed {
		return nil
	}
	r.closed = true
	if err := r.client.Close(); err != nil {
		return bizerr.WrapCode(err, core.CodeCoordinationRedisUnavailable)
	}
	return nil
}

// consumeEvents decodes Redis pub/sub messages until subscription shutdown.
func (r *redisBackend) consumeEvents(
	ctx context.Context,
	pubsub *redis.PubSub,
	handler core.EventHandler,
	done chan<- struct{},
) {
	defer close(done)
	defer r.health.setSubscriberRunning(false)

	channel := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case message, ok := <-channel:
			if !ok {
				return
			}
			var event core.Event
			if err := json.Unmarshal([]byte(message.Payload), &event); err != nil {
				r.health.recordFailure(err)
				logger.Warningf(ctx, "[coordination] decode redis event failed: %v", err)
				continue
			}
			r.health.recordEventReceived()
			if err := handler(ctx, event); err != nil {
				r.health.recordFailure(err)
				logger.Warningf(ctx, "[coordination] handle redis event failed: %v", err)
			}
		}
	}
}

// Close stops the subscription and releases resources.
func (s *redisSubscription) Close(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	var closeErr error
	s.once.Do(func() {
		s.cancel()
		closeErr = s.pubsub.Close()
		<-s.done
	})
	if closeErr != nil {
		return bizerr.WrapCode(closeErr, core.CodeCoordinationEventSubscribeFailed)
	}
	return nil
}

// recordSuccess stores one successful Redis coordination operation.
func (h *redisHealth) recordSuccess() {
	h.mu.Lock()
	h.lastSuccessAt = time.Now()
	h.lastError = ""
	h.mu.Unlock()
}

// recordFailure stores the latest Redis coordination failure.
func (h *redisHealth) recordFailure(err error) {
	if err == nil {
		return
	}
	h.mu.Lock()
	h.lastError = err.Error()
	h.mu.Unlock()
}

// setSubscriberRunning records pub/sub lifecycle state.
func (h *redisHealth) setSubscriberRunning(running bool) {
	h.mu.Lock()
	h.subscriberRunning = running
	h.mu.Unlock()
}

// recordEventReceived stores the latest consumed event timestamp.
func (h *redisHealth) recordEventReceived() {
	h.mu.Lock()
	h.lastEventReceivedAt = time.Now()
	h.mu.Unlock()
}

// snapshot returns a detached Redis health snapshot.
func (h *redisHealth) snapshot() core.HealthSnapshot {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return core.HealthSnapshot{
		Backend:             core.BackendRedis,
		Healthy:             h.lastError == "",
		LastSuccessAt:       h.lastSuccessAt,
		LastError:           h.lastError,
		SubscriberRunning:   h.subscriberRunning,
		LastEventReceivedAt: h.lastEventReceivedAt,
	}
}
