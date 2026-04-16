// This file defines wasm custom section names and runtime metadata used by
// plugin bridge artifacts.

package pluginbridge

const (
	// WasmSectionManifest stores plugin identity metadata.
	WasmSectionManifest = "lina.plugin.manifest"
	// WasmSectionRuntime stores host-owned runtime metadata.
	WasmSectionRuntime = "lina.plugin.dynamic"
	// WasmSectionLegacyRuntime stores the legacy runtime metadata name.
	WasmSectionLegacyRuntime = "lina.plugin.runtime"
	// WasmSectionFrontendAssets stores embedded frontend assets.
	WasmSectionFrontendAssets = "lina.plugin.frontend.assets"
	// WasmSectionInstallSQL stores install-time SQL assets.
	WasmSectionInstallSQL = "lina.plugin.install.sql"
	// WasmSectionUninstallSQL stores uninstall-time SQL assets.
	WasmSectionUninstallSQL = "lina.plugin.uninstall.sql"
	// WasmSectionBackendHooks stores backend hook contracts.
	WasmSectionBackendHooks = "lina.plugin.backend.hooks"
	// WasmSectionBackendResources stores backend resource contracts.
	WasmSectionBackendResources = "lina.plugin.backend.resources"
	// WasmSectionBackendRoutes stores backend dynamic route contracts.
	WasmSectionBackendRoutes = "lina.plugin.backend.routes"
	// WasmSectionBackendBridge stores backend bridge ABI contracts.
	WasmSectionBackendBridge = "lina.plugin.backend.bridge"
	// WasmSectionBackendCapabilities stores a deprecated legacy capability section.
	WasmSectionBackendCapabilities = "lina.plugin.backend.capabilities"
	// WasmSectionBackendHostServices stores structured host service declarations.
	WasmSectionBackendHostServices = "lina.plugin.backend.host-services"
)

// RuntimeArtifactMetadata stores the host-owned runtime metadata section.
type RuntimeArtifactMetadata struct {
	RuntimeKind        string `json:"runtimeKind" yaml:"runtimeKind"`
	ABIVersion         string `json:"abiVersion" yaml:"abiVersion"`
	FrontendAssetCount int    `json:"frontendAssetCount,omitempty" yaml:"frontendAssetCount,omitempty"`
	SQLAssetCount      int    `json:"sqlAssetCount,omitempty" yaml:"sqlAssetCount,omitempty"`
	RouteCount         int    `json:"routeCount,omitempty" yaml:"routeCount,omitempty"`
}
