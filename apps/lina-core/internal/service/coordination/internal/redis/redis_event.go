// This file implements Redis pub/sub event publishing and subscription
// handling for cross-node coordination notifications.

package redis

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"

	"lina-core/internal/service/coordination/internal/core"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
)

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
