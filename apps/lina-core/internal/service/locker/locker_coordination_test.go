// This file verifies the coordination-backed locker implementation.

package locker

import (
	"context"
	"testing"
	"time"

	"lina-core/internal/service/coordination"
)

// TestCoordinationLockerLifecycle verifies locks use coordination ownership
// tokens for acquire, renew, held-state checks, and release.
func TestCoordinationLockerLifecycle(t *testing.T) {
	ctx := context.Background()
	ConfigureCoordination(coordination.NewMemory(nil))
	t.Cleanup(func() {
		ConfigureCoordination(nil)
	})

	svc := New()
	instance, ok, err := svc.Lock(ctx, "unit-coord-lock", "node-a", "unit", time.Second)
	if err != nil {
		t.Fatalf("acquire coordination lock: %v", err)
	}
	if !ok || instance == nil {
		t.Fatal("expected coordination lock acquisition")
	}
	if instance.ID() <= 0 || instance.Holder() == "node-a" || instance.Name() != "unit-coord-lock" {
		t.Fatalf("expected coordination instance metadata, got id=%d holder=%q name=%q", instance.ID(), instance.Holder(), instance.Name())
	}
	if held, err := instance.IsHeld(ctx); err != nil || !held {
		t.Fatalf("expected coordination lock held, held=%t err=%v", held, err)
	}
	if err = instance.Renew(ctx); err != nil {
		t.Fatalf("renew coordination lock through instance: %v", err)
	}
	if err = svc.RenewByName(ctx, instance.Name(), instance.Holder(), time.Second); err != nil {
		t.Fatalf("renew coordination lock by name: %v", err)
	}
	if err = svc.UnlockByName(ctx, instance.Name(), "wrong-token"); err != ErrLockNotHeld {
		t.Fatalf("expected wrong-token release to fail, got %v", err)
	}
	if err = svc.UnlockByName(ctx, instance.Name(), instance.Holder()); err != nil {
		t.Fatalf("release coordination lock by name: %v", err)
	}
	if held, err := instance.IsHeld(ctx); err != nil || held {
		t.Fatalf("expected coordination lock released, held=%t err=%v", held, err)
	}
}

// TestCoordinationLockerIsolatesNames verifies logical lock names remain
// independent even when held by the same owner.
func TestCoordinationLockerIsolatesNames(t *testing.T) {
	ctx := context.Background()
	ConfigureCoordination(coordination.NewMemory(nil))
	t.Cleanup(func() {
		ConfigureCoordination(nil)
	})

	svc := New()
	first, ok, err := svc.Lock(ctx, "plugin:a:sync", "node-a", "first", time.Second)
	if err != nil || !ok || first == nil {
		t.Fatalf("acquire first coordination lock, ok=%t err=%v", ok, err)
	}
	second, ok, err := svc.Lock(ctx, "plugin:b:sync", "node-a", "second", time.Second)
	if err != nil || !ok || second == nil {
		t.Fatalf("acquire second coordination lock, ok=%t err=%v", ok, err)
	}
	if first.Holder() == second.Holder() {
		t.Fatal("expected distinct coordination owner tokens for isolated locks")
	}
	if err = first.Unlock(ctx); err != nil {
		t.Fatalf("release first coordination lock: %v", err)
	}
	if held, err := second.IsHeld(ctx); err != nil || !held {
		t.Fatalf("expected second coordination lock to remain held, held=%t err=%v", held, err)
	}
}

// TestCoordinationLockerFailureReturnsError verifies coordination backend
// failures are surfaced instead of being treated as acquisition misses.
func TestCoordinationLockerFailureReturnsError(t *testing.T) {
	ctx := context.Background()
	coordSvc := coordination.NewMemory(nil)
	ConfigureCoordination(coordSvc)
	t.Cleanup(func() {
		ConfigureCoordination(nil)
	})
	if err := coordSvc.Close(ctx); err != nil {
		t.Fatalf("close coordination backend: %v", err)
	}

	if instance, ok, err := New().Lock(ctx, "unit-closed-lock", "node-a", "unit", time.Second); err == nil || ok || instance != nil {
		t.Fatalf("expected coordination lock failure, instance=%#v ok=%t err=%v", instance, ok, err)
	}
}
