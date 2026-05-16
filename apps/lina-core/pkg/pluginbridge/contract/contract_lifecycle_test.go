// This file tests dynamic lifecycle bridge contracts.

package contract

import "testing"

// TestValidateLifecycleContractsAcceptsSourceNamedHooks verifies dynamic
// lifecycle operations use the same Before* and After* names as source plugins.
func TestValidateLifecycleContractsAcceptsSourceNamedHooks(t *testing.T) {
	t.Parallel()

	items := []*LifecycleContract{
		{
			Operation:    LifecycleOperationBeforeInstall,
			RequestType:  "DynamicBeforeInstallReq",
			InternalPath: "__lifecycle/before-install/",
			TimeoutMs:    50,
		},
		{
			Operation:    LifecycleOperationBeforeUpgrade,
			RequestType:  "DynamicBeforeUpgradeReq",
			InternalPath: "/__lifecycle/before-upgrade",
		},
		{
			Operation:    LifecycleOperationAfterInstall,
			RequestType:  "DynamicAfterInstallReq",
			InternalPath: "/__lifecycle/after-install",
		},
		{
			Operation:    LifecycleOperationAfterUpgrade,
			RequestType:  "DynamicAfterUpgradeReq",
			InternalPath: "/__lifecycle/after-upgrade",
		},
	}

	if err := ValidateLifecycleContracts("plugin-dynamic-lifecycle", items); err != nil {
		t.Fatalf("expected lifecycle contracts to validate, got %v", err)
	}
	if items[0].InternalPath != "/__lifecycle/before-install" {
		t.Fatalf("expected internal path to normalize, got %s", items[0].InternalPath)
	}
}

// TestValidateLifecycleContractsRejectsParallelNaming verifies old guard-style
// aliases cannot be used for dynamic lifecycle declarations.
func TestValidateLifecycleContractsRejectsParallelNaming(t *testing.T) {
	t.Parallel()

	err := ValidateLifecycleContracts("plugin-dynamic-lifecycle", []*LifecycleContract{
		{
			Operation:    LifecycleOperation("CanInstall"),
			RequestType:  "DynamicCanInstallReq",
			InternalPath: "/__lifecycle/can-install",
		},
	})
	if err == nil {
		t.Fatal("expected unsupported lifecycle operation to be rejected")
	}
}
