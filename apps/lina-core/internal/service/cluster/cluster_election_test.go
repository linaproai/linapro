// This file tests distributed leader-election lifecycle behavior for the
// cluster service.

package cluster

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/test/gtest"
	"testing"
	"time"

	"lina-core/internal/service/config"
	"lina-core/internal/service/locker"
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
	return newElectionService(locker.New(), testElectionCfg, generateNodeIdentifier())
}

// TestElectionServiceNew verifies a new election service exposes an identifier
// and starts in follower mode.
func TestElectionServiceNew(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		svc := newTestElectionService()

		t.AssertNE(svc, nil)
		t.AssertNE(svc.Holder(), "")
		t.Assert(svc.IsLeader(), false)
	})
}

// TestElectionServiceStartAndBecomeLeader verifies an unlocked election starts
// and transitions the current node into leader state.
func TestElectionServiceStartAndBecomeLeader(t *testing.T) {
	var (
		svc = newTestElectionService()
		ctx = context.Background()
	)

	cleanupLock()

	gtest.C(t, func(t *gtest.T) {
		svc.Start(ctx)

		t.Assert(waitForElectionState(svc, true, electionStateWait), true)

		count, err := g.DB().Model("sys_locker").Where("name", lockName).Count()
		t.AssertNil(err)
		t.Assert(count, 1)

		svc.Stop(ctx)

		t.Assert(svc.IsLeader(), false)
	})

	cleanupLock()
}

// TestElectionServiceAlreadyLeader verifies an existing unexpired leader lock
// prevents the current node from becoming leader.
func TestElectionServiceAlreadyLeader(t *testing.T) {
	var (
		svc = newTestElectionService()
		ctx = context.Background()
	)

	cleanupLock()

	_, err := g.DB().Model("sys_locker").Data(g.Map{
		"name":        lockName,
		"reason":      "election",
		"holder":      "other-node",
		"expire_time": gtime.Now().Add(30 * time.Second),
	}).Insert()
	if err != nil {
		t.Fatal(err)
	}

	gtest.C(t, func(t *gtest.T) {
		svc.Start(ctx)

		time.Sleep(300 * time.Millisecond)

		t.Assert(svc.IsLeader(), false)

		svc.Stop(ctx)
	})

	cleanupLock()
}

// TestElectionServiceTakeOverExpiredLock verifies the service can take over an
// expired leader lock left by another node.
func TestElectionServiceTakeOverExpiredLock(t *testing.T) {
	var (
		svc = newTestElectionService()
		ctx = context.Background()
	)

	cleanupLock()

	_, err := g.DB().Model("sys_locker").Data(g.Map{
		"name":        lockName,
		"reason":      "election",
		"holder":      "other-node",
		"expire_time": gtime.Now().Add(-10 * time.Second),
	}).Insert()
	if err != nil {
		t.Fatal(err)
	}

	gtest.C(t, func(t *gtest.T) {
		svc.Start(ctx)

		t.Assert(waitForElectionState(svc, true, electionStateWait), true)

		var row struct{ Holder string }
		err = g.DB().Model("sys_locker").Where("name", lockName).Scan(&row)
		t.AssertNil(err)
		t.Assert(row.Holder, svc.Holder())

		svc.Stop(ctx)
	})

	cleanupLock()
}

// TestElectionServiceStepDown verifies Stop releases leadership after the node
// has successfully become leader.
func TestElectionServiceStepDown(t *testing.T) {
	var (
		svc = newTestElectionService()
		ctx = context.Background()
	)

	cleanupLock()

	gtest.C(t, func(t *gtest.T) {
		svc.Start(ctx)

		t.Assert(waitForElectionState(svc, true, electionStateWait), true)

		svc.Stop(ctx)

		t.Assert(svc.IsLeader(), false)
	})

	cleanupLock()
}

// TestElectionServiceTwoNodesFailOver verifies two independent election loops
// share one persistent lock and fail over without clearing volatile tables.
func TestElectionServiceTwoNodesFailOver(t *testing.T) {
	var (
		cfg = &config.ElectionConfig{
			Lease:         2 * time.Second,
			RenewInterval: 100 * time.Millisecond,
		}
		first  = newElectionService(locker.New(), cfg, "node-a-"+gtime.TimestampMilliStr())
		second = newElectionService(locker.New(), cfg, "node-b-"+gtime.TimestampMilliStr())
		ctx    = context.Background()
	)

	cleanupLock()

	gtest.C(t, func(t *gtest.T) {
		first.Start(ctx)
		second.Start(ctx)

		t.Assert(waitForAnyElectionLeader(first, second, electionStateWait), true)
		t.Assert(first.IsLeader() && second.IsLeader(), false)

		firstWasLeader := first.IsLeader()
		if firstWasLeader {
			first.Stop(ctx)
			t.Assert(waitForElectionState(second, true, 4*time.Second), true)
			second.Stop(ctx)
		} else {
			second.Stop(ctx)
			t.Assert(waitForElectionState(first, true, 4*time.Second), true)
			first.Stop(ctx)
		}
	})

	cleanupLock()
}

// TestElectionServiceStopWithoutStart verifies Stop is safe before Start is
// called.
func TestElectionServiceStopWithoutStart(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		svc := newTestElectionService()
		svc.Stop(context.Background())
	})
}

// TestElectionServiceNonLeaderRetry verifies the retry loop eventually acquires
// leadership after a competing expired lock is observed.
func TestElectionServiceNonLeaderRetry(t *testing.T) {
	var (
		retryCfg = &config.ElectionConfig{
			Lease:         30 * time.Second,
			RenewInterval: 200 * time.Millisecond,
		}
		svc = newElectionService(locker.New(), retryCfg, generateNodeIdentifier())
		ctx = context.Background()
	)

	cleanupLock()

	_, err := g.DB().Model("sys_locker").Data(g.Map{
		"name":        lockName,
		"reason":      "election",
		"holder":      "other-node",
		"expire_time": gtime.Now().Add(-5 * time.Second),
	}).Insert()
	if err != nil {
		t.Fatal(err)
	}

	gtest.C(t, func(t *gtest.T) {
		svc.Start(ctx)

		t.Assert(waitForElectionState(svc, true, electionStateWait), true)

		svc.Stop(ctx)
	})

	cleanupLock()
}

// cleanupLock removes the shared election lock row between test runs.
func cleanupLock() {
	if _, err := g.DB().Model("sys_locker").Where("name", lockName).Delete(); err != nil {
		panic(fmt.Sprintf("cleanup leader-election lock failed: %v", err))
	}
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
