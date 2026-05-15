// This file tests plugin-facing distributed lock adapter logic, ticket
// encoding and validation, and bizerr code completeness.

package hostlock

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/internal/service/locker"
	"lina-core/pkg/bizerr"
)

// fakeLocker stubs the locker.Service calls recorded by hostlock Acquire,
// Renew, and Release tests.
type fakeLocker struct {
	lockFn         func(ctx context.Context, name, holder, reason string, lease time.Duration) (*locker.Instance, bool, error)
	lockFuncFn     func(ctx context.Context, name, holder, reason string, lease time.Duration, f func() error) (bool, error)
	unlockFn       func(ctx context.Context, lockID int64, holder string) error
	renewFn        func(ctx context.Context, lockID int64, holder string, lease time.Duration) error
	unlockByNameFn func(ctx context.Context, name string, holder string) error
	renewByNameFn  func(ctx context.Context, name string, holder string, lease time.Duration) error
}

func (f *fakeLocker) Lock(ctx context.Context, name, holder, reason string, lease time.Duration) (*locker.Instance, bool, error) {
	return f.lockFn(ctx, name, holder, reason, lease)
}

func (f *fakeLocker) LockFunc(ctx context.Context, name, holder, reason string, lease time.Duration, fn func() error) (bool, error) {
	if f.lockFuncFn != nil {
		return f.lockFuncFn(ctx, name, holder, reason, lease, fn)
	}
	return false, nil
}

func (f *fakeLocker) Unlock(ctx context.Context, lockID int64, holder string) error {
	return f.unlockFn(ctx, lockID, holder)
}

func (f *fakeLocker) Renew(ctx context.Context, lockID int64, holder string, lease time.Duration) error {
	return f.renewFn(ctx, lockID, holder, lease)
}

func (f *fakeLocker) UnlockByName(ctx context.Context, name string, holder string) error {
	return f.unlockByNameFn(ctx, name, holder)
}

func (f *fakeLocker) RenewByName(ctx context.Context, name string, holder string, lease time.Duration) error {
	return f.renewByNameFn(ctx, name, holder, lease)
}

// newHostlockService creates a serviceImpl backed by a fake locker.
func newHostlockService(l *fakeLocker) *serviceImpl {
	return &serviceImpl{lockerSvc: l}
}

