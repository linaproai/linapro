package cluster

import (
	"context"
	"fmt"
	"testing"
	"time"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/test/gtest"

	"lina-core/internal/service/config"
	"lina-core/internal/service/locker"
)

// testElectionCfg is the default election config used in tests.
var testElectionCfg = &config.ElectionConfig{
	Lease:         30 * time.Second,
	RenewInterval: 1 * time.Second,
}

func newTestElectionService() *electionService {
	return newElectionService(locker.New(), testElectionCfg, generateNodeIdentifier())
}

func TestElectionServiceNew(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		svc := newTestElectionService()

		t.AssertNE(svc, nil)
		t.AssertNE(svc.Holder(), "")
		t.Assert(svc.IsLeader(), false)
	})
}

func TestElectionServiceStartAndBecomeLeader(t *testing.T) {
	var (
		svc = newTestElectionService()
		ctx = context.Background()
	)

	cleanupLock()

	gtest.C(t, func(t *gtest.T) {
		svc.Start(ctx)

		time.Sleep(200 * time.Millisecond)

		t.Assert(svc.IsLeader(), true)

		count, err := g.DB().Model("sys_locker").Where("name", lockName).Count()
		t.AssertNil(err)
		t.Assert(count, 1)

		svc.Stop(ctx)

		t.Assert(svc.IsLeader(), false)
	})

	cleanupLock()
}

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

		time.Sleep(200 * time.Millisecond)

		t.Assert(svc.IsLeader(), false)

		svc.Stop(ctx)
	})

	cleanupLock()
}

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

		time.Sleep(200 * time.Millisecond)

		t.Assert(svc.IsLeader(), true)

		var row struct{ Holder string }
		err = g.DB().Model("sys_locker").Where("name", lockName).Scan(&row)
		t.AssertNil(err)
		t.Assert(row.Holder, svc.Holder())

		svc.Stop(ctx)
	})

	cleanupLock()
}

func TestElectionServiceStepDown(t *testing.T) {
	var (
		svc = newTestElectionService()
		ctx = context.Background()
	)

	cleanupLock()

	gtest.C(t, func(t *gtest.T) {
		svc.Start(ctx)

		time.Sleep(200 * time.Millisecond)

		t.Assert(svc.IsLeader(), true)

		svc.Stop(ctx)

		t.Assert(svc.IsLeader(), false)
	})

	cleanupLock()
}

func TestElectionServiceStopWithoutStart(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		svc := newTestElectionService()
		svc.Stop(context.Background())
	})
}

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

		time.Sleep(500 * time.Millisecond)

		t.Assert(svc.IsLeader(), true)

		svc.Stop(ctx)
	})

	cleanupLock()
}

func cleanupLock() {
	if _, err := g.DB().Model("sys_locker").Where("name", lockName).Delete(); err != nil {
		panic(fmt.Sprintf("cleanup leader-election lock failed: %v", err))
	}
}
