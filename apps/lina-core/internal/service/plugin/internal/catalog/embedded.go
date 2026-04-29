// This file scans source plugins backed by embedded filesystems and resolves
// embedded manifest, SQL, and frontend assets from registered source plugins.

package catalog

import (
	"io/fs"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gfile"
	"gopkg.in/yaml.v3"

	"lina-core/pkg/pluginfs"
	"lina-core/pkg/pluginhost"
)

// ScanEmbeddedSourceManifests discovers manifests from all registered embedded source plugins.
func (s *serviceImpl) ScanEmbeddedSourceManifests() ([]*Manifest, error) {
	sourcePlugins := pluginhost.ListSourcePlugins()
	if len(sourcePlugins) == 0 {
		return []*Manifest{}, nil
	}

	sort.Slice(sourcePlugins, func(i, j int) bool {
		return sourcePlugins[i].ID() < sourcePlugins[j].ID()
	})

	manifests := make([]*Manifest, 0, len(sourcePlugins))
	for _, sourcePlugin := range sourcePlugins {
		if sourcePlugin == nil {
			continue
		}

		embeddedFiles := sourcePlugin.GetEmbeddedFiles()
		if embeddedFiles == nil {
			return nil, gerror.Newf("source plugin is missing embedded resource declaration: %s", sourcePlugin.ID())
		}

		manifestContent, err := fs.ReadFile(embeddedFiles, pluginfs.EmbeddedManifestPath)
		if err != nil {
			return nil, gerror.Wrapf(err, "read source plugin embedded manifest failed: %s", sourcePlugin.ID())
		}

		manifest := &Manifest{
			ManifestPath: pluginfs.BuildEmbeddedManifestPath(sourcePlugin.ID(), pluginfs.EmbeddedManifestPath),
			SourcePlugin: sourcePlugin,
		}
		if err = yaml.Unmarshal(manifestContent, manifest); err != nil {
			return nil, gerror.Wrapf(err, "parse source plugin embedded manifest failed: %s", sourcePlugin.ID())
		}
		if err = s.ValidateManifest(manifest, manifest.ManifestPath); err != nil {
			return nil, err
		}
		if s.backendLoader != nil {
			if err = s.backendLoader.LoadPluginBackendConfig(manifest); err != nil {
				return nil, err
			}
		}

		manifests = append(manifests, manifest)
	}
	return manifests, nil
}

// GetSourcePluginEmbeddedFiles returns the embedded filesystem for a source plugin manifest.
func GetSourcePluginEmbeddedFiles(manifest *Manifest) fs.FS {
	if manifest == nil || manifest.SourcePlugin == nil {
		return nil
	}
	return manifest.SourcePlugin.GetEmbeddedFiles()
}

// HasSourcePluginEmbeddedFiles reports whether a manifest has an associated embedded filesystem.
func HasSourcePluginEmbeddedFiles(manifest *Manifest) bool {
	return GetSourcePluginEmbeddedFiles(manifest) != nil
}

// ReadSourcePluginManifestContent reads the raw manifest content from an embedded or
// filesystem-backed source plugin.
func (s *serviceImpl) ReadSourcePluginManifestContent(manifest *Manifest) ([]byte, error) {
	if embeddedFiles := GetSourcePluginEmbeddedFiles(manifest); embeddedFiles != nil {
		content, err := fs.ReadFile(embeddedFiles, pluginfs.EmbeddedManifestPath)
		if err != nil {
			return nil, gerror.Wrapf(err, "read source plugin embedded manifest failed: %s", manifest.ID)
		}
		return content, nil
	}
	if manifest == nil || strings.TrimSpace(manifest.ManifestPath) == "" {
		return nil, gerror.New("source plugin manifest path cannot be empty")
	}
	content := gfile.GetBytes(manifest.ManifestPath)
	if len(content) == 0 {
		return nil, gerror.Newf("plugin manifest is empty: %s", manifest.ManifestPath)
	}
	return content, nil
}

