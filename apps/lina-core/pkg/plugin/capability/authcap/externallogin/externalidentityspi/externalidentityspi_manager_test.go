// This file verifies the manager-backed external-identity Provider: lazy
// factory resolution gated by plugin enablement, fail-closed behavior without
// an enabled provider, and immediate disable/re-enable transitions.

package externalidentityspi

import (
	"context"
	"errors"
	"testing"
)

// stubProvider is a minimal Provider double resolving one fixed linkage.
type stubProvider struct {
	userID int
}

func (s stubProvider) Resolve(context.Context, ResolveInput) (int, bool, error) {
	return s.userID, true, nil
}
func (s stubProvider) Provision(context.Context, ProvisionInput) (int, error) {
	return s.userID, nil
}
func (s stubProvider) Bind(context.Context, BindInput) error     { return nil }
func (s stubProvider) Unbind(context.Context, UnbindInput) error { return nil }
func (s stubProvider) List(context.Context, int) ([]BoundIdentity, error) {
	return []BoundIdentity{{Provider: "stub"}}, nil
}

// toggleEnablement reports plugin enablement from a mutable flag.
type toggleEnablement struct {
	enabled bool
}

func (t *toggleEnablement) IsProviderEnabled(context.Context, string) bool {
	return t.enabled
}

// TestManagedProviderFailsClosedWithoutEnabledProvider verifies every seam
// method is fail-closed when no provider plugin is enabled.
func TestManagedProviderFailsClosedWithoutEnabledProvider(t *testing.T) {
	ctx := context.Background()
	manager := NewManager()
	if err := manager.RegisterFactory("test-plugin", func(context.Context, ProviderEnv) (Provider, error) {
		return stubProvider{userID: 7}, nil
	}); err != nil {
		t.Fatalf("register factory: %v", err)
	}
	provider := New(manager, &toggleEnablement{enabled: false}, nil)

	if _, found, err := provider.Resolve(ctx, ResolveInput{Provider: "google", Subject: "s"}); err != nil || found {
		t.Fatalf("disabled resolve: expected found=false without error, got found=%v err=%v", found, err)
	}
	if _, err := provider.Provision(ctx, ProvisionInput{Provider: "google", Subject: "s"}); !errors.Is(err, ErrProviderUnavailable) {
		t.Fatalf("disabled provision: expected provider-unavailable, got %v", err)
	}
	if err := provider.Bind(ctx, BindInput{UserID: 1, Provider: "google", Subject: "s"}); !errors.Is(err, ErrProviderUnavailable) {
		t.Fatalf("disabled bind: expected provider-unavailable, got %v", err)
	}
	if err := provider.Unbind(ctx, UnbindInput{UserID: 1, Provider: "google", Subject: "s"}); !errors.Is(err, ErrProviderUnavailable) {
		t.Fatalf("disabled unbind: expected provider-unavailable, got %v", err)
	}
	if identities, err := provider.List(ctx, 1); err != nil || len(identities) != 0 {
		t.Fatalf("disabled list: expected empty without error, got %v err=%v", identities, err)
	}
}

// TestManagedProviderFollowsEnablementTransitions verifies disabling the
// provider plugin immediately fails closed and re-enabling restores delegation.
func TestManagedProviderFollowsEnablementTransitions(t *testing.T) {
	ctx := context.Background()
	manager := NewManager()
	if err := manager.RegisterFactory("test-plugin", func(context.Context, ProviderEnv) (Provider, error) {
		return stubProvider{userID: 42}, nil
	}); err != nil {
		t.Fatalf("register factory: %v", err)
	}
	enablement := &toggleEnablement{enabled: true}
	provider := New(manager, enablement, nil)

	userID, found, err := provider.Resolve(ctx, ResolveInput{Provider: "google", Subject: "s"})
	if err != nil || !found || userID != 42 {
		t.Fatalf("enabled resolve: expected user 42, got id=%d found=%v err=%v", userID, found, err)
	}

	enablement.enabled = false
	if _, found, err = provider.Resolve(ctx, ResolveInput{Provider: "google", Subject: "s"}); err != nil || found {
		t.Fatalf("after disable: expected fail-closed, got found=%v err=%v", found, err)
	}

	enablement.enabled = true
	userID, found, err = provider.Resolve(ctx, ResolveInput{Provider: "google", Subject: "s"})
	if err != nil || !found || userID != 42 {
		t.Fatalf("after re-enable: expected user 42 restored, got id=%d found=%v err=%v", userID, found, err)
	}
}

// TestManagedProviderNilManagerFailsClosed verifies the zero-configuration
// provider (nil manager/enablement) never resolves or provisions.
func TestManagedProviderNilManagerFailsClosed(t *testing.T) {
	ctx := context.Background()
	provider := New(nil, nil, nil)
	if _, found, err := provider.Resolve(ctx, ResolveInput{Provider: "google", Subject: "s"}); err != nil || found {
		t.Fatalf("nil manager resolve: expected fail-closed, got found=%v err=%v", found, err)
	}
	if _, err := provider.Provision(ctx, ProvisionInput{Provider: "google", Subject: "s"}); !errors.Is(err, ErrProviderUnavailable) {
		t.Fatalf("nil manager provision: expected provider-unavailable, got %v", err)
	}
}