// TestBuildActualLockName verifies lock name construction and validation.
func TestBuildActualLockName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		pluginID    string
		tenantID    int64
		resourceRef string
		wantErr     *bizerr.Code
		wantPrefix  string
	}{
		{
			"normal build",
			"my-plugin", 0, "res-1",
			nil,
			"plugin:my-plugin:tenant=0:res-1",
		},
		{
			"empty pluginID",
			"", 0, "res-1",
			CodeHostLockPluginIDRequired,
			"",
		},
		{
			"whitespace-only pluginID",
			"   ", 0, "res-1",
			CodeHostLockPluginIDRequired,
			"",
		},
		{
			"empty resourceRef",
			"p", 0, "",
			CodeHostLockResourceRequired,
			"",
		},
		{
			"name at max length allowed",
			"p", 0, strings.Repeat("x", maxLockBytes-len("plugin:p:tenant=0:")),
			nil,
			"",
		},
		{
			"name exceeds max length",
			"p", 0, strings.Repeat("y", maxLockBytes),
			CodeHostLockNameTooLong,
			"",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := buildActualLockName(tc.pluginID, tc.tenantID, tc.resourceRef)
			if tc.wantErr != nil {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !bizerr.Is(err, tc.wantErr) {
					t.Fatalf("expected %s, got %v", tc.wantErr.RuntimeCode(), err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantPrefix != "" && got != tc.wantPrefix {
				t.Fatalf("expected %q, got %q", tc.wantPrefix, got)
			}
		})
	}
}

// TestNormalizeLease verifies lease duration validation and defaulting.
func TestNormalizeLease(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		leaseMillis int64
		wantLease   time.Duration
		wantErr     *bizerr.Code
	}{
		{"zero uses default", 0, defaultLease, nil},
		{"negative uses default", -1, defaultLease, nil},
		{"below min rejected", 500, 0, CodeHostLockLeaseTooShort},
		{"at min accepted", 1000, minLease, nil},
		{"above max rejected", 301000, 0, CodeHostLockLeaseTooLong},
		{"normal accepted", 60000, 60 * time.Second, nil},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := normalizeLease(tc.leaseMillis)
			if tc.wantErr != nil {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !bizerr.Is(err, tc.wantErr) {
					t.Fatalf("expected %s, got %v", tc.wantErr.RuntimeCode(), err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.wantLease {
				t.Fatalf("expected lease %v, got %v", tc.wantLease, got)
			}
		})
	}
}

// TestBuildLockReason verifies audit reason string construction.
func TestBuildLockReason(t *testing.T) {
	t.Parallel()

	t.Run("without requestID", func(t *testing.T) {
		t.Parallel()
		got := buildLockReason("my-resource", "")
		if !strings.Contains(got, "my-resource") {
			t.Fatalf("expected reason to contain resource ref, got %q", got)
		}
		if strings.Contains(got, "request=") {
			t.Fatalf("expected reason without request=, got %q", got)
		}
	})

	t.Run("with requestID", func(t *testing.T) {
		t.Parallel()
		got := buildLockReason("my-resource", "req-abc")
		if !strings.Contains(got, "request=req-abc") {
			t.Fatalf("expected reason to contain request=req-abc, got %q", got)
		}
	})
}

// TestTenantIDFromIdentity verifies tenant identity normalization.
func TestTenantIDFromIdentity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  int32
		expect int64
	}{
		{"negative maps to platform", -1, 0},
		{"zero maps to platform", 0, 0},
		{"positive keeps value", 5, 5},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := TenantIDFromIdentity(tc.input); got != tc.expect {
				t.Fatalf("expected %d, got %d", tc.expect, got)
			}
		})
	}
}

// TestEncodeDecodeTicketRoundTrip verifies a valid ticket survives a full
// encode-decode cycle.
func TestEncodeDecodeTicketRoundTrip(t *testing.T) {
	t.Parallel()

	claims := lockTicketClaims{
		LockID:      42,
		LockName:    "plugin:p:tenant=0:r",
		TenantID:    0,
		PluginID:    "p",
		ResourceRef: "r",
		Holder:      "pl:abc",
		LeaseMillis: 30000,
	}

	ticket, err := encodeLockTicket(claims)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	if ticket == "" {
		t.Fatal("expected non-empty ticket")
	}

	decoded, err := decodeAndValidateTicket(ticket, "p", 0, "r")
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if decoded.LockID != claims.LockID {
		t.Fatalf("expected LockID %d, got %d", claims.LockID, decoded.LockID)
	}
	if decoded.Holder != claims.Holder {
		t.Fatalf("expected Holder %q, got %q", claims.Holder, decoded.Holder)
	}
	if decoded.LeaseMillis != claims.LeaseMillis {
		t.Fatalf("expected LeaseMillis %d, got %d", claims.LeaseMillis, decoded.LeaseMillis)
	}
}

