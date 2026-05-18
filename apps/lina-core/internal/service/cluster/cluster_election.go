// This file implements the distributed primary-election loop used when
// clustered deployment mode is enabled.

package cluster

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"lina-core/internal/service/config"
	"lina-core/internal/service/coordination"
	"lina-core/pkg/logger"
)

// lockName is the distributed lock name used for leader election.
const lockName = "leader-election"

// electionService coordinates distributed leader election using the locker
// service and a renewable lease.
type electionService struct {
	lockStore   coordination.LockStore // lockStore manages distributed lock ownership.
	cfg         *config.ElectionConfig // cfg stores election lease and renew settings.
	holder      string                 // holder is the current node identifier.
	isLeader    atomic.Bool            // isLeader reports whether the current node owns leadership.
	handle      *coordination.LockHandle
	leaseMgr    *electionLeaseManager
	stopChan    chan struct{}
	stoppedChan chan struct{}
	once        sync.Once // once ensures Start is only executed once.
	stoppedOnce sync.Once // stoppedOnce ensures stoppedChan is closed exactly once.
}

// newElectionService constructs one election service with safe stop semantics
// before Start is ever called.
func newElectionService(
	lockStore coordination.LockStore,
	cfg *config.ElectionConfig,
	holder string,
) *electionService {
	service := &electionService{
		lockStore: lockStore,
		cfg:       cfg,
		holder:    holder,
		stopChan:  make(chan struct{}),
	}

	// Pre-close stoppedChan so Stop is safe before Start is called.
	service.stoppedChan = make(chan struct{})
	close(service.stoppedChan)
	return service
}

// Start launches the background election loop once.
func (s *electionService) Start(ctx context.Context) {
	s.once.Do(func() {
		s.stoppedChan = make(chan struct{})
		go s.run(ctx)
	})
}

// Stop signals the background election loop to exit and waits for shutdown.
func (s *electionService) Stop(ctx context.Context) {
	select {
	case <-s.stopChan:
	default:
		close(s.stopChan)
	}
	<-s.stoppedChan
}

// IsLeader reports whether the current node currently owns the leader lease.
func (s *electionService) IsLeader() bool {
	return s.isLeader.Load()
}

// Holder returns the node identifier that this service uses when competing for
// leadership.
func (s *electionService) Holder() string {
	return s.holder
}

// run drives acquisition, renewal-failure recovery, and shutdown handling for
// leader election.
func (s *electionService) run(ctx context.Context) {
	defer s.stoppedOnce.Do(func() { close(s.stoppedChan) })

	s.tryAcquire(ctx)

	retryTicker := time.NewTicker(s.cfg.RenewInterval)
	defer retryTicker.Stop()

	for {
		select {
		case <-s.stopChan:
			s.stepDown(ctx)
			logger.Infof(ctx, "[cluster] leader election stopped")
			return
		case <-s.leaseStoppedChan():
			logger.Warningf(ctx, "[cluster] lease renewal stopped, attempting to re-acquire")
			s.handle = nil
			s.leaseMgr = nil
			s.isLeader.Store(false)
			retryTicker.Reset(s.cfg.RenewInterval)
			s.tryAcquire(ctx)
		case <-retryTicker.C:
			if !s.isLeader.Load() {
				s.tryAcquire(ctx)
			}
		}
	}
}

// tryAcquire attempts to obtain the leader lock and starts lease renewal when
// successful.
func (s *electionService) tryAcquire(ctx context.Context) {
	if s.lockStore == nil {
		s.isLeader.Store(false)
		return
	}
	handle, ok, err := s.lockStore.Acquire(ctx, lockName, s.holder, "leader election", s.cfg.Lease)
	if err != nil {
		logger.Warningf(ctx, "[cluster] failed to acquire leader lock: %v", err)
		s.isLeader.Store(false)
		return
	}

	if ok {
		s.handle = handle
		s.isLeader.Store(true)
		logger.Infof(ctx, "[cluster] became leader (holder: %s)", s.holder)

		s.leaseMgr = newElectionLeaseManager(s.lockStore, handle, s.cfg.Lease, s.cfg.RenewInterval)
		s.leaseMgr.Start(ctx)
		return
	}

	s.isLeader.Store(false)
	logger.Debugf(ctx, "[cluster] not leader, waiting for lease expiry")
}

// stepDown stops renewal and releases the leader lock when the node currently
// holds it.
func (s *electionService) stepDown(ctx context.Context) {
	if s.leaseMgr != nil {
		s.leaseMgr.Stop()
		s.leaseMgr = nil
	}
	if s.handle != nil && s.lockStore != nil {
		if err := s.lockStore.Release(ctx, s.handle); err != nil {
			logger.Warningf(ctx, "[cluster] failed to release leader lock: %v", err)
		}
		s.handle = nil
	}

	s.isLeader.Store(false)
	logger.Infof(ctx, "[cluster] stepped down from leadership")
}

// leaseStoppedChan returns the renewal manager stop signal when leadership is
// currently held.
func (s *electionService) leaseStoppedChan() <-chan struct{} {
	if s.leaseMgr != nil {
		return s.leaseMgr.StoppedChan()
	}
	return nil
}
