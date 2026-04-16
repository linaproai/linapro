package locker

import (
	"context"
	"fmt"
	"testing"
	"time"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/test/gtest"
)

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

func TestService_New(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		svc := newTestService()
		t.AssertNE(svc, nil)
	})
}

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