// TestDecodeAndValidateTicketErrors verifies every ticket validation failure
// produces the expected bizerr code.
func TestDecodeAndValidateTicketErrors(t *testing.T) {
	t.Parallel()

	t.Run("empty ticket", func(t *testing.T) {
		t.Parallel()
		_, err := decodeAndValidateTicket("", "p", 0, "r")
		if !bizerr.Is(err, CodeHostLockTicketRequired) {
			t.Fatalf("expected CodeHostLockTicketRequired, got %v", err)
		}
	})

	t.Run("bad base64", func(t *testing.T) {
		t.Parallel()
		_, err := decodeAndValidateTicket("!!!not-base64", "p", 0, "r")
		if !bizerr.Is(err, CodeHostLockTicketParseFailed) {
			t.Fatalf("expected CodeHostLockTicketParseFailed, got %v", err)
		}
	})

	t.Run("valid base64 not json", func(t *testing.T) {
		t.Parallel()
		// RawURLEncoding of "not-json", no padding
		ticket := "bm90LWpzb24"
		_, err := decodeAndValidateTicket(ticket, "p", 0, "r")
		if !bizerr.Is(err, CodeHostLockTicketUnmarshalFailed) {
			t.Fatalf("expected CodeHostLockTicketUnmarshalFailed, got %v", err)
		}
	})

	t.Run("missing holder", func(t *testing.T) {
		t.Parallel()
		claims := lockTicketClaims{LockID: 1, Holder: "", LeaseMillis: 30000, PluginID: "p", ResourceRef: "r"}
		ticket, err := encodeLockTicket(claims)
		if err != nil {
			t.Fatalf("encode failed: %v", err)
		}
		_, err = decodeAndValidateTicket(ticket, "p", 0, "r")
		if !bizerr.Is(err, CodeHostLockTicketInvalid) {
			t.Fatalf("expected CodeHostLockTicketInvalid, got %v", err)
		}
	})

	t.Run("plugin mismatch", func(t *testing.T) {
		t.Parallel()
		claims := lockTicketClaims{LockID: 1, Holder: "pl:abc", LeaseMillis: 30000, PluginID: "p", ResourceRef: "r"}
		ticket, err := encodeLockTicket(claims)
		if err != nil {
			t.Fatalf("encode failed: %v", err)
		}
		_, err = decodeAndValidateTicket(ticket, "other-plugin", 0, "r")
		if !bizerr.Is(err, CodeHostLockTicketPluginMismatch) {
			t.Fatalf("expected CodeHostLockTicketPluginMismatch, got %v", err)
		}
	})

	t.Run("tenant mismatch", func(t *testing.T) {
		t.Parallel()
		claims := lockTicketClaims{LockID: 1, Holder: "pl:abc", LeaseMillis: 30000, PluginID: "p", ResourceRef: "r", TenantID: 1}
		ticket, err := encodeLockTicket(claims)
		if err != nil {
			t.Fatalf("encode failed: %v", err)
		}
		_, err = decodeAndValidateTicket(ticket, "p", 2, "r")
		if !bizerr.Is(err, CodeHostLockTicketTenantMismatch) {
			t.Fatalf("expected CodeHostLockTicketTenantMismatch, got %v", err)
		}
	})

	t.Run("resource mismatch", func(t *testing.T) {
		t.Parallel()
		claims := lockTicketClaims{LockID: 1, Holder: "pl:abc", LeaseMillis: 30000, PluginID: "p", ResourceRef: "r1"}
		ticket, err := encodeLockTicket(claims)
		if err != nil {
			t.Fatalf("encode failed: %v", err)
		}
		_, err = decodeAndValidateTicket(ticket, "p", 0, "r2")
		if !bizerr.Is(err, CodeHostLockTicketResourceMismatch) {
			t.Fatalf("expected CodeHostLockTicketResourceMismatch, got %v", err)
		}
	})
}

