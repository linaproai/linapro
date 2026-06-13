// This file creates runtime artifact files in the isolated plugin test storage.

package testutil

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/pkg/plugin/pluginbridge/protocol"
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
			Type:                plugintypes.TypeDynamic.String(),
			ScopeNature:         plugintypes.ScopeNatureTenantAware.String(),
			SupportsMultiTenant: &DefaultTestSupportsMultiTenant,
			DefaultInstallMode:  plugintypes.InstallModeTenantScoped.String(),
			PublicAssets:        runtimePublicAssetsForFrontendAssets(DefaultTestRuntimeFrontendAssets()),
		},
		&catalog.ArtifactSpec{
			RuntimeKind:        protocol.RuntimeKindWasm,
			ABIVersion:         protocol.SupportedABIVersion,
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
			Type:                plugintypes.TypeDynamic.String(),
			ScopeNature:         plugintypes.ScopeNatureTenantAware.String(),
			SupportsMultiTenant: &DefaultTestSupportsMultiTenant,
			DefaultInstallMode:  plugintypes.InstallModeTenantScoped.String(),
			Menus:               menus,
			PublicAssets:        runtimePublicAssetsForFrontendAssets(DefaultTestRuntimeFrontendAssets()),
		},
		&catalog.ArtifactSpec{
			RuntimeKind:        protocol.RuntimeKindWasm,
			ABIVersion:         protocol.SupportedABIVersion,
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
	routeContracts []*protocol.RouteContract,
	bridgeSpec *protocol.BridgeSpec,
) string {
	t.Helper()

	return CreateTestRuntimeStorageArtifactWithFrontendAssetsMenusAndBackendContracts(
		t,
		pluginID,
		pluginName,
		version,
		frontendAssets,
		nil,
		installSQLAssets,
		uninstallSQLAssets,
		routeContracts,
		bridgeSpec,
	)
}

// CreateTestRuntimeStorageArtifactWithHostServices creates one runtime
// artifact carrying structured host service declarations.
func CreateTestRuntimeStorageArtifactWithHostServices(
	t *testing.T,
	pluginID string,
	pluginName string,
	version string,
	hostServices []*protocol.HostServiceSpec,
	routeContracts []*protocol.RouteContract,
	bridgeSpec *protocol.BridgeSpec,
) string {
	t.Helper()

	storageDir := testDynamicStorageDir
	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		t.Fatalf("failed to create dynamic storage dir: %v", err)
	}

	artifactPath := filepath.Join(storageDir, runtime.BuildArtifactFileName(pluginID))
	t.Cleanup(func() {
		if cleanupErr := os.Remove(artifactPath); cleanupErr != nil && !os.IsNotExist(cleanupErr) {
			t.Fatalf("failed to remove runtime host-service artifact %s: %v", artifactPath, cleanupErr)
		}
	})

	frontendAssets := DefaultTestRuntimeFrontendAssets()
	WriteRuntimeWasmArtifact(
		t,
		artifactPath,
		&catalog.ArtifactManifest{
			ID:                  pluginID,
			Name:                pluginName,
			Version:             version,
			Type:                plugintypes.TypeDynamic.String(),
			ScopeNature:         plugintypes.ScopeNatureTenantAware.String(),
			SupportsMultiTenant: &DefaultTestSupportsMultiTenant,
			DefaultInstallMode:  plugintypes.InstallModeTenantScoped.String(),
			PublicAssets:        runtimePublicAssetsForFrontendAssets(frontendAssets),
		},
		&catalog.ArtifactSpec{
			RuntimeKind:        protocol.RuntimeKindWasm,
			ABIVersion:         protocol.SupportedABIVersion,
			FrontendAssetCount: len(frontendAssets),
			RouteCount:         len(routeContracts),
			HostServices:       hostServices,
		},
		frontendAssets,
		nil,
		nil,
		nil,
		routeContracts,
		bridgeSpec,
	)
	return artifactPath
}

// CreateTestRuntimeStorageArtifactWithFrontendAssetsMenusAndBackendContracts creates one runtime artifact with menu and backend contract sections.
func CreateTestRuntimeStorageArtifactWithFrontendAssetsMenusAndBackendContracts(
	t *testing.T,
	pluginID string,
	pluginName string,
	version string,
	frontendAssets []*catalog.ArtifactFrontendAsset,
	menus []*catalog.MenuSpec,
	installSQLAssets []*catalog.ArtifactSQLAsset,
	uninstallSQLAssets []*catalog.ArtifactSQLAsset,
	routeContracts []*protocol.RouteContract,
	bridgeSpec *protocol.BridgeSpec,
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
			Type:                plugintypes.TypeDynamic.String(),
			ScopeNature:         plugintypes.ScopeNatureTenantAware.String(),
			SupportsMultiTenant: &DefaultTestSupportsMultiTenant,
			DefaultInstallMode:  plugintypes.InstallModeTenantScoped.String(),
			Menus:               menus,
			PublicAssets:        runtimePublicAssetsForFrontendAssets(frontendAssets),
		},
		&catalog.ArtifactSpec{
			RuntimeKind:        protocol.RuntimeKindWasm,
			ABIVersion:         protocol.SupportedABIVersion,
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
			Path:          "frontend/pages/index.html",
			ContentBase64: base64.StdEncoding.EncodeToString([]byte("<html><body>dynamic frontend</body></html>")),
			ContentType:   "text/html; charset=utf-8",
		},
		{
			Path:          "frontend/pages/assets/app.js",
			ContentBase64: base64.StdEncoding.EncodeToString([]byte("console.log('dynamic frontend');")),
			ContentType:   "application/javascript",
		},
	}
}

// runtimePublicAssetsForFrontendAssets exposes the default runtime test bundle
// only when the artifact actually carries matching frontend/pages assets.
func runtimePublicAssetsForFrontendAssets(frontendAssets []*catalog.ArtifactFrontendAsset) []*catalog.PublicAssetSpec {
	for _, asset := range frontendAssets {
		if asset == nil {
			continue
		}
		assetPath := strings.Trim(strings.ReplaceAll(strings.TrimSpace(asset.Path), "\\", "/"), "/")
		if assetPath == "frontend/pages" || strings.HasPrefix(assetPath, "frontend/pages/") {
			return defaultTestRuntimePublicAssets()
		}
	}
	return nil
}

// defaultTestRuntimePublicAssets exposes the default runtime test bundle
// through the same public_assets declaration required in production packages.
func defaultTestRuntimePublicAssets() []*catalog.PublicAssetSpec {
	return []*catalog.PublicAssetSpec{
		{Source: "frontend/pages", Mount: "/"},
	}
}
