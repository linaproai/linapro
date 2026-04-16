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

func TestLeaseManager_StartAndStop(t *testing.T) {
	var (
		svc    = New()
		name   = "test-lease-start-stop-" + gtime.TimestampMilliStr()
		reason = "test reason"
		ctx    = context.Background()
	)

	cleanupLock(name)

	// Acquire lock
	instance, ok, err := svc.Lock(ctx, name, testHolder, reason, 30*time.Second)
	if err != nil || !ok {
		t.Fatal(err)
	}

	gtest.C(t, func(t *gtest.T) {
		// Create lease manager
		lm := NewLeaseManager(instance, 1*time.Second)
		t.AssertNE(lm, nil)

		// Start lease renewal
		lm.Start(ctx)

		// Wait for at least one renewal cycle
		time.Sleep(1500 * time.Millisecond)

		// Verify lease was renewed (expire_time should be in the future)
		var locker struct {
			ExpireTime *gtime.Time
		}
		err = g.DB().Model("sys_locker").Where("name", name).Scan(&locker)
		t.AssertNil(err)
		t.AssertGT(locker.ExpireTime.Unix(), gtime.Now().Unix())

		// Stop lease manager
		lm.Stop()
	})

	if err = instance.Unlock(ctx); err != nil {
		t.Fatal(err)
	}
	cleanupLock(name)
}

func TestLeaseManager_StopChan(t *testing.T) {
	var (
		svc    = New()
		name   = "test-lease-chan-" + gtime.TimestampMilliStr()
		reason = "test reason"
		ctx    = context.Background()
	)

	cleanupLock(name)

	// Acquire lock
	instance, ok, err := svc.Lock(ctx, name, testHolder, reason, 30*time.Second)
	if err != nil || !ok {
		t.Fatal(err)
	}

	gtest.C(t, func(t *gtest.T) {
		// Create lease manager
		lm := NewLeaseManager(instance, 1*time.Second)

		// StoppedChan should not be closed initially
		select {
		case <-lm.StoppedChan():
			t.Fatal("StoppedChan should not be closed")
		default:
			// Expected
		}

		// Start and stop
		lm.Start(ctx)
		lm.Stop()

		// StoppedChan should be closed after stop
		select {
		case <-lm.StoppedChan():
			// Expected
		case <-time.After(1 * time.Second):
			t.Fatal("StoppedChan should be closed after stop")
		}
	})

	if err = instance.Unlock(ctx); err != nil {
		t.Fatal(err)
	}
	cleanupLock(name)
}

func TestLeaseManager_RenewalFailure(t *testing.T) {
	var (
		svc    = New()
		name   = "test-lease-fail-" + gtime.TimestampMilliStr()
		reason = "test reason"
		ctx    = context.Background()
	)

	cleanupLock(name)

	// Acquire lock
	instance, ok, err := svc.Lock(ctx, name, testHolder, reason, 30*time.Second)
	if err != nil || !ok {
		t.Fatal(err)
	}

	gtest.C(t, func(t *gtest.T) {
		// Create lease manager
		lm := NewLeaseManager(instance, 500*time.Millisecond)

		// Start lease renewal
		lm.Start(ctx)

		// Wait for lease to start
		time.Sleep(200 * time.Millisecond)

		// Simulate lock being taken by another node
		_, err = g.DB().Model("sys_locker").Data(g.Map{
			"expire_time": gtime.Now().Add(-1 * time.Second),
		}).Where("name", name).Update()
		t.AssertNil(err)

		// Wait for renewal to fail and lease manager to stop
		select {
		case <-lm.StoppedChan():
			// Expected - lease manager should stop on renewal failure
		case <-time.After(2 * time.Second):
			t.Fatal("LeaseManager should stop on renewal failure")
		}
	})

	cleanupLock(name)
}