// TestAcquire verifies the Acquire method delegates to locker.Service correctly.
func TestAcquire(t *testing.T) {
	t.Parallel()

	t.Run("successful acquire", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		locker := &fakeLocker{
			lockFn: func(_ context.Context, name, holder, reason string, lease time.Duration) (*locker.Instance, bool, error) {
				return &locker.Instance{}, true, nil
			},
		}
		svc := newHostlockService(locker)

		out, err := svc.Acquire(ctx, AcquireInput{
			PluginID:    "p",
			TenantID:    0,
			ResourceRef: "r",
			LeaseMillis: 30000,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !out.Acquired {
			t.Fatal("expected lock to be acquired")
		}
		if out.Ticket == "" {
			t.Fatal("expected non-empty ticket")
		}
		if out.ExpireAt == nil {
			t.Fatal("expected ExpireAt")
		}
	})

	t.Run("lock held by another", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		locker := &fakeLocker{
			lockFn: func(_ context.Context, name, holder, reason string, lease time.Duration) (*locker.Instance, bool, error) {
				return nil, false, nil
			},
		}
		svc := newHostlockService(locker)

		out, err := svc.Acquire(ctx, AcquireInput{
			PluginID:    "p",
			TenantID:    0,
			ResourceRef: "r",
			LeaseMillis: 30000,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out.Acquired {
			t.Fatal("expected lock not to be acquired")
		}
		if out.Ticket != "" {
			t.Fatalf("expected empty ticket, got %q", out.Ticket)
		}
	})

	t.Run("locker internal error", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		locker := &fakeLocker{
			lockFn: func(_ context.Context, name, holder, reason string, lease time.Duration) (*locker.Instance, bool, error) {
				return nil, false, bizerr.NewCode(CodeHostLockTicketInvalid)
			},
		}
		svc := newHostlockService(locker)

		_, err := svc.Acquire(ctx, AcquireInput{
			PluginID:    "p",
			TenantID:    0,
			ResourceRef: "r",
			LeaseMillis: 30000,
		})
		if err == nil {
			t.Fatal("expected error from locker")
		}
	})

	t.Run("empty pluginID rejected before locker call", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		lockerCalled := false
		locker := &fakeLocker{
			lockFn: func(_ context.Context, name, holder, reason string, lease time.Duration) (*locker.Instance, bool, error) {
				lockerCalled = true
				return nil, false, nil
			},
		}
		svc := newHostlockService(locker)

		_, err := svc.Acquire(ctx, AcquireInput{
			PluginID:    "",
			TenantID:    0,
			ResourceRef: "r",
			LeaseMillis: 30000,
		})
		if err == nil {
			t.Fatal("expected validation error")
		}
		if lockerCalled {
			t.Fatal("expected locker not to be called before input validation")
		}
	})
}

// TestRenew verifies Renew dispatches to the correct locker method.
func TestRenew(t *testing.T) {
	t.Parallel()

	t.Run("renews by name when lock name present", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		renewedByName := false
		locker := &fakeLocker{
			renewByNameFn: func(_ context.Context, name, holder string, lease time.Duration) error {
				renewedByName = true
				return nil
			},
		}
		svc := newHostlockService(locker)

		claims := lockTicketClaims{
			LockID: 1, LockName: "plugin:p:tenant=0:r", Holder: "pl:abc",
			LeaseMillis: 30000, PluginID: "p", ResourceRef: "r",
		}
		ticket, err := encodeLockTicket(claims)
		if err != nil {
			t.Fatalf("encode failed: %v", err)
		}

		_, err = svc.Renew(ctx, "p", 0, "r", ticket)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !renewedByName {
			t.Fatal("expected RenewByName to be called")
		}
	})

	t.Run("renews by ID when lock name empty", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		renewedByID := false
		locker := &fakeLocker{
			renewFn: func(_ context.Context, lockID int64, holder string, lease time.Duration) error {
				renewedByID = true
				return nil
			},
		}
		svc := newHostlockService(locker)

		claims := lockTicketClaims{
			LockID: 1, LockName: "", Holder: "pl:abc",
			LeaseMillis: 30000, PluginID: "p", ResourceRef: "r",
		}
		ticket, err := encodeLockTicket(claims)
		if err != nil {
			t.Fatalf("encode failed: %v", err)
		}

		_, err = svc.Renew(ctx, "p", 0, "r", ticket)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !renewedByID {
			t.Fatal("expected Renew to be called")
		}
	})

	t.Run("invalid ticket rejected", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		svc := newHostlockService(&fakeLocker{})

		_, err := svc.Renew(ctx, "p", 0, "r", "")
		if !bizerr.Is(err, CodeHostLockTicketRequired) {
			t.Fatalf("expected CodeHostLockTicketRequired, got %v", err)
		}
	})
}

