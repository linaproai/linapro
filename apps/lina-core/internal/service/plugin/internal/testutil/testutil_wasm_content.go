// This file assembles and writes synthetic WebAssembly artifact content for plugin tests.

package testutil

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/pkg/pluginbridge"
)

// BuildTestRuntimeWasmContent assembles a synthetic WASM artifact byte slice for use in tests
// that need the raw bytes, such as upload-path tests that call StoreUploadedPackage directly.
func BuildTestRuntimeWasmContent(
	t *testing.T,
	manifest *catalog.ArtifactManifest,
	runtimeMetadata *catalog.ArtifactSpec,
	frontendAssets []*catalog.ArtifactFrontendAsset,
	installSQLAssets []*catalog.ArtifactSQLAsset,
	uninstallSQLAssets []*catalog.ArtifactSQLAsset,
	routeContracts []*pluginbridge.RouteContract,
	bridgeSpec *pluginbridge.BridgeSpec,
) []byte {
	t.Helper()
	return buildTestRuntimeWasmArtifactContent(
		t,
		manifest,
		runtimeMetadata,
		frontendAssets,
		installSQLAssets,
		uninstallSQLAssets,
		nil,
		routeContracts,
		bridgeSpec,
	)
}

// WriteRuntimeWasmArtifact writes one synthetic runtime WASM artifact fixture to disk.
func WriteRuntimeWasmArtifact(
	t *testing.T,
	filePath string,
	manifest *catalog.ArtifactManifest,
	runtimeMetadata *catalog.ArtifactSpec,
	frontendAssets []*catalog.ArtifactFrontendAsset,
	installSQLAssets []*catalog.ArtifactSQLAsset,
	uninstallSQLAssets []*catalog.ArtifactSQLAsset,
	cronContracts []*pluginbridge.CronContract,
	routeContracts []*pluginbridge.RouteContract,
	bridgeSpec *pluginbridge.BridgeSpec,
) {
	t.Helper()

	wasm := buildTestRuntimeWasmArtifactContent(
		t,
		manifest,
		runtimeMetadata,
		frontendAssets,
		installSQLAssets,
		uninstallSQLAssets,
		cronContracts,
		routeContracts,
		bridgeSpec,
	)
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		t.Fatalf("failed to create runtime artifact dir %s: %v", filePath, err)
	}
	tempFile, err := os.CreateTemp(filepath.Dir(filePath), filepath.Base(filePath)+".tmp-*")
	if err != nil {
		t.Fatalf("failed to create temp runtime wasm artifact %s: %v", filePath, err)
	}
	tempPath := tempFile.Name()
	defer func() {
		if cleanupErr := os.Remove(tempPath); cleanupErr != nil && !os.IsNotExist(cleanupErr) {
			t.Fatalf("failed to remove temp runtime wasm artifact %s: %v", tempPath, cleanupErr)
		}
	}()

	if _, err = tempFile.Write(wasm); err != nil {
		if closeErr := tempFile.Close(); closeErr != nil {
			t.Fatalf("failed to close temp runtime wasm artifact %s after write error: %v", filePath, closeErr)
		}
		t.Fatalf("failed to write temp runtime wasm artifact %s: %v", filePath, err)
	}
	if err = tempFile.Chmod(0o644); err != nil {
		if closeErr := tempFile.Close(); closeErr != nil {
			t.Fatalf("failed to close temp runtime wasm artifact %s after chmod error: %v", filePath, closeErr)
		}
		t.Fatalf("failed to chmod temp runtime wasm artifact %s: %v", filePath, err)
	}
	if err = tempFile.Close(); err != nil {
		t.Fatalf("failed to close temp runtime wasm artifact %s: %v", filePath, err)
	}
	if err = os.Rename(tempPath, filePath); err != nil {
		t.Fatalf("failed to move runtime wasm artifact into place %s: %v", filePath, err)
	}
}

