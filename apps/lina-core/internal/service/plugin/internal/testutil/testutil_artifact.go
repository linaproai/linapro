// This file creates runtime artifact files in the isolated plugin test storage.

package testutil

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/pkg/pluginbridge"
)

// DefaultTestSupportsMultiTenant is the tenant-governance flag used by runtime artifact fixtures.
var DefaultTestSupportsMultiTenant = true

// CreateTestRuntimeStorageArtifact creates one runtime artifact in the isolated test storage directory.
func CreateTestRuntimeStorageArtifact(
	t *testing.T,
	pluginID string,
	pluginName string,
	version string,
	installSQLAssets []*catalog.ArtifactSQLAsset,
	uninstallSQLAssets []*catalog.ArtifactSQLAsset,
) string {
	return CreateTestRuntimeStorageArtifactWithFrontendAssets(
		t,
		pluginID,
		pluginName,
		version,
		DefaultTestRuntimeFrontendAssets(),
		installSQLAssets,
		uninstallSQLAssets,
	)
}

// CreateTestRuntimeStorageArtifactWithFilename creates one runtime artifact with a custom storage file name.
// This is the low-level variant used when the test needs to place two artifacts with the same plugin ID
// under different file names in order to exercise duplicate-detection logic.
func CreateTestRuntimeStorageArtifactWithFilename(
	t *testing.T,
	fileName string,
	pluginID string,
	pluginName string,
	version string,
	installSQLAssets []*catalog.ArtifactSQLAsset,
	uninstallSQLAssets []*catalog.ArtifactSQLAsset,
) string {
	t.Helper()

	storageDir := testDynamicStorageDir
	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		t.Fatalf("failed to create dynamic storage dir: %v", err)
	}

	artifactPath := filepath.Join(storageDir, fileName)
	t.Cleanup(func() {
		if cleanupErr := os.Remove(artifactPath); cleanupErr != nil && !os.IsNotExist(cleanupErr) {
			t.Fatalf("failed to remove runtime storage artifact %s: %v", artifactPath, cleanupErr)
		}
	})

	WriteRuntimeWasmArtifact(
		t,
		artifactPath,
		&catalog.ArtifactManifest{
			ID:                  pluginID,
			Name:                pluginName,
			Version:             version,
			Type:                catalog.TypeDynamic.String(),
			ScopeNature:         catalog.ScopeNatureTenantAware.String(),
			SupportsMultiTenant: &DefaultTestSupportsMultiTenant,
			DefaultInstallMode:  catalog.InstallModeTenantScoped.String(),
		},
		&catalog.ArtifactSpec{
			RuntimeKind:        pluginbridge.RuntimeKindWasm,
			ABIVersion:         pluginbridge.SupportedABIVersion,
			FrontendAssetCount: len(DefaultTestRuntimeFrontendAssets()),
			SQLAssetCount:      len(installSQLAssets) + len(uninstallSQLAssets),
		},
		DefaultTestRuntimeFrontendAssets(),
		installSQLAssets,
		uninstallSQLAssets,
		nil,
		nil,
		nil,
	)
	return artifactPath
}

// CreateTestRuntimeStorageArtifactWithFrontendAssets creates one runtime artifact with custom frontend assets.
func CreateTestRuntimeStorageArtifactWithFrontendAssets(
	t *testing.T,
	pluginID string,
	pluginName string,
	version string,
	frontendAssets []*catalog.ArtifactFrontendAsset,
	installSQLAssets []*catalog.ArtifactSQLAsset,
	uninstallSQLAssets []*catalog.ArtifactSQLAsset,
) string {
	return CreateTestRuntimeStorageArtifactWithFrontendAssetsAndBackendContracts(
		t,
		pluginID,
		pluginName,
		version,
		frontendAssets,
		installSQLAssets,
		uninstallSQLAssets,
		nil,
		nil,
	)
}

