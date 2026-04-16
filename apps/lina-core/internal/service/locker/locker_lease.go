package locker

import (
	"context"
	"lina-core/pkg/logger"
	"time"
)

// LeaseManager manages automatic lease renewal for a lock.
type LeaseManager struct {
	instance    *Instance
	renewIntvl  time.Duration
	stopChan    chan struct{}
	stoppedChan chan struct{}
}

// NewLeaseManager creates a new lease manager.
func NewLeaseManager(instance *Instance, renewInterval time.Duration) *LeaseManager {
	return &LeaseManager{
		instance:    instance,
		renewIntvl:  renewInterval,
		stopChan:    make(chan struct{}),
		stoppedChan: make(chan struct{}),
	}
}

// Start begins the automatic lease renewal process.
// It runs in a goroutine and renews the lock at the specified interval.
func (lm *LeaseManager) Start(ctx context.Context) {
	go func() {
		defer close(lm.stoppedChan)

		ticker := time.NewTicker(lm.renewIntvl)
		defer ticker.Stop()

		for {
			select {
			case <-lm.stopChan:
				logger.Infof(ctx, "[locker] lease renewal stopped")
				return
			case <-ticker.C:
				if err := lm.instance.Renew(ctx); err != nil {
					logger.Warningf(ctx, "[locker] lease renewal failed: %v", err)
					return
				}
				logger.Debugf(ctx, "[locker] lease renewed successfully")
			}
		}
	}()
}

// Stop stops the lease renewal process.
func (lm *LeaseManager) Stop() {
	close(lm.stopChan)
	<-lm.stoppedChan // Wait for the goroutine to finish
}

// StoppedChan returns a channel that is closed when the lease manager stops.
func (lm *LeaseManager) StoppedChan() <-chan struct{} {
	return lm.stoppedChan
}