// buildTestRuntimeWasmArtifactContent assembles one synthetic WASM binary with
// Lina custom sections for manifest, runtime metadata, routes, SQL, and bridge
// contracts used by component tests. Mock SQL assets are pulled from the
// runtimeMetadata.MockSQLAssets field so callers can opt in by populating that
// slice before invoking this helper.
func buildTestRuntimeWasmArtifactContent(
	t *testing.T,
	manifest *catalog.ArtifactManifest,
	runtimeMetadata *catalog.ArtifactSpec,
	frontendAssets []*catalog.ArtifactFrontendAsset,
	installSQLAssets []*catalog.ArtifactSQLAsset,
	uninstallSQLAssets []*catalog.ArtifactSQLAsset,
	_ []*pluginbridge.CronContract,
	routeContracts []*pluginbridge.RouteContract,
	bridgeSpec *pluginbridge.BridgeSpec,
) []byte {
	t.Helper()

	normalizeTestArtifactManifest(manifest)
	manifestContent, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("failed to marshal dynamic manifest: %v", err)
	}
	runtimeContent, err := json.Marshal(runtimeMetadata)
	if err != nil {
		t.Fatalf("failed to marshal runtime metadata: %v", err)
	}

	wasm := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}
	wasm = appendWasmCustomSection(wasm, pluginbridge.WasmSectionManifest, manifestContent)
	wasm = appendWasmCustomSection(wasm, pluginbridge.WasmSectionRuntime, runtimeContent)
	if len(frontendAssets) > 0 {
		frontendContent, marshalErr := json.Marshal(frontendAssets)
		if marshalErr != nil {
			t.Fatalf("failed to marshal frontend assets: %v", marshalErr)
		}
		wasm = appendWasmCustomSection(wasm, pluginbridge.WasmSectionFrontendAssets, frontendContent)
	}
	if len(installSQLAssets) > 0 {
		installContent, marshalErr := json.Marshal(installSQLAssets)
		if marshalErr != nil {
			t.Fatalf("failed to marshal install sql assets: %v", marshalErr)
		}
		wasm = appendWasmCustomSection(wasm, pluginbridge.WasmSectionInstallSQL, installContent)
	}
	if len(uninstallSQLAssets) > 0 {
		uninstallContent, marshalErr := json.Marshal(uninstallSQLAssets)
		if marshalErr != nil {
			t.Fatalf("failed to marshal uninstall sql assets: %v", marshalErr)
		}
		wasm = appendWasmCustomSection(wasm, pluginbridge.WasmSectionUninstallSQL, uninstallContent)
	}
	if len(runtimeMetadata.MockSQLAssets) > 0 {
		mockContent, marshalErr := json.Marshal(runtimeMetadata.MockSQLAssets)
		if marshalErr != nil {
			t.Fatalf("failed to marshal mock sql assets: %v", marshalErr)
		}
		wasm = appendWasmCustomSection(wasm, pluginbridge.WasmSectionMockSQL, mockContent)
	}
	if len(routeContracts) > 0 {
		routeContent, marshalErr := json.Marshal(routeContracts)
		if marshalErr != nil {
			t.Fatalf("failed to marshal route contracts: %v", marshalErr)
		}
		wasm = appendWasmCustomSection(wasm, pluginbridge.WasmSectionBackendRoutes, routeContent)
	}
	if bridgeSpec != nil {
		bridgeContent, marshalErr := json.Marshal(bridgeSpec)
		if marshalErr != nil {
			t.Fatalf("failed to marshal bridge spec: %v", marshalErr)
		}
		wasm = appendWasmCustomSection(wasm, pluginbridge.WasmSectionBackendBridge, bridgeContent)
	}
	if len(runtimeMetadata.HostServices) > 0 {
		hostServiceContent, marshalErr := json.Marshal(runtimeMetadata.HostServices)
		if marshalErr != nil {
			t.Fatalf("failed to marshal runtime host services: %v", marshalErr)
		}
		wasm = appendWasmCustomSection(wasm, pluginbridge.WasmSectionBackendHostServices, hostServiceContent)
	}
	return wasm
}

// normalizeTestArtifactManifest fills governance fields that every runtime
// artifact manifest must declare in production plugin packages.
func normalizeTestArtifactManifest(manifest *catalog.ArtifactManifest) {
	if manifest == nil {
		return
	}
	if manifest.ScopeNature == "" {
		manifest.ScopeNature = catalog.ScopeNatureTenantAware.String()
	}
	if manifest.SupportsMultiTenant == nil {
		manifest.SupportsMultiTenant = &DefaultTestSupportsMultiTenant
	}
	if manifest.DefaultInstallMode == "" {
		manifest.DefaultInstallMode = catalog.InstallModeTenantScoped.String()
	}
}

// appendWasmCustomSection appends one custom section payload to the in-memory
// WASM binary using the standard section-length encoding.
func appendWasmCustomSection(content []byte, name string, payload []byte) []byte {
	sectionPayload := append([]byte{}, encodeWasmULEB128(uint32(len(name)))...)
	sectionPayload = append(sectionPayload, []byte(name)...)
	sectionPayload = append(sectionPayload, payload...)

	result := append([]byte{}, content...)
	result = append(result, 0x00)
	result = append(result, encodeWasmULEB128(uint32(len(sectionPayload)))...)
	result = append(result, sectionPayload...)
	return result
}

// encodeWasmULEB128 encodes one unsigned integer using the WASM ULEB128 format.
func encodeWasmULEB128(value uint32) []byte {
	result := make([]byte, 0, 5)
	for {
		current := byte(value & 0x7f)
		value >>= 7
		if value != 0 {
			current |= 0x80
		}
		result = append(result, current)
		if value == 0 {
			return result
		}
	}
}
