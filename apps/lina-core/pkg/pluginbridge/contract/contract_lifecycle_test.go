// This file tests dynamic lifecycle bridge contracts.

package contract

import (
	"encoding/json"
	"testing"
)

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
			Operation:    LifecycleOperationUpgrade,
			RequestType:  "DynamicUpgradeReq",
			InternalPath: "/__lifecycle/upgrade",
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
		{
			Operation:    LifecycleOperationUninstall,
			RequestType:  "DynamicUninstallReq",
			InternalPath: "/__lifecycle/uninstall",
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

// TestLifecycleRequestUsesTypedManifestSnapshot verifies manifest snapshots are
// published as the shared typed bridge contract, not unstructured value maps.
func TestLifecycleRequestUsesTypedManifestSnapshot(t *testing.T) {
	t.Parallel()

	content, err := json.Marshal(&LifecycleRequest{
		PluginID:  "plugin-dynamic-lifecycle",
		Operation: LifecycleOperationBeforeUpgrade.String(),
		FromManifest: &ManifestSnapshotV1{
			ID:                      "plugin-dynamic-lifecycle",
			Version:                 "v0.1.0",
			Type:                    "dynamic",
			HostServiceAuthRequired: false,
		},
		ToManifest: &ManifestSnapshotV1{
			ID:                      "plugin-dynamic-lifecycle",
			Version:                 "v0.2.0",
			Type:                    "dynamic",
			ScopeNature:             "tenant_aware",
			SupportsMultiTenant:     true,
			DefaultInstallMode:      "tenant_scoped",
			InstallSQLCount:         2,
			ResourceSpecCount:       3,
			HostServiceAuthRequired: true,
		},
	})
	if err != nil {
		t.Fatalf("expected lifecycle request to marshal, got %v", err)
	}

	decoded := &LifecycleRequest{}
	if err = json.Unmarshal(content, decoded); err != nil {
		t.Fatalf("expected lifecycle request to unmarshal, got %v", err)
	}
	if decoded.ToManifest == nil ||
		decoded.ToManifest.ScopeNature != "tenant_aware" ||
		decoded.ToManifest.DefaultInstallMode != "tenant_scoped" ||
		decoded.ToManifest.InstallSQLCount != 2 ||
		decoded.ToManifest.ResourceSpecCount != 3 ||
		decoded.ToManifest.HostServiceAuthRequired != true {
		t.Fatalf("unexpected typed manifest snapshot: %#v", decoded.ToManifest)
	}
}
