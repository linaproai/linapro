// This file covers lifecycle-owned install helper behavior.

package lifecycle

import (
	pluginv1 "lina-core/api/plugin/v1"
	"testing"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/store"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// TestApplyInstallModeSelectionRejectsInvalidMode verifies service-layer install
// validation rejects unsupported install-mode values before registry sync.
func TestApplyInstallModeSelectionRejectsInvalidMode(t *testing.T) {
	manifest := &catalog.Manifest{
		ID:                 "plugin-invalid-install-mode",
		ScopeNature:        pluginv1.ScopeNatureTenantAware.String(),
		DefaultInstallMode: pluginv1.InstallModeTenantScoped.String(),
	}

	err := applyInstallModeSelection(manifest, "per_tenant")
	if !bizerr.Is(err, CodePluginInstallModeInvalid) {
		t.Fatalf("expected invalid install mode bizerr, got %v", err)
	}
}

// TestApplyInstallModeSelectionRejectsPlatformOnlyTenantScoped verifies
// platform-only plugins cannot be installed with tenant-scoped enablement.
func TestApplyInstallModeSelectionRejectsPlatformOnlyTenantScoped(t *testing.T) {
	manifest := &catalog.Manifest{
		ID:                 "plugin-platform-only-install-mode",
		ScopeNature:        pluginv1.ScopeNaturePlatformOnly.String(),
		DefaultInstallMode: pluginv1.InstallModeGlobal.String(),
	}

	err := applyInstallModeSelection(manifest, pluginv1.InstallModeTenantScoped.String())
	if !bizerr.Is(err, CodePluginInstallModeInvalidForScopeNature) {
		t.Fatalf("expected scope/install-mode mismatch bizerr, got %v", err)
	}
}

// TestApplyInstallModeSelectionPersistsExplicitTenantAwareMode verifies an
// explicit platform selection overrides the manifest default before install.
func TestApplyInstallModeSelectionPersistsExplicitTenantAwareMode(t *testing.T) {
	manifest := &catalog.Manifest{
		ID:                 "plugin-tenant-aware-install-mode",
		ScopeNature:        pluginv1.ScopeNatureTenantAware.String(),
		DefaultInstallMode: pluginv1.InstallModeTenantScoped.String(),
	}

	if err := applyInstallModeSelection(manifest, pluginv1.InstallModeGlobal.String()); err != nil {
		t.Fatalf("expected explicit global install mode to be accepted, got %v", err)
	}
	if manifest.DefaultInstallMode != pluginv1.InstallModeGlobal.String() {
		t.Fatalf("expected explicit global install mode to be applied, got %s", manifest.DefaultInstallMode)
	}
}

// TestApplyInstallModeSelectionRejectsUnsupportedTenantScoped verifies explicit
// manifest opt-out from tenant governance also rejects tenant-scoped install.
func TestApplyInstallModeSelectionRejectsUnsupportedTenantScoped(t *testing.T) {
	supportsMultiTenant := false
	manifest := &catalog.Manifest{
		ID:                  "plugin-tenant-unsupported-install-mode",
		ScopeNature:         pluginv1.ScopeNatureTenantAware.String(),
		SupportsMultiTenant: &supportsMultiTenant,
		DefaultInstallMode:  pluginv1.InstallModeGlobal.String(),
	}

	err := applyInstallModeSelection(manifest, pluginv1.InstallModeTenantScoped.String())
	if !bizerr.Is(err, CodePluginInstallModeInvalidForScopeNature) {
		t.Fatalf("expected unsupported tenant-scoped install mode bizerr, got %v", err)
	}
	if manifest.DefaultInstallMode != pluginv1.InstallModeGlobal.String() {
		t.Fatalf("expected unsupported tenant governance to keep global install mode, got %s", manifest.DefaultInstallMode)
	}
}

// TestBuildLifecycleAuthorizedHostServicesDropsUnconfirmedResources verifies
// lifecycle handlers do not receive resource-scoped host services unless the
// operation carries an explicit host authorization decision.
func TestBuildLifecycleAuthorizedHostServicesDropsUnconfirmedResources(t *testing.T) {
	hostServices := []*protocol.HostServiceSpec{
		{
			Service: protocol.HostServiceRuntime,
			Methods: []string{
				protocol.HostServiceMethodRuntimeLogWrite,
			},
		},
		{
			Service: protocol.HostServiceStorage,
			Methods: []string{
				protocol.HostServiceMethodStorageGet,
			},
			Paths: []string{"private-files/"},
		},
		{
			Service: protocol.HostServiceHostConfig,
			Methods: []string{
				protocol.HostServiceMethodHostConfigGet,
			},
			Keys: []string{"ai.default"},
		},
	}

	withoutAuthorization, err := buildLifecycleAuthorizedHostServices("plugin-test-lifecycle", hostServices, nil)
	if err != nil {
		t.Fatalf("expected lifecycle host services to normalize, got error: %v", err)
	}
	if len(withoutAuthorization) != 1 || withoutAuthorization[0].Service != protocol.HostServiceRuntime {
		t.Fatalf("expected only capability host service without authorization, got %#v", withoutAuthorization)
	}

	withAuthorization, err := buildLifecycleAuthorizedHostServices(
		"plugin-test-lifecycle",
		hostServices,
		&store.HostServiceAuthorizationInput{
			Services: []*store.HostServiceAuthorizationDecision{
				{
					Service: protocol.HostServiceStorage,
					Paths:   []string{"private-files/"},
				},
				{
					Service: protocol.HostServiceHostConfig,
					Keys:    []string{"ai.default"},
				},
			},
		},
	)
	if err != nil {
		t.Fatalf("expected lifecycle host service authorization to normalize, got error: %v", err)
	}
	if len(withAuthorization) != 3 {
		t.Fatalf("expected runtime and authorized storage/config host services, got %#v", withAuthorization)
	}
}
