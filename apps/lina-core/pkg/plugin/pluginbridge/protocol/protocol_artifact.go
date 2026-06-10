// protocol_artifact.go exposes dynamic artifact metadata and custom-section helpers through the public protocol facade.
// These aliases keep WASM artifact section names stable for host and guest callers while the implementation remains internal.

package protocol

import "lina-core/pkg/plugin/pluginbridge/internal/artifact"

type RuntimeArtifactMetadata = artifact.RuntimeArtifactMetadata

const (
	WasmSectionManifest            = artifact.WasmSectionManifest
	WasmSectionRuntime             = artifact.WasmSectionRuntime
	WasmSectionFrontendAssets      = artifact.WasmSectionFrontendAssets
	WasmSectionI18NAssets          = artifact.WasmSectionI18NAssets
	WasmSectionAPIDocI18NAssets    = artifact.WasmSectionAPIDocI18NAssets
	WasmSectionInstallSQL          = artifact.WasmSectionInstallSQL
	WasmSectionUninstallSQL        = artifact.WasmSectionUninstallSQL
	WasmSectionMockSQL             = artifact.WasmSectionMockSQL
	WasmSectionManifestResources   = artifact.WasmSectionManifestResources
	WasmSectionBackendHooks        = artifact.WasmSectionBackendHooks
	WasmSectionBackendLifecycle    = artifact.WasmSectionBackendLifecycle
	WasmSectionBackendResources    = artifact.WasmSectionBackendResources
	WasmSectionBackendRoutes       = artifact.WasmSectionBackendRoutes
	WasmSectionBackendBridge       = artifact.WasmSectionBackendBridge
	WasmSectionBackendHostServices = artifact.WasmSectionBackendHostServices
)

var (
	ReadCustomSection  = artifact.ReadCustomSection
	ListCustomSections = artifact.ListCustomSections
)
