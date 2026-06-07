// This file scans registered source plugins and runtime artifacts while keeping
// directory-convention helpers for manifest-owned resources.

package catalog

import (
	"bytes"
	"context"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gfile"
	"gopkg.in/yaml.v3"

	"lina-core/internal/service/plugin/internal/resourcefs"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// ScanManifests merges source-plugin discovery and runtime-wasm discovery
// into one normalized manifest list used by lifecycle and governance services.
func (s *serviceImpl) ScanManifests() ([]*Manifest, error) {
	sourceManifests, err := s.scanSourceManifests()
	if err != nil {
		return nil, err
	}
	runtimeManifests, err := s.scanRuntimeManifests(context.Background())
	if err != nil {
		return nil, err
	}

	manifests := make([]*Manifest, 0, len(sourceManifests)+len(runtimeManifests))
	seenIDs := make(map[string]string, len(sourceManifests)+len(runtimeManifests))
	for _, items := range [][]*Manifest{sourceManifests, runtimeManifests} {
		for _, manifest := range items {
			if manifest == nil {
				continue
			}
			location := buildDiscoveryLocation(manifest)
			if previousFile, ok := seenIDs[manifest.ID]; ok {
				return nil, gerror.Newf(
					"plugin ID is duplicated: %s appears in both %s and %s",
					manifest.ID,
					previousFile,
					location,
				)
			}
			seenIDs[manifest.ID] = location
			manifests = append(manifests, manifest)
		}
	}

	sort.Slice(manifests, func(i, j int) bool {
		return manifests[i].ID < manifests[j].ID
	})
	return manifests, nil
}

// scanSourceManifests scans source plugins registered into the host binary.
func (s *serviceImpl) scanSourceManifests() ([]*Manifest, error) {
	return s.ScanEmbeddedSourceManifests()
}

// scanRuntimeManifests scans the configured runtime wasm storage directory.
// Discovery is intentionally non-recursive so the host does not impose any extra
// outer directory convention beyond dropping .wasm files into storagePath.
func (s *serviceImpl) scanRuntimeManifests(ctx context.Context) ([]*Manifest, error) {
	storageDir, err := s.resolveRuntimeStorageDir(ctx)
	if err != nil {
		return nil, err
	}
	if !gfile.Exists(storageDir) || !gfile.IsDir(storageDir) {
		return []*Manifest{}, nil
	}

	artifactFiles, err := gfile.ScanDirFile(storageDir, "*.wasm", false)
	if err != nil {
		return nil, err
	}
	sort.Strings(artifactFiles)

	manifests := make([]*Manifest, 0, len(artifactFiles))
	seenIDs := make(map[string]string, len(artifactFiles))
	for _, artifactPath := range artifactFiles {
		manifest, loadErr := s.loadRuntimeManifestFromArtifact(artifactPath)
		if loadErr != nil {
			return nil, gerror.Wrapf(loadErr, "parse dynamic plugin artifact failed: %s", artifactPath)
		}
		if previousPath, ok := seenIDs[manifest.ID]; ok {
			return nil, gerror.Newf(
				"dynamic plugin ID is duplicated: %s appears in both %s and %s",
				manifest.ID,
				previousPath,
				artifactPath,
			)
		}
		seenIDs[manifest.ID] = artifactPath
		manifests = append(manifests, manifest)
	}
	return manifests, nil
}

// buildDiscoveryLocation returns the most relevant source location string for
// duplicate-discovery and error messages.
func buildDiscoveryLocation(manifest *Manifest) string {
	if manifest == nil {
		return ""
	}
	if manifest.RuntimeArtifact != nil && strings.TrimSpace(manifest.RuntimeArtifact.Path) != "" {
		return manifest.RuntimeArtifact.Path
	}
	if strings.TrimSpace(manifest.ManifestPath) != "" {
		return manifest.ManifestPath
	}
	return manifest.RootDir
}

// loadRuntimeManifestFromArtifact reads and validates a WASM artifact file and
// returns its embedded plugin manifest with fully-hydrated hook/resource specs.
func (s *serviceImpl) loadRuntimeManifestFromArtifact(artifactPath string) (*Manifest, error) {
	if s.artifactParser == nil {
		return nil, gerror.New("artifact parser not configured")
	}
	artifact, err := s.artifactParser.ParseRuntimeWasmArtifact(artifactPath)
	if err != nil {
		return nil, err
	}
	if artifact.Manifest == nil {
		return nil, gerror.Newf("dynamic plugin is missing embedded manifest: %s", artifactPath)
	}

	pluginID := strings.TrimSpace(artifact.Manifest.ID)
	hostServices, err := protocol.NormalizeHostServiceSpecsForPlugin(pluginID, artifact.HostServices)
	if err != nil {
		return nil, gerror.Wrapf(err, "dynamic plugin host service declaration is invalid: %s", artifactPath)
	}
	manifest := &Manifest{
		ID:                  pluginID,
		Name:                strings.TrimSpace(artifact.Manifest.Name),
		Version:             strings.TrimSpace(artifact.Manifest.Version),
		Type:                NormalizeType(artifact.Manifest.Type).String(),
		ScopeNature:         strings.TrimSpace(artifact.Manifest.ScopeNature),
		SupportsMultiTenant: artifact.Manifest.SupportsMultiTenant,
		DefaultInstallMode:  strings.TrimSpace(artifact.Manifest.DefaultInstallMode),
		Description:         strings.TrimSpace(artifact.Manifest.Description),
		Dependencies:        CloneDependencySpec(artifact.Manifest.Dependencies),
		Menus:               artifact.Manifest.Menus,
		PublicAssets:        ClonePublicAssetSpecs(artifact.Manifest.PublicAssets),
		ManifestPath:        "",
		RootDir:             filepath.Dir(artifactPath),
		LifecycleHandlers:   CloneLifecycleContracts(artifact.LifecycleContracts),
		Routes:              artifact.RouteContracts,
		BridgeSpec:          artifact.BridgeSpec,
		HostCapabilities:    protocol.CapabilityMapFromHostServices(hostServices),
		HostServices:        hostServices,
		RuntimeArtifact:     artifact,
	}
	if err = s.ValidateUploadedRuntimeManifest(manifest); err != nil {
		return nil, gerror.Wrapf(err, "dynamic plugin embedded manifest is invalid: %s", artifactPath)
	}
	artifact.Manifest.Type = manifest.Type
	// Runtime manifests are reloaded from both the mutable staging artifact and
	// archived active releases. Always hydrate embedded backend contracts here so
	// every caller receives a complete runtime manifest with hook/resource specs.
	if s.backendLoader != nil {
		if err = s.backendLoader.LoadPluginBackendConfig(manifest); err != nil {
			return nil, err
		}
	}
	return manifest, nil
}

// LoadManifestFromYAML parses a plugin.yaml file at the given path into a Manifest.
func (s *serviceImpl) LoadManifestFromYAML(filePath string, manifest *Manifest) error {
	content := gfile.GetBytes(filePath)
	if len(content) == 0 {
		return gerror.Newf("plugin manifest file is empty: %s", filePath)
	}
	if err := validateManifestDependencySchema(content, filePath); err != nil {
		return err
	}
	return yaml.Unmarshal(content, manifest)
}

// validateManifestDependencySchema validates current dependency entry fields
// before lenient manifest decoding can ignore unsupported plugin policies.
func validateManifestDependencySchema(content []byte, fileLabel string) error {
	var root yaml.Node
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	if err := decoder.Decode(&root); err != nil {
		return err
	}
	document := &root
	if root.Kind == yaml.DocumentNode && len(root.Content) > 0 {
		document = root.Content[0]
	}
	if document == nil || document.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(document.Content); i += 2 {
		key := strings.TrimSpace(document.Content[i].Value)
		if key == "dependencies" {
			if err := rejectUnsupportedDependencyFields(document.Content[i+1], fileLabel); err != nil {
				return err
			}
		}
	}
	return nil
}

// rejectUnsupportedDependencyFields rejects removed dependency policy fields.
func rejectUnsupportedDependencyFields(node *yaml.Node, fileLabel string) error {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		key := strings.TrimSpace(node.Content[i].Value)
		if key == "plugins" {
			if err := rejectUnsupportedPluginDependencyFields(node.Content[i+1], fileLabel); err != nil {
				return err
			}
		}
	}
	return nil
}

