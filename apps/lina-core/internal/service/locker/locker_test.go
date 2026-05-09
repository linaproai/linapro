// This file tests locker service acquisition and lock-function behavior
// against the persistent lock table.

package locker

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/test/gtest"
)

// testHolder is the holder token shared by locker integration tests.
const testHolder = "test-node"

// newTestService creates a new locker service for testing.
func newTestService() *serviceImpl {
	return New().(*serviceImpl)
}

// cleanupLock removes the lock by name after test.
func cleanupLock(name string) {
	if _, err := g.DB().Model("sys_locker").Where("name", name).Delete(); err != nil {
		panic(fmt.Sprintf("cleanup locker row failed name=%s err=%v", name, err))
	}
}

// TestService_New verifies New returns a non-nil locker service.
func TestService_New(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		svc := newTestService()
		t.AssertNE(svc, nil)
	})
}

// TestService_Lock_NewLock verifies acquiring a missing lock creates one new
// persistent lock row.
func TestService_Lock_NewLock(t *testing.T) {
	var (
		svc    = newTestService()
		name   = "test-lock-new-" + gtime.TimestampMilliStr()
		reason = "test reason"
		ctx    = context.Background()
	)

	cleanupLock(name)

	gtest.C(t, func(t *gtest.T) {
		instance, ok, err := svc.Lock(ctx, name, testHolder, reason, 30*time.Second)
		t.AssertNil(err)
		t.Assert(ok, true)
		t.AssertNE(instance, nil)

		count, err := g.DB().Model("sys_locker").Where("name", name).Count()
		t.AssertNil(err)
		t.Assert(count, 1)

		err = instance.Unlock(ctx)
		t.AssertNil(err)
	})

	cleanupLock(name)
}

// TestService_Lock_ExistingExpiredLock verifies an expired lock can be taken
// over and rewritten by the current holder.
func TestService_Lock_ExistingExpiredLock(t *testing.T) {
	var (
		svc    = newTestService()
		name   = "test-lock-expired-" + gtime.TimestampMilliStr()
		reason = "test reason"
		ctx    = context.Background()
	)

	cleanupLock(name)

	_, err := g.DB().Model("sys_locker").Data(g.Map{
		"name":        name,
		"reason":      "old reason",
		"holder":      "other-node",
		"expire_time": gtime.Now().Add(-10 * time.Second),
	}).Insert()
	if err != nil {
		t.Fatal(err)
	}

	gtest.C(t, func(t *gtest.T) {
		instance, ok, err := svc.Lock(ctx, name, testHolder, reason, 30*time.Second)
		t.AssertNil(err)
		t.Assert(ok, true)
		t.AssertNE(instance, nil)

		var row struct {
			Holder string
			Reason string
		}
		err = g.DB().Model("sys_locker").Where("name", name).Scan(&row)
		t.AssertNil(err)
		t.Assert(row.Holder, testHolder)
		t.Assert(row.Reason, reason)

		err = instance.Unlock(ctx)
		t.AssertNil(err)
	})

	cleanupLock(name)
}

// TestIsExpiredLockUsesExpireTime verifies lock takeover decisions are based
// on expire_time before holder data is reused.
func TestIsExpiredLockUsesExpireTime(t *testing.T) {
	now := gtime.Now()
	if !isExpiredLock(now.Add(-time.Second), now) {
		t.Fatal("expected past expire_time to be expired")
	}
	if isExpiredLock(now.Add(time.Second), now) {
		t.Fatal("expected future expire_time to remain held")
	}
}

// TestService_Lock_ExistingNonExpiredLock verifies a lock held by another node
// cannot be acquired before expiry.
func TestService_Lock_ExistingNonExpiredLock(t *testing.T) {
	var (
		svc    = newTestService()
		name   = "test-lock-active-" + gtime.TimestampMilliStr()
		reason = "test reason"
		ctx    = context.Background()
	)

	cleanupLock(name)

	_, err := g.DB().Model("sys_locker").Data(g.Map{
		"name":        name,
		"reason":      "old reason",
		"holder":      "other-node",
		"expire_time": gtime.Now().Add(30 * time.Second),
	}).Insert()
	if err != nil {
		t.Fatal(err)
	}

	gtest.C(t, func(t *gtest.T) {
		instance, ok, err := svc.Lock(ctx, name, testHolder, reason, 30*time.Second)
		t.AssertNil(err)
		t.Assert(ok, false)
		t.Assert(instance, nil)
	})

	cleanupLock(name)
}

// TestService_Lock_RecordSurvivesServiceRecreation verifies a valid lock row
// remains effective when a new service instance is constructed after restart.
func TestService_Lock_RecordSurvivesServiceRecreation(t *testing.T) {
	var (
		firstService  = newTestService()
		secondService = newTestService()
		name          = "test-lock-restart-" + gtime.TimestampMilliStr()
		reason        = "test reason"
		ctx           = context.Background()
	)

	cleanupLock(name)

	instance, ok, err := firstService.Lock(ctx, name, testHolder, reason, 30*time.Second)
	if err != nil {
		t.Fatalf("acquire lock before service recreation: %v", err)
	}
	if !ok || instance == nil {
		t.Fatal("expected first service to acquire lock")
	}

	restartedInstance, ok, err := secondService.Lock(ctx, name, "other-node", "after restart", 30*time.Second)
	if err != nil {
		t.Fatalf("acquire lock after service recreation: %v", err)
	}
	if ok || restartedInstance != nil {
		t.Fatal("expected valid lock to remain held after service recreation")
	}

	if err = instance.Unlock(ctx); err != nil {
		t.Fatalf("unlock retained lock: %v", err)
	}
	cleanupLock(name)
}

