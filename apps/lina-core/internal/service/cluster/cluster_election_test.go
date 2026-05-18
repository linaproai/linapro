// This file tests distributed leader-election lifecycle behavior for the
// cluster service.

package cluster

import (
	"context"
	"testing"
	"time"

	"lina-core/internal/service/config"
	"lina-core/internal/service/coordination"
)

// testElectionCfg is the default election config used in tests.
var testElectionCfg = &config.ElectionConfig{
	Lease:         30 * time.Second,
	RenewInterval: 1 * time.Second,
}

// electionStateWait bounds asynchronous election-state assertions in tests.
const electionStateWait = 3 * time.Second

// newTestElectionService constructs one election service using the shared test
// timing configuration.
func newTestElectionService() *electionService {
	return newElectionService(coordination.NewMemory(nil).Lock(), testElectionCfg, generateNodeIdentifier())
}

// TestElectionServiceNew verifies a new election service exposes an identifier
// and starts in follower mode.
func TestElectionServiceNew(t *testing.T) {
	svc := newTestElectionService()

	if svc == nil {
		t.Fatal("expected election service")
	}
	if svc.Holder() == "" {
		t.Fatal("expected node holder")
	}
	if svc.IsLeader() {
		t.Fatal("expected service to start as follower")
	}
}

// TestElectionServiceStartAndBecomeLeader verifies an unlocked election starts
// and transitions the current node into leader state.
func TestElectionServiceStartAndBecomeLeader(t *testing.T) {
	var (
		svc = newTestElectionService()
		ctx = context.Background()
	)

	svc.Start(ctx)

	if !waitForElectionState(svc, true, electionStateWait) {
		t.Fatal("expected election service to become leader")
	}

	svc.Stop(ctx)

	if svc.IsLeader() {
		t.Fatal("expected service to step down after stop")
	}
}

// TestElectionServiceAlreadyLeader verifies an existing unexpired leader lock
// prevents the current node from becoming leader.
func TestElectionServiceAlreadyLeader(t *testing.T) {
	var (
		coordSvc = coordination.NewMemory(nil)
		svc      = newElectionService(coordSvc.Lock(), testElectionCfg, "node-b")
		ctx      = context.Background()
	)

	_, ok, err := coordSvc.Lock().Acquire(ctx, lockName, "node-a", "leader election", 30*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected first node to acquire leader lock")
	}

	svc.Start(ctx)

	time.Sleep(300 * time.Millisecond)

	if svc.IsLeader() {
		t.Fatal("expected existing leader lock to keep service as follower")
	}

	svc.Stop(ctx)
}

// TestElectionServiceTakeOverExpiredLock verifies the service can take over an
// expired leader lock left by another node.
func TestElectionServiceTakeOverExpiredLock(t *testing.T) {
	var (
		coordSvc = coordination.NewMemory(nil)
		svc      = newElectionService(coordSvc.Lock(), testElectionCfg, "node-b")
		ctx      = context.Background()
	)

	_, ok, err := coordSvc.Lock().Acquire(ctx, lockName, "node-a", "leader election", 20*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected first node to acquire short leader lock")
	}
	time.Sleep(40 * time.Millisecond)

	svc.Start(ctx)

	if !waitForElectionState(svc, true, electionStateWait) {
		t.Fatal("expected service to take over expired leader lock")
	}

	svc.Stop(ctx)
}

// TestElectionServiceStepDown verifies Stop releases leadership after the node
// has successfully become leader.
func TestElectionServiceStepDown(t *testing.T) {
	var (
		svc = newTestElectionService()
		ctx = context.Background()
	)

	svc.Start(ctx)

	if !waitForElectionState(svc, true, electionStateWait) {
		t.Fatal("expected election service to become leader")
	}

	svc.Stop(ctx)

	if svc.IsLeader() {
		t.Fatal("expected service to step down")
	}
}

// TestElectionServiceTwoNodesFailOver verifies two independent election loops
// share one persistent lock and fail over without clearing volatile tables.
func TestElectionServiceTwoNodesFailOver(t *testing.T) {
	var (
		cfg = &config.ElectionConfig{
			Lease:         2 * time.Second,
			RenewInterval: 100 * time.Millisecond,
		}
		coordSvc = coordination.NewMemory(nil)
		first    = newElectionService(coordSvc.Lock(), cfg, "node-a")
		second   = newElectionService(coordSvc.Lock(), cfg, "node-b")
		ctx      = context.Background()
	)

	first.Start(ctx)
	second.Start(ctx)

	if !waitForAnyElectionLeader(first, second, electionStateWait) {
		t.Fatal("expected exactly one election service to become leader")
	}
	if first.IsLeader() && second.IsLeader() {
		t.Fatal("expected at most one leader")
	}

	firstWasLeader := first.IsLeader()
	if firstWasLeader {
		first.Stop(ctx)
		if !waitForElectionState(second, true, 4*time.Second) {
			t.Fatal("expected second node to become leader after first stops")
		}
		second.Stop(ctx)
	} else {
		second.Stop(ctx)
		if !waitForElectionState(first, true, 4*time.Second) {
			t.Fatal("expected first node to become leader after second stops")
		}
		first.Stop(ctx)
	}
}

// TestElectionServiceStopWithoutStart verifies Stop is safe before Start is
// called.
func TestElectionServiceStopWithoutStart(t *testing.T) {
	svc := newTestElectionService()
	svc.Stop(context.Background())
}

// TestElectionServiceNonLeaderRetry verifies the retry loop eventually acquires
// leadership after a competing lock is released.
func TestElectionServiceNonLeaderRetry(t *testing.T) {
	var (
		coordSvc = coordination.NewMemory(nil)
		retryCfg = &config.ElectionConfig{
			Lease:         30 * time.Second,
			RenewInterval: 200 * time.Millisecond,
		}
		svc = newElectionService(coordSvc.Lock(), retryCfg, "node-b")
		ctx = context.Background()
	)

	handle, ok, err := coordSvc.Lock().Acquire(ctx, lockName, "node-a", "leader election", 30*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected first node to acquire leader lock")
	}

	svc.Start(ctx)

	time.Sleep(300 * time.Millisecond)
	if svc.IsLeader() {
		t.Fatal("expected service to remain follower before lock release")
	}
	if err = coordSvc.Lock().Release(ctx, handle); err != nil {
		t.Fatal(err)
	}

	if !waitForElectionState(svc, true, electionStateWait) {
		t.Fatal("expected retry loop to acquire released leader lock")
	}

	svc.Stop(ctx)
}

// waitForElectionState polls the asynchronous election loop until it reaches
// the expected leadership state or the bounded timeout expires.
func waitForElectionState(svc *electionService, expected bool, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if svc.IsLeader() == expected {
			return true
		}
		time.Sleep(20 * time.Millisecond)
	}
	return svc.IsLeader() == expected
}

// waitForAnyElectionLeader polls until exactly one of two election services is
// leader or the bounded timeout expires.
func waitForAnyElectionLeader(first *electionService, second *electionService, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if first.IsLeader() != second.IsLeader() {
			return true
		}
		time.Sleep(20 * time.Millisecond)
	}
	return first.IsLeader() != second.IsLeader()
}
