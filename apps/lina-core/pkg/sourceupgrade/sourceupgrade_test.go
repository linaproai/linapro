// This file verifies the published source-upgrade facade delegates to the
// host plugin upgrade governance service without altering results.

package sourceupgrade

import (
	"context"
	"testing"

	sourceupgradecontract "lina-core/pkg/sourceupgrade/contract"
)

// fakeSourceUpgradeGovernanceService captures calls made through the published facade.
type fakeSourceUpgradeGovernanceService struct {
	listErr          error
	listOutput       []*SourcePluginStatus
	readinessErr     error
	upgradeErr       error
	upgradeOutput    *SourcePluginUpgradeResult
	upgradedPluginID string
	validateCalled   bool
	listCalled       bool
	upgradeCalled    bool
}

// ListSourceUpgradeStatuses records the call and returns the configured list result.
func (f *fakeSourceUpgradeGovernanceService) ListSourceUpgradeStatuses(_ context.Context) ([]*sourceupgradecontract.SourcePluginStatus, error) {
	f.listCalled = true
	return f.listOutput, f.listErr
}

// UpgradeSourcePlugin records the target plugin ID and returns the configured upgrade result.
func (f *fakeSourceUpgradeGovernanceService) UpgradeSourcePlugin(_ context.Context, pluginID string) (*sourceupgradecontract.SourcePluginUpgradeResult, error) {
	f.upgradeCalled = true
	f.upgradedPluginID = pluginID
	return f.upgradeOutput, f.upgradeErr
}

// ValidateSourcePluginUpgradeReadiness records the readiness validation call.
func (f *fakeSourceUpgradeGovernanceService) ValidateSourcePluginUpgradeReadiness(_ context.Context) error {
	f.validateCalled = true
	return f.readinessErr
}

// TestNewRequiresPluginService verifies source-upgrade facade construction
// returns an error instead of creating an isolated host service graph.
func TestNewRequiresPluginService(t *testing.T) {
	if _, err := New(nil); err == nil {
		t.Fatal("expected New to return an error when plugin upgrade service is nil")
	}
}

// TestServiceImplDelegatesListSourcePluginStatuses verifies the published
// facade returns the host service list output unchanged.
func TestServiceImplDelegatesListSourcePluginStatuses(t *testing.T) {
	expected := []*SourcePluginStatus{{
		PluginID:          "plugin-demo-source",
		EffectiveVersion:  "v0.1.0",
		DiscoveredVersion: "v0.5.0",
	}}
	fakeService := &fakeSourceUpgradeGovernanceService{listOutput: expected}
	service, err := New(fakeService)
	if err != nil {
		t.Fatalf("expected source-upgrade facade construction to succeed, got error: %v", err)
	}

	items, err := service.ListSourcePluginStatuses(context.Background())
	if err != nil {
		t.Fatalf("expected list delegation to succeed, got error: %v", err)
	}
	if !fakeService.listCalled {
		t.Fatal("expected host service list delegation to be called")
	}
	if len(items) != 1 || items[0] != expected[0] {
		t.Fatalf("expected published facade to return delegated list output, got %#v", items)
	}
}

// TestServiceImplDelegatesUpgradeSourcePlugin verifies the published facade
// forwards the requested plugin ID and preserves the host result.
func TestServiceImplDelegatesUpgradeSourcePlugin(t *testing.T) {
	expected := &SourcePluginUpgradeResult{
		PluginID:    "plugin-demo-source",
		FromVersion: "v0.1.0",
		ToVersion:   "v0.5.0",
		Executed:    true,
	}
	fakeService := &fakeSourceUpgradeGovernanceService{upgradeOutput: expected}
	service, err := New(fakeService)
	if err != nil {
		t.Fatalf("expected source-upgrade facade construction to succeed, got error: %v", err)
	}

	result, err := service.UpgradeSourcePlugin(context.Background(), "plugin-demo-source")
	if err != nil {
		t.Fatalf("expected upgrade delegation to succeed, got error: %v", err)
	}
	if !fakeService.upgradeCalled {
		t.Fatal("expected host service upgrade delegation to be called")
	}
	if fakeService.upgradedPluginID != "plugin-demo-source" {
		t.Fatalf("expected upgrade delegation to forward plugin ID, got %q", fakeService.upgradedPluginID)
	}
	if result != expected {
		t.Fatalf("expected published facade to return delegated upgrade result, got %#v", result)
	}
}

// TestServiceImplDelegatesValidateSourcePluginUpgradeReadiness verifies the
// published facade delegates the non-blocking drift scan.
func TestServiceImplDelegatesValidateSourcePluginUpgradeReadiness(t *testing.T) {
	fakeService := &fakeSourceUpgradeGovernanceService{}
	service, err := New(fakeService)
	if err != nil {
		t.Fatalf("expected source-upgrade facade construction to succeed, got error: %v", err)
	}

	err = service.ValidateSourcePluginUpgradeReadiness(context.Background())
	if !fakeService.validateCalled {
		t.Fatal("expected host service readiness validation to be called")
	}
	if err != nil {
		t.Fatalf("expected published facade scan to succeed, got error: %v", err)
	}
}