// rejectUnsupportedPluginDependencyFields rejects required/install policy fields
// from dependencies.plugins entries. Declaring a plugin dependency is always a
// hard dependency; automatic install policy is outside plugin manifests.
func rejectUnsupportedPluginDependencyFields(node *yaml.Node, fileLabel string) error {
	if node == nil || node.Kind != yaml.SequenceNode {
		return nil
	}
	for index, item := range node.Content {
		if item == nil || item.Kind != yaml.MappingNode {
			continue
		}
		for i := 0; i+1 < len(item.Content); i += 2 {
			key := strings.TrimSpace(item.Content[i].Value)
			if key != "id" && key != "version" {
				return gerror.Newf("plugin manifest field dependencies.plugins[%d].%s is not supported; plugin dependencies only support id and version: %s", index, key, fileLabel)
			}
		}
	}
	return nil
}

// resolveRuntimeStorageDir resolves the configured runtime WASM storage
// directory. The config service already anchors relative paths so catalog
// scanning, uploads, and host-service storage all share one directory.
func (s *serviceImpl) resolveRuntimeStorageDir(ctx context.Context) (string, error) {
	storagePath := strings.TrimSpace(s.configSvc.GetPluginDynamicStoragePath(ctx))
	if storagePath == "" {
		return "", gerror.New("runtime WASM storage path cannot be empty")
	}
	if filepath.IsAbs(storagePath) {
		return filepath.Clean(storagePath), nil
	}

	absolutePath, err := filepath.Abs(storagePath)
	if err != nil {
		return "", err
	}
	return filepath.Clean(absolutePath), nil
}

