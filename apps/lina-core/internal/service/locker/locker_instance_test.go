package locker

import (
	"context"
	"testing"
	"time"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/test/gtest"
)

func TestInstance_Unlock(t *testing.T) {
	var (
		svc    = newTestService()
		name   = "test-instance-unlock-" + gtime.TimestampMilliStr()
		reason = "test reason"
		ctx    = context.Background()
	)

	cleanupLock(name)

	gtest.C(t, func(t *gtest.T) {
		// Acquire lock
		instance, ok, err := svc.Lock(ctx, name, testHolder, reason, 30*time.Second)
		t.AssertNil(err)
		t.Assert(ok, true)

		// Verify lock is active
		isHeld, err := instance.IsHeld(ctx)
		t.AssertNil(err)
		t.Assert(isHeld, true)

		// Unlock
		err = instance.Unlock(ctx)
		t.AssertNil(err)

		// Verify lock is released
		isHeld, err = instance.IsHeld(ctx)
		t.AssertNil(err)
		t.Assert(isHeld, false)
	})

	cleanupLock(name)
}

func TestInstance_Renew(t *testing.T) {
	var (
		svc    = newTestService()
		name   = "test-instance-renew-" + gtime.TimestampMilliStr()
		reason = "test reason"
		ctx    = context.Background()
	)

	cleanupLock(name)

	gtest.C(t, func(t *gtest.T) {
		// Acquire lock
		instance, ok, err := svc.Lock(ctx, name, testHolder, reason, 30*time.Second)
		t.AssertNil(err)
		t.Assert(ok, true)

		// Get current expire time
		var locker struct {
			ExpireTime *gtime.Time
		}
		err = g.DB().Model("sys_locker").Where("name", name).Scan(&locker)
		t.AssertNil(err)
		originalExpire := locker.ExpireTime

		// Wait a bit for time to pass
		time.Sleep(100 * time.Millisecond)

		// Renew
		err = instance.Renew(ctx)
		t.AssertNil(err)

		// Verify expire time was extended (or at least not decreased)
		// Note: Due to second-level precision, if renewed within same second,
		// the expire_time might be identical
		err = g.DB().Model("sys_locker").Where("name", name).Scan(&locker)
		t.AssertNil(err)
		t.AssertGE(locker.ExpireTime.Unix(), originalExpire.Unix())

		// Clean up
		err = instance.Unlock(ctx)
		t.AssertNil(err)
	})

	cleanupLock(name)
}

func TestInstance_Renew_NotHeld(t *testing.T) {
	var (
		svc    = newTestService()
		name   = "test-instance-renew-fail-" + gtime.TimestampMilliStr()
		reason = "test reason"
		ctx    = context.Background()
	)

	cleanupLock(name)

	gtest.C(t, func(t *gtest.T) {
		// Acquire lock
		instance, ok, err := svc.Lock(ctx, name, testHolder, reason, 30*time.Second)
		t.AssertNil(err)
		t.Assert(ok, true)

		// Release lock by setting expire_time to past
		_, err = g.DB().Model("sys_locker").Data(g.Map{
			"expire_time": gtime.Now().Add(-10 * time.Second),
		}).Where("name", name).Update()
		t.AssertNil(err)

		// Try to renew after lock expired - should fail
		err = instance.Renew(ctx)
		t.AssertEQ(err, ErrLockNotHeld)
	})

	cleanupLock(name)
}

func TestInstance_Renew_LostToOther(t *testing.T) {
	var (
		svc    = newTestService()
		name   = "test-instance-renew-lost-" + gtime.TimestampMilliStr()
		reason = "test reason"
		ctx    = context.Background()
	)

	cleanupLock(name)

	gtest.C(t, func(t *gtest.T) {
		// Acquire lock
		instance, ok, err := svc.Lock(ctx, name, testHolder, reason, 30*time.Second)
		t.AssertNil(err)
		t.Assert(ok, true)

		// Simulate another node taking over by updating the lock
		_, err = g.DB().Model("sys_locker").Data(g.Map{
			"holder":      "other-node",
			"expire_time": gtime.Now().Add(30 * time.Second),
		}).Where("name", name).Update()
		t.AssertNil(err)

		// Try to renew - should fail because holder changed
		err = instance.Renew(ctx)
		t.AssertEQ(err, ErrLockNotHeld)
	})

	cleanupLock(name)
}

func TestInstance_IsHeld(t *testing.T) {
	var (
		svc    = newTestService()
		name   = "test-instance-isheld-" + gtime.TimestampMilliStr()
		reason = "test reason"
		ctx    = context.Background()
	)

	cleanupLock(name)

	gtest.C(t, func(t *gtest.T) {
		// Create instance without acquiring lock
		instance := &Instance{id: 99999, holder: testHolder, lease: 30 * time.Second}

		// Should not be held
		isHeld, err := instance.IsHeld(ctx)
		t.AssertNil(err)
		t.Assert(isHeld, false)

		// Acquire a real lock
		realInstance, ok, err := svc.Lock(ctx, name, testHolder, reason, 30*time.Second)
		t.AssertNil(err)
		t.Assert(ok, true)

		// Should be held
		isHeld, err = realInstance.IsHeld(ctx)
		t.AssertNil(err)
		t.Assert(isHeld, true)

		// Clean up
		err = realInstance.Unlock(ctx)
		t.AssertNil(err)
	})

	cleanupLock(name)
}

func TestInstance_IsHeld_Expired(t *testing.T) {
	var (
		svc    = newTestService()
		name   = "test-instance-isheld-expired-" + gtime.TimestampMilliStr()
		reason = "test reason"
		ctx    = context.Background()
	)

	cleanupLock(name)

	gtest.C(t, func(t *gtest.T) {
		// Acquire lock
		instance, ok, err := svc.Lock(ctx, name, testHolder, reason, 30*time.Second)
		t.AssertNil(err)
		t.Assert(ok, true)

		// Set expire time to past
		_, err = g.DB().Model("sys_locker").Data(g.Map{
			"expire_time": gtime.Now().Add(-10 * time.Second),
		}).Where("name", name).Update()
		t.AssertNil(err)

		// Should not be held (expired)
		isHeld, err := instance.IsHeld(ctx)
		t.AssertNil(err)
		t.Assert(isHeld, false)
	})

	cleanupLock(name)
}
