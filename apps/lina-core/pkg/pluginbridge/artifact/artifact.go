// Package artifact defines WASM custom-section names, artifact metadata, and
// artifact discovery helpers for Lina dynamic plugin bridge artifacts.
package artifact

// Wasm custom section constants define the metadata buckets embedded into
// plugin artifacts for host discovery and execution.
const (
	// WasmSectionManifest stores plugin identity metadata.
	WasmSectionManifest = "lina.plugin.manifest"
	// WasmSectionRuntime stores host-owned runtime metadata.
	WasmSectionRuntime = "lina.plugin.dynamic"
	// WasmSectionLegacyRuntime stores the legacy runtime metadata name.
	WasmSectionLegacyRuntime = "lina.plugin.runtime"
	// WasmSectionFrontendAssets stores embedded frontend assets.
	WasmSectionFrontendAssets = "lina.plugin.frontend.assets"
	// WasmSectionI18NAssets stores embedded plugin i18n message assets.
	WasmSectionI18NAssets = "lina.plugin.i18n.assets"
	// WasmSectionAPIDocI18NAssets stores embedded plugin API-documentation i18n assets.
	WasmSectionAPIDocI18NAssets = "lina.plugin.apidoc.i18n.assets"
	// WasmSectionInstallSQL stores install-time SQL assets.
	WasmSectionInstallSQL = "lina.plugin.install.sql"
	// WasmSectionUninstallSQL stores uninstall-time SQL assets.
	WasmSectionUninstallSQL = "lina.plugin.uninstall.sql"
	// WasmSectionMockSQL stores optional mock-data SQL assets that the host
	// only loads when the operator explicitly opts in at install time. Mock
	// data ships in its own section so install/uninstall counts stay
	// independent and consumers can detect the presence of mock data without
	// scanning the install section.
	WasmSectionMockSQL = "lina.plugin.mock.sql"
	// WasmSectionBackendHooks stores backend hook contracts.
	WasmSectionBackendHooks = "lina.plugin.backend.hooks"
	// WasmSectionBackendResources stores backend resource contracts.
	WasmSectionBackendResources = "lina.plugin.backend.resources"
	// WasmSectionBackendCrons stores deprecated legacy backend scheduled-job
	// contracts kept only for backward-tolerant artifact parsing tests.
	WasmSectionBackendCrons = "lina.plugin.backend.crons"
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
	// RuntimeKind identifies the runtime family required by the artifact.
	RuntimeKind string `json:"runtimeKind" yaml:"runtimeKind"`
	// ABIVersion identifies the bridge ABI version embedded into the artifact.
	ABIVersion string `json:"abiVersion" yaml:"abiVersion"`
	// FrontendAssetCount reports the number of embedded frontend asset entries.
	FrontendAssetCount int `json:"frontendAssetCount,omitempty" yaml:"frontendAssetCount,omitempty"`
	// I18NAssetCount reports the number of embedded i18n locale asset entries.
	I18NAssetCount int `json:"i18nAssetCount,omitempty" yaml:"i18nAssetCount,omitempty"`
	// APIDocI18NAssetCount reports the number of embedded API-documentation i18n locale asset entries.
	APIDocI18NAssetCount int `json:"apiDocI18nAssetCount,omitempty" yaml:"apiDocI18nAssetCount,omitempty"`
	// SQLAssetCount reports the number of embedded install + uninstall + mock SQL assets.
	SQLAssetCount int `json:"sqlAssetCount,omitempty" yaml:"sqlAssetCount,omitempty"`
	// MockSQLAssetCount reports the number of embedded mock-data SQL assets,
	// kept as a separate field so consumers can detect mock-data presence
	// without scanning the artifact sections.
	MockSQLAssetCount int `json:"mockSqlAssetCount,omitempty" yaml:"mockSqlAssetCount,omitempty"`
	// RouteCount reports the number of embedded backend route contracts.
	RouteCount int `json:"routeCount,omitempty" yaml:"routeCount,omitempty"`
}