// RuntimeStorageDir returns the absolute path of the runtime WASM storage directory
// configured in plugin.dynamic.storagePath.
func (s *serviceImpl) RuntimeStorageDir(ctx context.Context) (string, error) {
	return s.resolveRuntimeStorageDir(ctx)
}

// LoadManifestFromArtifactPath loads and validates a dynamic plugin manifest from
// the given absolute WASM artifact file path.
func (s *serviceImpl) LoadManifestFromArtifactPath(artifactPath string) (*Manifest, error) {
	return s.loadRuntimeManifestFromArtifact(artifactPath)
}

// DiscoverSQLPaths discovers plugin SQL files by directory convention.
func (s *serviceImpl) DiscoverSQLPaths(rootDir string, uninstall bool) []string {
	return resourcefs.DiscoverSQLPaths(rootDir, uninstall)
}

// DiscoverMockSQLPaths discovers plugin mock-data SQL files by directory convention.
func (s *serviceImpl) DiscoverMockSQLPaths(rootDir string) []string {
	return resourcefs.DiscoverMockSQLPaths(rootDir)
}

// DiscoverPagePaths discovers plugin page source files by directory convention.
func (s *serviceImpl) DiscoverPagePaths(rootDir string) []string {
	return resourcefs.DiscoverVuePaths(rootDir, filepath.Join("frontend", "pages"))
}

// DiscoverSlotPaths discovers plugin slot source files by directory convention.
func (s *serviceImpl) DiscoverSlotPaths(rootDir string) []string {
	return resourcefs.DiscoverVuePaths(rootDir, filepath.Join("frontend", "slots"))
}
