// This file implements lock lease renewal for Redis-backed leader election.

package cluster

import (
	"context"
	"sync"
	"time"

	"lina-core/internal/service/coordination"
	"lina-core/pkg/logger"
)

// electionLeaseManager renews one leader-election lock until renewal fails or
// the manager is explicitly stopped.
type electionLeaseManager struct {
	lockStore   coordination.LockStore
	handle      *coordination.LockHandle
	lease       time.Duration
	renewIntvl  time.Duration
	stopChan    chan struct{}
	stoppedChan chan struct{}
	stopOnce    sync.Once
}

// newElectionLeaseManager creates one leader lock renewal manager.
func newElectionLeaseManager(
	lockStore coordination.LockStore,
	handle *coordination.LockHandle,
	lease time.Duration,
	renewInterval time.Duration,
) *electionLeaseManager {
	return &electionLeaseManager{
		lockStore:   lockStore,
		handle:      handle,
		lease:       lease,
		renewIntvl:  renewInterval,
		stopChan:    make(chan struct{}),
		stoppedChan: make(chan struct{}),
	}
}

// Start begins automatic lease renewal.
func (m *electionLeaseManager) Start(ctx context.Context) {
	go func() {
		defer close(m.stoppedChan)

		ticker := time.NewTicker(m.renewIntvl)
		defer ticker.Stop()

		for {
			select {
			case <-m.stopChan:
				logger.Infof(ctx, "[cluster] leader lease renewal stopped")
				return
			case <-ticker.C:
				if err := m.lockStore.Renew(ctx, m.handle, m.lease); err != nil {
					logger.Warningf(ctx, "[cluster] leader lease renewal failed: %v", err)
					return
				}
				logger.Debugf(ctx, "[cluster] leader lease renewed")
			}
		}
	}()
}

// Stop stops automatic lease renewal and waits for the goroutine to exit.
func (m *electionLeaseManager) Stop() {
	m.stopOnce.Do(func() {
		close(m.stopChan)
	})
	<-m.stoppedChan
}

// StoppedChan returns a channel closed when renewal exits.
func (m *electionLeaseManager) StoppedChan() <-chan struct{} {
	return m.stoppedChan
}