// TestService_Lock_ConcurrentFreshLockRace verifies duplicate insert races are
// reported as clean acquisition misses instead of database errors.
func TestService_Lock_ConcurrentFreshLockRace(t *testing.T) {
	var (
		name   = "test-lock-concurrent-" + gtime.TimestampMilliStr()
		reason = "test concurrent reason"
		ctx    = context.Background()
	)

	cleanupLock(name)
	t.Cleanup(func() {
		cleanupLock(name)
	})

	const contenders = 16
	var (
		start      = make(chan struct{})
		ready      sync.WaitGroup
		done       sync.WaitGroup
		successes  int32
		failures   int32
		firstError = make(chan error, 1)
	)

	for i := 0; i < contenders; i++ {
		holder := fmt.Sprintf("test-holder-%02d", i)
		ready.Add(1)
		done.Add(1)
		go func(holder string) {
			defer done.Done()
			svc := newTestService()
			ready.Done()
			<-start
			instance, ok, err := svc.Lock(ctx, name, holder, reason, 30*time.Second)
			if err != nil {
				select {
				case firstError <- err:
				default:
				}
				atomic.AddInt32(&failures, 1)
				return
			}
			if ok {
				atomic.AddInt32(&successes, 1)
				if instance == nil {
					select {
					case firstError <- fmt.Errorf("nil instance for holder %s", holder):
					default:
					}
				}
				return
			}
			if instance != nil {
				select {
				case firstError <- fmt.Errorf("unexpected instance for holder %s", holder):
				default:
				}
			}
		}(holder)
	}
	ready.Wait()
	close(start)
	done.Wait()

	select {
	case errValue := <-firstError:
		t.Fatalf("concurrent lock acquisition surfaced error: %v", errValue)
	default:
	}
	if failures != 0 {
		t.Fatalf("expected no failed acquisitions, got %d", failures)
	}
	if successes != 1 {
		t.Fatalf("expected exactly one successful acquisition, got %d", successes)
	}
	count, err := g.DB().Model("sys_locker").Where("name", name).Count()
	if err != nil {
		t.Fatalf("count lock row failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected exactly one lock row, got %d", count)
	}
}

// TestService_Lock_SameHolder verifies the current holder can reacquire its
// own lock and extend ownership.
func TestService_Lock_SameHolder(t *testing.T) {
	var (
		svc    = newTestService()
		name   = "test-lock-same-holder-" + gtime.TimestampMilliStr()
		reason = "test reason"
		ctx    = context.Background()
	)

	cleanupLock(name)

	gtest.C(t, func(t *gtest.T) {
		instance1, ok1, err1 := svc.Lock(ctx, name, testHolder, reason, 30*time.Second)
		t.AssertNil(err1)
		t.Assert(ok1, true)

		instance2, ok2, err2 := svc.Lock(ctx, name, testHolder, "new reason", 30*time.Second)
		t.AssertNil(err2)
		t.Assert(ok2, true)

		unlockErr := instance1.Unlock(ctx)
		t.AssertNil(unlockErr)
		unlockErr = instance2.Unlock(ctx)
		t.AssertNil(unlockErr)
	})

	cleanupLock(name)
}

// TestService_LockFunc verifies LockFunc executes the callback and releases the
// lock afterward.
func TestService_LockFunc(t *testing.T) {
	var (
		svc      = newTestService()
		name     = "test-lock-func-" + gtime.TimestampMilliStr()
		executed = false
		reason   = "test reason"
		ctx      = context.Background()
	)

	cleanupLock(name)

	gtest.C(t, func(t *gtest.T) {
		ok, err := svc.LockFunc(ctx, name, testHolder, reason, 30*time.Second, func() error {
			executed = true
			return nil
		})
		t.AssertNil(err)
		t.Assert(ok, true)
		t.Assert(executed, true)

		count, err := g.DB().Model("sys_locker").Where("name", name).
			WhereGTE("expire_time", gtime.Now()).Count()
		t.AssertNil(err)
		t.Assert(count, 0)
	})

	cleanupLock(name)

	gtest.C(t, func(t *gtest.T) {
		executed = false
		ok, err := svc.LockFunc(ctx, name, testHolder, reason, 30*time.Second, func() error {
			executed = true
			return ErrLockNotHeld
		})
		t.AssertNE(err, nil)
		t.Assert(ok, true)
		t.Assert(executed, true)
	})

	cleanupLock(name)
}

// TestService_LockFunc_AlreadyLocked verifies LockFunc does not execute the
// callback when another holder already owns the lock.
func TestService_LockFunc_AlreadyLocked(t *testing.T) {
	var (
		svc    = newTestService()
		name   = "test-lock-func-locked-" + gtime.TimestampMilliStr()
		reason = "test reason"
		ctx    = context.Background()
	)

	cleanupLock(name)

	_, err := g.DB().Model("sys_locker").Data(g.Map{
		"name":        name,
		"reason":      "other reason",
		"holder":      "other-node",
		"expire_time": gtime.Now().Add(30 * time.Second),
	}).Insert()
	if err != nil {
		t.Fatal(err)
	}

	gtest.C(t, func(t *gtest.T) {
		executed := false
		ok, err := svc.LockFunc(ctx, name, testHolder, reason, 30*time.Second, func() error {
			executed = true
			return nil
		})
		t.AssertNil(err)
		t.Assert(ok, false)
		t.Assert(executed, false)
	})

	cleanupLock(name)
}
