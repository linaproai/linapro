package cluster

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"lina-core/internal/service/config"
	"lina-core/internal/service/locker"
	"lina-core/pkg/logger"
)

// lockName is the distributed lock name used for leader election.
const lockName = "leader-election"

type electionService struct {
	locker      locker.Service         // locker manages distributed lock ownership.
	cfg         *config.ElectionConfig // cfg stores election lease and renew settings.
	holder      string                 // holder is the current node identifier.
	isLeader    atomic.Bool            // isLeader reports whether the current node owns leadership.
	instance    *locker.Instance       // instance is the current lock instance when leadership is held.
	leaseMgr    *locker.LeaseManager   // leaseMgr keeps the lock lease renewed while leader.
	stopChan    chan struct{}
	stoppedChan chan struct{}
	once        sync.Once // once ensures Start is only executed once.
	stoppedOnce sync.Once // stoppedOnce ensures stoppedChan is closed exactly once.
}

func newElectionService(
	lockerSvc locker.Service,
	cfg *config.ElectionConfig,
	holder string,
) *electionService {
	service := &electionService{
		locker:   lockerSvc,
		cfg:      cfg,
		holder:   holder,
		stopChan: make(chan struct{}),
	}

	// Pre-close stoppedChan so Stop is safe before Start is called.
	service.stoppedChan = make(chan struct{})
	close(service.stoppedChan)
	return service
}

func (s *electionService) Start(ctx context.Context) {
	s.once.Do(func() {
		s.stoppedChan = make(chan struct{})
		go s.run(ctx)
	})
}

func (s *electionService) Stop(ctx context.Context) {
	select {
	case <-s.stopChan:
	default:
		close(s.stopChan)
	}
	<-s.stoppedChan
}

func (s *electionService) IsLeader() bool {
	return s.isLeader.Load()
}

func (s *electionService) Holder() string {
	return s.holder
}

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
			s.instance = nil
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

func (s *electionService) tryAcquire(ctx context.Context) {
	instance, ok, err := s.locker.Lock(ctx, lockName, s.holder, "leader election", s.cfg.Lease)
	if err != nil {
		logger.Warningf(ctx, "[cluster] failed to acquire leader lock: %v", err)
		s.isLeader.Store(false)
		return
	}

	if ok {
		s.instance = instance
		s.isLeader.Store(true)
		logger.Infof(ctx, "[cluster] became leader (holder: %s)", s.holder)

		s.leaseMgr = locker.NewLeaseManager(instance, s.cfg.RenewInterval)
		s.leaseMgr.Start(ctx)
		return
	}

	s.isLeader.Store(false)
	logger.Debugf(ctx, "[cluster] not leader, waiting for lease expiry")
}

func (s *electionService) stepDown(ctx context.Context) {
	if s.leaseMgr != nil {
		s.leaseMgr.Stop()
		s.leaseMgr = nil
	}
	if s.instance != nil {
		if err := s.instance.Unlock(ctx); err != nil {
			logger.Warningf(ctx, "[cluster] failed to release leader lock: %v", err)
		}
		s.instance = nil
	}

	s.isLeader.Store(false)
	logger.Infof(ctx, "[cluster] stepped down from leadership")
}

func (s *electionService) leaseStoppedChan() <-chan struct{} {
	if s.leaseMgr != nil {
		return s.leaseMgr.StoppedChan()
	}
	return nil
}
