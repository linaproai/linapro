// This file scans plugin directories and validates convention-based manifest,
// SQL, page, and slot resources discovered from the plugin workspace.

package catalog

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gfile"
	"gopkg.in/yaml.v3"

	"lina-core/pkg/pluginbridge"
	"lina-core/pkg/pluginfs"
	"lina-core/pkg/pluginhost"
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
					"插件ID重复: %s 同时出现在 %s 和 %s",
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

// scanSourceManifests scans source plugins from apps/lina-plugins. Runtime
// sample directories are skipped because their clear-text plugin.yaml files
// are only build inputs, not runtime discovery sources.
func (s *serviceImpl) scanSourceManifests() ([]*Manifest, error) {
	pluginRootDir, err := s.resolvePluginRootDir()
	if err != nil {
		return s.ScanEmbeddedSourceManifests()
	}

	manifestFiles, err := gfile.ScanDirFile(pluginRootDir, "plugin.yaml", true)
	if err != nil {
		return nil, err
	}
	sort.Strings(manifestFiles)
	if len(manifestFiles) == 0 {
		return s.ScanEmbeddedSourceManifests()
	}

	manifests := make([]*Manifest, 0, len(manifestFiles))
	for _, manifestFile := range manifestFiles {
		content := gfile.GetBytes(manifestFile)
		if len(content) == 0 {
			return nil, gerror.Newf("插件清单为空: %s", manifestFile)
		}

		manifest := &Manifest{}
		if err = yaml.Unmarshal(content, manifest); err != nil {
			return nil, gerror.Wrapf(err, "解析插件清单失败: %s", manifestFile)
		}
		if NormalizeType(manifest.Type) == TypeDynamic {
			continue
		}
		if sourcePlugin, ok := pluginhost.GetSourcePlugin(strings.TrimSpace(manifest.ID)); ok {
			manifest.SourcePlugin = sourcePlugin
		}
		if err = s.ValidateManifest(manifest, manifestFile); err != nil {
			return nil, err
		}
		manifest.ManifestPath = manifestFile
		manifest.RootDir = filepath.Dir(manifestFile)
		// Load backend declarations after the manifest passes structural validation so
		// source-plugin resource scanning always starts from a trusted plugin root.
		if s.backendLoader != nil {
			if err = s.backendLoader.LoadPluginBackendConfig(manifest); err != nil {
				return nil, err
			}
		}

		manifests = append(manifests, manifest)
	}
	if len(manifests) == 0 {
		return s.ScanEmbeddedSourceManifests()
	}
	return manifests, nil
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
			return nil, gerror.Wrapf(loadErr, "解析动态插件产物失败: %s", artifactPath)
		}
		if previousPath, ok := seenIDs[manifest.ID]; ok {
			return nil, gerror.Newf(
				"动态插件ID重复: %s 同时出现在 %s 和 %s",
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
		return nil, gerror.Newf("动态插件缺少嵌入清单: %s", artifactPath)
	}

	manifest := &Manifest{
		ID:               strings.TrimSpace(artifact.Manifest.ID),
		Name:             strings.TrimSpace(artifact.Manifest.Name),
		Version:          strings.TrimSpace(artifact.Manifest.Version),
		Type:             NormalizeType(artifact.Manifest.Type).String(),
		Description:      strings.TrimSpace(artifact.Manifest.Description),
		Menus:            artifact.Manifest.Menus,
		ManifestPath:     "",
		RootDir:          filepath.Dir(artifactPath),
		Routes:           artifact.RouteContracts,
		BridgeSpec:       artifact.BridgeSpec,
		HostCapabilities: pluginbridge.CapabilityMapFromHostServices(artifact.HostServices),
		HostServices:     pluginbridge.NormalizeHostServiceSpecs(artifact.HostServices),
		RuntimeArtifact:  artifact,
	}
	if err = s.ValidateUploadedRuntimeManifest(manifest); err != nil {
		return nil, gerror.Wrapf(err, "动态插件嵌入清单不合法: %s", artifactPath)
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
		return gerror.Newf("插件清单文件为空: %s", filePath)
	}
	return yaml.Unmarshal(content, manifest)
}

// resolvePluginRootDir resolves the plugin root directory from the current working directory.
func (s *serviceImpl) resolvePluginRootDir() (string, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	repoRoot, err := FindRepoRoot(workingDir)
	if err == nil {
		pluginRootDir := filepath.Join(repoRoot, "apps", "lina-plugins")
		if gfile.Exists(pluginRootDir) && gfile.IsDir(pluginRootDir) {
			return pluginRootDir, nil
		}
	}

	candidateDirs := []string{
		filepath.Join(workingDir, "apps", "lina-plugins"),
		filepath.Join(workingDir, "..", "lina-plugins"),
		filepath.Join(workingDir, "..", "..", "lina-plugins"),
	}

	for _, dir := range candidateDirs {
		cleanPath := filepath.Clean(dir)
		if gfile.Exists(cleanPath) && gfile.IsDir(cleanPath) {
			return cleanPath, nil
		}
	}

	return "", gerror.Newf("未找到插件目录，候选路径: %s", strings.Join(candidateDirs, ", "))
}

// resolveRuntimeStorageDir resolves the configured runtime WASM storage directory.
// Relative paths are anchored at the repository root when available so uploads,
// manual copies, and automated scans all agree on one shared path.
func (s *serviceImpl) resolveRuntimeStorageDir(ctx context.Context) (string, error) {
	storagePath := s.configSvc.GetPluginDynamicStoragePath(ctx)
	if filepath.IsAbs(storagePath) {
		return filepath.Clean(storagePath), nil
	}

	workingDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if repoRoot, repoErr := FindRepoRoot(workingDir); repoErr == nil {
		return filepath.Clean(filepath.Join(repoRoot, storagePath)), nil
	}
	return filepath.Clean(filepath.Join(workingDir, storagePath)), nil
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
	return pluginfs.DiscoverSQLPaths(rootDir, uninstall)
}

// DiscoverPagePaths discovers plugin page source files by directory convention.
func (s *serviceImpl) DiscoverPagePaths(rootDir string) []string {
	return pluginfs.DiscoverVuePaths(rootDir, filepath.Join("frontend", "pages"))
}

// DiscoverSlotPaths discovers plugin slot source files by directory convention.
func (s *serviceImpl) DiscoverSlotPaths(rootDir string) []string {
	return pluginfs.DiscoverVuePaths(rootDir, filepath.Join("frontend", "slots"))
}