// TestRelease verifies Release dispatches to the correct locker method.
func TestRelease(t *testing.T) {
	t.Parallel()

	t.Run("releases by name when lock name present", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		unlockedByName := false
		locker := &fakeLocker{
			unlockByNameFn: func(_ context.Context, name, holder string) error {
				unlockedByName = true
				return nil
			},
		}
		svc := newHostlockService(locker)

		claims := lockTicketClaims{
			LockID: 1, LockName: "plugin:p:tenant=0:r", Holder: "pl:abc",
			LeaseMillis: 30000, PluginID: "p", ResourceRef: "r",
		}
		ticket, err := encodeLockTicket(claims)
		if err != nil {
			t.Fatalf("encode failed: %v", err)
		}

		err = svc.Release(ctx, "p", 0, "r", ticket)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !unlockedByName {
			t.Fatal("expected UnlockByName to be called")
		}
	})

	t.Run("releases by ID when lock name empty", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		unlockedByID := false
		locker := &fakeLocker{
			unlockFn: func(_ context.Context, lockID int64, holder string) error {
				unlockedByID = true
				return nil
			},
		}
		svc := newHostlockService(locker)

		claims := lockTicketClaims{
			LockID: 1, LockName: "", Holder: "pl:abc",
			LeaseMillis: 30000, PluginID: "p", ResourceRef: "r",
		}
		ticket, err := encodeLockTicket(claims)
		if err != nil {
			t.Fatalf("encode failed: %v", err)
		}

		err = svc.Release(ctx, "p", 0, "r", ticket)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !unlockedByID {
			t.Fatal("expected Unlock to be called")
		}
	})

	t.Run("invalid ticket rejected", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		svc := newHostlockService(&fakeLocker{})

		err := svc.Release(ctx, "p", 0, "r", "")
		if !bizerr.Is(err, CodeHostLockTicketRequired) {
			t.Fatalf("expected CodeHostLockTicketRequired, got %v", err)
		}
	})
}

// TestCodeDefinitions verifies every hostlock bizerr.Code is properly defined.
func TestCodeDefinitions(t *testing.T) {
	t.Parallel()

	codes := []struct {
		name string
		code *bizerr.Code
	}{
		{"CodeHostLockPluginIDRequired", CodeHostLockPluginIDRequired},
		{"CodeHostLockResourceRequired", CodeHostLockResourceRequired},
		{"CodeHostLockNameTooLong", CodeHostLockNameTooLong},
		{"CodeHostLockLeaseTooShort", CodeHostLockLeaseTooShort},
		{"CodeHostLockLeaseTooLong", CodeHostLockLeaseTooLong},
		{"CodeHostLockTicketMarshalFailed", CodeHostLockTicketMarshalFailed},
		{"CodeHostLockTicketRequired", CodeHostLockTicketRequired},
		{"CodeHostLockTicketParseFailed", CodeHostLockTicketParseFailed},
		{"CodeHostLockTicketUnmarshalFailed", CodeHostLockTicketUnmarshalFailed},
		{"CodeHostLockTicketInvalid", CodeHostLockTicketInvalid},
		{"CodeHostLockTicketPluginMismatch", CodeHostLockTicketPluginMismatch},
		{"CodeHostLockTicketTenantMismatch", CodeHostLockTicketTenantMismatch},
		{"CodeHostLockTicketResourceMismatch", CodeHostLockTicketResourceMismatch},
	}

	for _, tc := range codes {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if tc.code == nil {
				t.Fatal("expected code to be non-nil")
			}
			if tc.code.RuntimeCode() == "" {
				t.Fatalf("expected non-empty runtime code for %s", tc.name)
			}
			if tc.code.Fallback() == "" {
				t.Fatalf("expected non-empty fallback for %s", tc.name)
			}
			if tc.code.TypeCode() == gcode.CodeUnknown {
				t.Fatalf("expected valid GoFrame type code for %s, got CodeUnknown", tc.name)
			}
		})
	}
}

// TestNewPanicsOnNilLocker verifies the constructor enforces the locker dependency.
func TestNewPanicsOnNilLocker(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when lockerSvc is nil")
		}
	}()
	New(nil)
}

// TestNewReturnsService verifies the constructor returns a non-nil service.
func TestNewReturnsService(t *testing.T) {
	t.Parallel()

	svc := New(&fakeLocker{})
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
}
