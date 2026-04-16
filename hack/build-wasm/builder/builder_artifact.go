// This file assembles the final runtime artifact bytes and appends Lina custom
// sections for manifest, assets, routes, and bridge metadata.

package builder

import (
	"encoding/json"
	"os"
	"strings"

	"lina-core/pkg/pluginbridge"
)

func buildRuntimeArtifactContent(
	manifest *pluginManifest,
	frontendAssets []*frontendAsset,
	installSQLAssets []*sqlAsset,
	uninstallSQLAssets []*sqlAsset,
	hookSpecs []*hookSpec,
	resourceSpecs []*resourceSpec,
	routeContracts []*pluginbridge.RouteContract,
	bridgeSpec *pluginbridge.BridgeSpec,
	runtimePath string,
) ([]byte, error) {
	manifestPayload, err := json.Marshal(&dynamicArtifactManifest{
		ID:          manifest.ID,
		Name:        manifest.Name,
		Version:     manifest.Version,
		Type:        pluginTypeDynamic,
		Description: manifest.Description,
		Menus:       manifest.Menus,
	})
	if err != nil {
		return nil, err
	}
	runtimePayload, err := json.Marshal(&dynamicArtifactMetadata{
		RuntimeKind:        pluginDynamicKindWasm,
		ABIVersion:         pluginDynamicSupportedABIVersion,
		FrontendAssetCount: len(frontendAssets),
		SQLAssetCount:      len(installSQLAssets) + len(uninstallSQLAssets),
		RouteCount:         len(routeContracts),
	})
	if err != nil {
		return nil, err
	}

	content := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}
	if strings.TrimSpace(runtimePath) != "" {
		runtimeBytes, err := os.ReadFile(runtimePath)
		if err != nil {
			return nil, err
		}
		content = runtimeBytes
	}
	content = appendWasmCustomSection(content, pluginDynamicWasmSectionManifest, manifestPayload)
	content = appendWasmCustomSection(content, pluginDynamicWasmSectionDynamic, runtimePayload)

	if len(frontendAssets) > 0 {
		payload, err := json.Marshal(frontendAssets)
		if err != nil {
			return nil, err
		}
		content = appendWasmCustomSection(content, pluginDynamicWasmSectionFrontend, payload)
	}
	if len(installSQLAssets) > 0 {
		payload, err := json.Marshal(installSQLAssets)
		if err != nil {
			return nil, err
		}
		content = appendWasmCustomSection(content, pluginDynamicWasmSectionInstallSQL, payload)
	}
	if len(uninstallSQLAssets) > 0 {
		payload, err := json.Marshal(uninstallSQLAssets)
		if err != nil {
			return nil, err
		}
		content = appendWasmCustomSection(content, pluginDynamicWasmSectionUninstallSQL, payload)
	}
	if len(hookSpecs) > 0 {
		payload, err := json.Marshal(hookSpecs)
		if err != nil {
			return nil, err
		}
		content = appendWasmCustomSection(content, pluginDynamicWasmSectionBackendHooks, payload)
	}
	if len(resourceSpecs) > 0 {
		payload, err := json.Marshal(resourceSpecs)
		if err != nil {
			return nil, err
		}
		content = appendWasmCustomSection(content, pluginDynamicWasmSectionBackendRes, payload)
	}
	if len(routeContracts) > 0 {
		payload, err := json.Marshal(routeContracts)
		if err != nil {
			return nil, err
		}
		content = appendWasmCustomSection(content, pluginDynamicWasmSectionBackendRoutes, payload)
	}
	if bridgeSpec != nil {
		payload, err := json.Marshal(bridgeSpec)
		if err != nil {
			return nil, err
		}
		content = appendWasmCustomSection(content, pluginDynamicWasmSectionBackendBridge, payload)
	}
	if len(manifest.HostServices) > 0 {
		payload, err := json.Marshal(manifest.HostServices)
		if err != nil {
			return nil, err
		}
		content = appendWasmCustomSection(content, pluginDynamicWasmSectionBackendHostServices, payload)
	}
	return content, nil
}

func appendWasmCustomSection(content []byte, name string, payload []byte) []byte {
	section := make([]byte, 0, len(name)+len(payload)+8)
	section = appendULEB128(section, uint32(len(name)))
	section = append(section, []byte(name)...)
	section = append(section, payload...)

	content = append(content, 0x00)
	content = appendULEB128(content, uint32(len(section)))
	content = append(content, section...)
	return content
}

func appendULEB128(content []byte, value uint32) []byte {
	current := value
	for {
		part := byte(current & 0x7f)
		current >>= 7
		if current != 0 {
			part |= 0x80
		}
		content = append(content, part)
		if current == 0 {
			return content
		}
	}
}
