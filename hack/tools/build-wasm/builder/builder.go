// Package builder implements a standalone dynamic wasm packer for plugin
// source trees. It intentionally lives outside lina-core so development-time
// packaging does not depend on the host service module.
package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"lina-core/pkg/pluginbridge"
)

// BuildRuntimeWasmArtifactFromSource builds one dynamic wasm artifact from a clear-text plugin directory.
func BuildRuntimeWasmArtifactFromSource(pluginDir string) (*RuntimeBuildOutput, error) {
	return buildRuntimeWasmArtifactFromSource(pluginDir, "")
}

func buildRuntimeWasmArtifactFromSource(pluginDir string, outputDir string) (*RuntimeBuildOutput, error) {
	embeddedResources, err := loadEmbeddedStaticResourceSet(pluginDir)
	if err != nil {
		return nil, err
	}

	manifestPath := filepath.Join(pluginDir, "plugin.yaml")
	manifest, err := loadRuntimeBuildManifest(pluginDir, embeddedResources)
	if err != nil {
		return nil, err
	}
	if err := validateRuntimeBuildManifest(manifest, manifestPath); err != nil {
		return nil, err
	}

	frontendAssets, err := collectFrontendAssets(pluginDir, embeddedResources)
	if err != nil {
		return nil, err
	}
	installSQLAssets, err := collectSQLAssets(pluginDir, embeddedResources, false)
	if err != nil {
		return nil, err
	}
	uninstallSQLAssets, err := collectSQLAssets(pluginDir, embeddedResources, true)
	if err != nil {
		return nil, err
	}
	hookSpecs, err := collectHookSpecs(pluginDir, manifest.ID)
	if err != nil {
		return nil, err
	}
	resourceSpecs, err := collectResourceSpecs(pluginDir, manifest.ID)
	if err != nil {
		return nil, err
	}
	routeContracts, err := collectRouteContracts(pluginDir, manifest.ID)
	if err != nil {
		return nil, err
	}
	runtimePath, err := buildGuestRuntimeWasm(pluginDir, manifest.ID, outputDir)
	if err != nil {
		return nil, err
	}
	bridgeSpec := buildBridgeSpec(runtimePath)
	if err = pluginbridge.ValidateBridgeSpec(bridgeSpec); err != nil {
		return nil, err
	}

	content, err := buildRuntimeArtifactContent(
		manifest,
		frontendAssets,
		installSQLAssets,
		uninstallSQLAssets,
		hookSpecs,
		resourceSpecs,
		routeContracts,
		bridgeSpec,
		runtimePath,
	)
	if err != nil {
		return nil, err
	}

	artifactPath := filepath.Join(pluginDir, buildRuntimeBuildOutputRelativePath(manifest.ID))
	if strings.TrimSpace(outputDir) != "" {
		artifactPath = filepath.Join(filepath.Clean(outputDir), buildRuntimeArtifactFileName(manifest.ID))
	}

	return &RuntimeBuildOutput{
		ArtifactPath: artifactPath,
		Content:      content,
		RuntimePath:  runtimePath,
	}, nil
}

// WriteRuntimeWasmArtifactFromSource builds and writes one dynamic artifact into
// the requested output directory. When outputDir is empty inside the Lina
// workspace, the generated artifact is written under temp/output/.
func WriteRuntimeWasmArtifactFromSource(pluginDir string, outputDir string) (*RuntimeBuildOutput, error) {
	resolvedOutputDir, err := resolveRuntimeArtifactOutputDir(pluginDir, outputDir)
	if err != nil {
		return nil, err
	}

	out, err := buildRuntimeWasmArtifactFromSource(pluginDir, resolvedOutputDir)
	if err != nil {
		return nil, err
	}
	if err = os.MkdirAll(filepath.Dir(out.ArtifactPath), 0o755); err != nil {
		return nil, fmt.Errorf("failed to create dynamic artifact directory: %w", err)
	}
	if err = os.WriteFile(out.ArtifactPath, out.Content, 0o644); err != nil {
		return nil, fmt.Errorf("failed to write dynamic artifact: %w", err)
	}
	return out, nil
}