// ReadSourcePluginAssetContent reads one asset relative path from an embedded or filesystem source plugin.
func (s *serviceImpl) ReadSourcePluginAssetContent(manifest *Manifest, relativePath string) (string, error) {
	normalizedPath, err := pluginfs.NormalizeRelativePath(relativePath)
	if err != nil {
		return "", err
	}

	if embeddedFiles := GetSourcePluginEmbeddedFiles(manifest); embeddedFiles != nil {
		content, err := fs.ReadFile(embeddedFiles, normalizedPath)
		if err != nil {
			return "", gerror.Wrapf(err, "read source plugin embedded asset failed: %s", normalizedPath)
		}
		return strings.TrimSpace(string(content)), nil
	}

	sqlPath, err := pluginfs.ResolveResourcePath(manifest.RootDir, normalizedPath)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(gfile.GetContents(sqlPath)), nil
}

// ListInstallSQLPaths returns the ordered install SQL file paths for a source plugin manifest.
func (s *serviceImpl) ListInstallSQLPaths(manifest *Manifest) []string {
	if embeddedFiles := GetSourcePluginEmbeddedFiles(manifest); embeddedFiles != nil {
		return pluginfs.DiscoverSQLPathsFromFS(embeddedFiles, false)
	}
	if manifest == nil {
		return []string{}
	}
	return s.DiscoverSQLPaths(manifest.RootDir, false)
}

// ListUninstallSQLPaths returns the ordered uninstall SQL file paths for a source plugin manifest.
func (s *serviceImpl) ListUninstallSQLPaths(manifest *Manifest) []string {
	if embeddedFiles := GetSourcePluginEmbeddedFiles(manifest); embeddedFiles != nil {
		return pluginfs.DiscoverSQLPathsFromFS(embeddedFiles, true)
	}
	if manifest == nil {
		return []string{}
	}
	return s.DiscoverSQLPaths(manifest.RootDir, true)
}

// ListMockSQLPaths returns the ordered mock-data SQL file paths for a source plugin manifest.
// Mock-data files are deliberately excluded from install/uninstall scans and are loaded only
// when the operator explicitly opts in at install time.
func (s *serviceImpl) ListMockSQLPaths(manifest *Manifest) []string {
	if embeddedFiles := GetSourcePluginEmbeddedFiles(manifest); embeddedFiles != nil {
		return pluginfs.DiscoverMockSQLPathsFromFS(embeddedFiles)
	}
	if manifest == nil {
		return []string{}
	}
	return s.DiscoverMockSQLPaths(manifest.RootDir)
}

// HasMockSQLData reports whether the manifest carries any mock-data SQL assets,
// covering both source-plugin directory scans and dynamic-plugin embedded artifacts.
func (s *serviceImpl) HasMockSQLData(manifest *Manifest) bool {
	if manifest == nil {
		return false
	}
	if manifest.RuntimeArtifact != nil {
		return len(manifest.RuntimeArtifact.MockSQLAssets) > 0
	}
	return len(s.ListMockSQLPaths(manifest)) > 0
}

// ListFrontendPagePaths returns the frontend page source paths for a source plugin manifest.
func (s *serviceImpl) ListFrontendPagePaths(manifest *Manifest) []string {
	if embeddedFiles := GetSourcePluginEmbeddedFiles(manifest); embeddedFiles != nil {
		return pluginfs.DiscoverVuePathsFromFS(embeddedFiles, "frontend/pages")
	}
	if manifest == nil {
		return []string{}
	}
	return s.DiscoverPagePaths(manifest.RootDir)
}

// ListFrontendSlotPaths returns the frontend slot source paths for a source plugin manifest.
func (s *serviceImpl) ListFrontendSlotPaths(manifest *Manifest) []string {
	if embeddedFiles := GetSourcePluginEmbeddedFiles(manifest); embeddedFiles != nil {
		return pluginfs.DiscoverVuePathsFromFS(embeddedFiles, "frontend/slots")
	}
	if manifest == nil {
		return []string{}
	}
	return s.DiscoverSlotPaths(manifest.RootDir)
}
