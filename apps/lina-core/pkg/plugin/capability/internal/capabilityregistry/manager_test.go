// This file verifies provider activation and status diagnostics in the shared
// capability registry used by framework capability adapters.

package capabilityregistry

import (
	"context"
	"errors"
	"testing"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/capmodel"
)

func TestStatusWithProviderReportsFactoryErrorAsUnavailable(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	manager := NewManager[string]()
	if err := manager.RegisterFactory("cap.test", "provider-a", func(context.Context, string) (any, error) {
		return nil, errors.New("missing dependency")
	}); err != nil {
		t.Fatalf("register provider: %v", err)
	}

	status := manager.StatusWithProvider(ctx, "cap.test", testEnablement{"provider-a": true}, nil)
	if status.Available {
		t.Fatalf("expected provider construction failure to mark capability unavailable, got %#v", status)
	}
	if status.ActiveProvider != "" {
		t.Fatalf("expected active provider to be cleared, got %q", status.ActiveProvider)
	}
	if status.Reason != ReasonProviderError {
		t.Fatalf("expected provider error reason, got %q", status.Reason)
	}
	if len(status.Providers) != 1 || status.Providers[0].Active || status.Providers[0].Reason != ReasonProviderError {
		t.Fatalf("expected inactive provider_error provider status, got %#v", status.Providers)
	}
}

func TestActiveProviderWithErrorReturnsFactoryError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	manager := NewManager[string]()
	if err := manager.RegisterFactory("cap.test", "provider-a", func(context.Context, string) (any, error) {
		return nil, errors.New("missing dependency")
	}); err != nil {
		t.Fatalf("register provider: %v", err)
	}

	provider, err := manager.ActiveProviderWithError(ctx, "cap.test", testEnablement{"provider-a": true}, nil)
	if err == nil {
		t.Fatal("expected factory error")
	}
	if provider != nil {
		t.Fatalf("expected no provider on factory error, got %#v", provider)
	}
}

func TestActiveProviderWithErrorRejectsMultipleEnabledProviders(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	manager := NewManager[string]()
	for _, pluginID := range []string{"provider-a", "provider-b"} {
		pluginID := pluginID
		if err := manager.RegisterFactory("cap.test", pluginID, func(context.Context, string) (any, error) {
			return pluginID, nil
		}); err != nil {
			t.Fatalf("register provider %s: %v", pluginID, err)
		}
	}

	provider, err := manager.ActiveProviderWithError(
		ctx,
		"cap.test",
		testEnablement{"provider-a": true, "provider-b": true},
		nil,
	)
	if err == nil {
		t.Fatal("expected conflict error")
	}
	if !bizerr.Is(err, capmodel.CodeCapabilityProviderConflict) {
		t.Fatalf("expected capability provider conflict error, got %v", err)
	}
	bizErr, ok := bizerr.As(err)
	if !ok {
		t.Fatalf("expected structured conflict error, got %v", err)
	}
	if bizErr.Params()["capability"] != "cap.test" {
		t.Fatalf("expected capability param cap.test, got %#v", bizErr.Params())
	}
	if bizErr.Params()["providerIds"] != "provider-a,provider-b" {
		t.Fatalf("expected deterministic provider IDs, got %#v", bizErr.Params())
	}
	if provider != nil {
		t.Fatalf("expected no provider on conflict, got %#v", provider)
	}

	status := manager.StatusWithProvider(
		ctx,
		"cap.test",
		testEnablement{"provider-a": true, "provider-b": true},
		nil,
	)
	if status.Available || status.ActiveProvider != "" || status.Reason != ReasonConflict {
		t.Fatalf("expected unavailable conflict status, got %#v", status)
	}
	for _, providerStatus := range status.Providers {
		if !providerStatus.Conflict || providerStatus.Reason != ReasonConflict {
			t.Fatalf("expected conflicted provider status, got %#v", status.Providers)
		}
	}
}

func TestStatusWithProviderConstructsAndCachesActiveProvider(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	manager := NewManager[string]()
	calls := 0
	if err := manager.RegisterFactory("cap.test", "provider-a", func(_ context.Context, env string) (any, error) {
		calls++
		return "provider:" + env, nil
	}); err != nil {
		t.Fatalf("register provider: %v", err)
	}

	envFactory := func(context.Context, string) string {
		return "env-a"
	}
	status := manager.StatusWithProvider(ctx, "cap.test", testEnablement{"provider-a": true}, envFactory)
	if !status.Available || status.ActiveProvider != "provider-a" || status.Reason != "" {
		t.Fatalf("expected available active provider status, got %#v", status)
	}
	provider, err := manager.ActiveProviderWithError(ctx, "cap.test", testEnablement{"provider-a": true}, envFactory)
	if err != nil {
		t.Fatalf("active provider returned error: %v", err)
	}
	if provider != "provider:env-a" {
		t.Fatalf("unexpected provider: %#v", provider)
	}
	if calls != 1 {
		t.Fatalf("expected provider to be cached after status validation, got calls=%d", calls)
	}
}

type testEnablement map[string]bool

func (s testEnablement) IsProviderEnabled(_ context.Context, pluginID string) bool {
	return s[pluginID]
}