// CreateTestRuntimeStorageArtifactWithMenus creates one runtime artifact with manifest menus.
func CreateTestRuntimeStorageArtifactWithMenus(
	t *testing.T,
	pluginID string,
	pluginName string,
	version string,
	menus []*catalog.MenuSpec,
	installSQLAssets []*catalog.ArtifactSQLAsset,
	uninstallSQLAssets []*catalog.ArtifactSQLAsset,
) string {
	t.Helper()

	storageDir := testDynamicStorageDir
	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		t.Fatalf("failed to create dynamic storage dir: %v", err)
	}

	artifactPath := filepath.Join(storageDir, runtime.BuildArtifactFileName(pluginID))
	t.Cleanup(func() {
		if cleanupErr := os.Remove(artifactPath); cleanupErr != nil && !os.IsNotExist(cleanupErr) {
			t.Fatalf("failed to remove runtime menu artifact %s: %v", artifactPath, cleanupErr)
		}
	})

	WriteRuntimeWasmArtifact(
		t,
		artifactPath,
		&catalog.ArtifactManifest{
			ID:                  pluginID,
			Name:                pluginName,
			Version:             version,
			Type:                catalog.TypeDynamic.String(),
			ScopeNature:         catalog.ScopeNatureTenantAware.String(),
			SupportsMultiTenant: &DefaultTestSupportsMultiTenant,
			DefaultInstallMode:  catalog.InstallModeTenantScoped.String(),
			Menus:               menus,
		},
		&catalog.ArtifactSpec{
			RuntimeKind:        pluginbridge.RuntimeKindWasm,
			ABIVersion:         pluginbridge.SupportedABIVersion,
			FrontendAssetCount: len(DefaultTestRuntimeFrontendAssets()),
			SQLAssetCount:      len(installSQLAssets) + len(uninstallSQLAssets),
		},
		DefaultTestRuntimeFrontendAssets(),
		installSQLAssets,
		uninstallSQLAssets,
		nil,
		nil,
		nil,
	)
	return artifactPath
}

// CreateTestRuntimeStorageArtifactWithFrontendAssetsAndBackendContracts creates one runtime artifact with full contract sections.
func CreateTestRuntimeStorageArtifactWithFrontendAssetsAndBackendContracts(
	t *testing.T,
	pluginID string,
	pluginName string,
	version string,
	frontendAssets []*catalog.ArtifactFrontendAsset,
	installSQLAssets []*catalog.ArtifactSQLAsset,
	uninstallSQLAssets []*catalog.ArtifactSQLAsset,
	routeContracts []*pluginbridge.RouteContract,
	bridgeSpec *pluginbridge.BridgeSpec,
) string {
	t.Helper()

	storageDir := testDynamicStorageDir
	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		t.Fatalf("failed to create dynamic storage dir: %v", err)
	}

	artifactPath := filepath.Join(storageDir, runtime.BuildArtifactFileName(pluginID))
	t.Cleanup(func() {
		if cleanupErr := os.Remove(artifactPath); cleanupErr != nil && !os.IsNotExist(cleanupErr) {
			t.Fatalf("failed to remove runtime contract artifact %s: %v", artifactPath, cleanupErr)
		}
	})

	WriteRuntimeWasmArtifact(
		t,
		artifactPath,
		&catalog.ArtifactManifest{
			ID:                  pluginID,
			Name:                pluginName,
			Version:             version,
			Type:                catalog.TypeDynamic.String(),
			ScopeNature:         catalog.ScopeNatureTenantAware.String(),
			SupportsMultiTenant: &DefaultTestSupportsMultiTenant,
			DefaultInstallMode:  catalog.InstallModeTenantScoped.String(),
		},
		&catalog.ArtifactSpec{
			RuntimeKind:        pluginbridge.RuntimeKindWasm,
			ABIVersion:         pluginbridge.SupportedABIVersion,
			FrontendAssetCount: len(frontendAssets),
			SQLAssetCount:      len(installSQLAssets) + len(uninstallSQLAssets),
			RouteCount:         len(routeContracts),
		},
		frontendAssets,
		installSQLAssets,
		uninstallSQLAssets,
		nil,
		routeContracts,
		bridgeSpec,
	)
	return artifactPath
}

// DefaultTestRuntimeFrontendAssets returns the default frontend assets used by runtime artifact fixtures.
func DefaultTestRuntimeFrontendAssets() []*catalog.ArtifactFrontendAsset {
	return []*catalog.ArtifactFrontendAsset{
		{
			Path:          "index.html",
			ContentBase64: base64.StdEncoding.EncodeToString([]byte("<html><body>dynamic frontend</body></html>")),
			ContentType:   "text/html; charset=utf-8",
		},
		{
			Path:          "assets/app.js",
			ContentBase64: base64.StdEncoding.EncodeToString([]byte("console.log('dynamic frontend');")),
			ContentType:   "application/javascript",
		},
	}
}
