// This file implements Redis coordination health tracking, ping verification,
// backend close handling, and detached health snapshots for observers.

package redis

import (
	"context"
	"time"

	"lina-core/internal/service/coordination/internal/core"
	"lina-core/pkg/bizerr"
)

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
