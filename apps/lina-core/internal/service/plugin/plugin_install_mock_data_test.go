// This file covers the install-mock-data wrapping helpers in the plugin
// facade so a rolled-back mock load surfaces a stable user-facing bizerr
// regardless of whether the failure originated from the source-plugin path
// or the dynamic-plugin reconciler.

package plugin

import (
	"context"
	"errors"
	"testing"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/lifecycle"
)

// TestWrapMockDataLoadErrorPassesThroughNonMockErrors verifies that arbitrary
// install errors are returned unchanged so callers can distinguish mock
// rollback from other install failures.
func TestWrapMockDataLoadErrorPassesThroughNonMockErrors(t *testing.T) {
	if got := wrapMockDataLoadError(nil); got != nil {
		t.Fatalf("expected nil to pass through, got %v", got)
	}
	plain := errors.New("plain install failure")
	if got := wrapMockDataLoadError(plain); got != plain {
		t.Fatalf("expected non-mock error to pass through unchanged, got %v", got)
	}
	wrapped := gerror.Wrap(errors.New("inner"), "outer")
	if got := wrapMockDataLoadError(wrapped); got != wrapped {
		t.Fatalf("expected wrapped non-mock error to pass through unchanged, got %v", got)
	}
}

// TestWrapMockDataLoadErrorWrapsTypedError verifies that a *MockDataLoadError
// (or any error chain that contains one) is converted into the stable bizerr
// carrying pluginId / failedFile / rolledBackFiles / cause for i18n rendering.
func TestWrapMockDataLoadErrorWrapsTypedError(t *testing.T) {
	cause := errors.New("Duplicate entry 'admin' for key sys_user.idx_username")
	loadErr := &lifecycle.MockDataLoadError{
		PluginID:        "content-notice",
		FailedFile:      "002-content-notice-mock-data.sql",
		RolledBackFiles: []string{"001-content-notice-mock-data.sql", "002-content-notice-mock-data.sql"},
		Cause:           cause,
	}

	got := wrapMockDataLoadError(loadErr)
	if got == nil {
		t.Fatalf("expected wrapped error, got nil")
	}
	if got.Error() == "" {
		t.Fatalf("expected wrapped error message to be non-empty")
	}

	// Wrapping a context-wrapped chain should still recover the bizerr.
	wrappedChain := gerror.Wrap(loadErr, "facade context")
	if got := wrapMockDataLoadError(wrappedChain); got == nil {
		t.Fatalf("expected wrapping to traverse error chain via errors.As")
	}
}

// TestInstallMockDataContextHelpers verifies the catalog-shared context
// helpers used to thread the install-mock-data opt-in flag without changing
// the lifecycle/runtime/reconciler method signatures.
func TestInstallMockDataContextHelpers(t *testing.T) {
	if shouldInstallMockData(context.Background()) {
		t.Fatalf("expected default context to opt out of mock data")
	}
	if shouldInstallMockData(nil) {
		t.Fatalf("expected nil context to opt out of mock data")
	}

	enabledCtx := withInstallMockData(context.Background(), true)
	if !shouldInstallMockData(enabledCtx) {
		t.Fatalf("expected explicit opt-in to be observable")
	}
	if !catalog.ShouldInstallMockData(enabledCtx) {
		t.Fatalf("expected catalog-side helper to read the same value")
	}

	disabledCtx := withInstallMockData(context.Background(), false)
	if shouldInstallMockData(disabledCtx) {
		t.Fatalf("expected explicit opt-out to be observable")
	}
}
