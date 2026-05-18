// This file implements filesystem and plugin bundle loading for i18n resources.

package i18nresource

import (
	"context"
	"io/fs"
	"path"
	"sort"
	"strings"

	"lina-core/pkg/logger"
)

// LoadHostBundle loads one host-owned locale bundle from HostFS.
func (l ResourceLoader) LoadHostBundle(ctx context.Context, locale string) map[string]string {
	return l.loadFilesystemBundle(ctx, l.HostFS, locale)
}

// LoadSourcePluginBundles loads locale bundles from registered source plugins.
func (l ResourceLoader) LoadSourcePluginBundles(ctx context.Context, locale string) map[string]map[string]string {
	bundles := make(map[string]map[string]string)
	rawSourcePlugins := l.sourcePlugins()
	if len(rawSourcePlugins) == 0 {
		return bundles
	}

	sourcePlugins := make([]SourcePlugin, 0, len(rawSourcePlugins))
	for _, sourcePlugin := range rawSourcePlugins {
		if sourcePlugin == nil || sourcePlugin.GetEmbeddedFiles() == nil || strings.TrimSpace(sourcePlugin.ID()) == "" {
			continue
		}
		sourcePlugins = append(sourcePlugins, sourcePlugin)
	}
	sort.Slice(sourcePlugins, func(i, j int) bool {
		return sourcePlugins[i].ID() < sourcePlugins[j].ID()
	})
	for _, sourcePlugin := range sourcePlugins {
		pluginID := strings.TrimSpace(sourcePlugin.ID())
		bundle := l.filterPluginBundle(ctx, pluginID, l.loadFilesystemBundle(ctx, sourcePlugin.GetEmbeddedFiles(), locale))
		if len(bundle) > 0 {
			bundles[pluginID] = bundle
		}
	}
	return bundles
}

// LoadDynamicPluginBundles loads locale bundles from already-extracted dynamic plugin release assets.
func (l ResourceLoader) LoadDynamicPluginBundles(ctx context.Context, locale string, releases []ReleaseRef) map[string]map[string]string {
	bundles := make(map[string]map[string]string)
	normalizedLocale := strings.TrimSpace(locale)
	for _, release := range releases {
		pluginID := strings.TrimSpace(release.PluginID)
		if pluginID == "" {
			continue
		}
		pluginBundle := make(map[string]string)
		for _, asset := range release.Assets {
			if strings.TrimSpace(asset.Locale) != normalizedLocale || strings.TrimSpace(asset.Content) == "" {
				continue
			}
			assetBundle, err := ParseCatalog([]byte(asset.Content), l.valueMode())
			if err != nil {
				logger.Warningf(ctx, "parse dynamic plugin i18n resource failed plugin=%s locale=%s err=%v", pluginID, locale, err)
				continue
			}
			mergeCatalog(pluginBundle, assetBundle)
		}
		pluginBundle = l.filterPluginBundle(ctx, pluginID, pluginBundle)
		if len(pluginBundle) > 0 {
			bundles[pluginID] = pluginBundle
		}
	}
	return bundles
}

// loadFilesystemBundle loads one locale bundle from the configured locale directory.
func (l ResourceLoader) loadFilesystemBundle(ctx context.Context, filesystem fs.FS, locale string) map[string]string {
	if filesystem == nil {
		return map[string]string{}
	}
	trimmedLocale := strings.TrimSpace(locale)
	dir := path.Join(l.subdir(), trimmedLocale)
	if localeSubdir := l.localeSubdir(); localeSubdir != "" {
		dir = path.Join(dir, localeSubdir)
	}
	return l.loadFilesystemBundleDirectory(ctx, filesystem, dir, l.Recursive)
}

// loadFilesystemBundleDirectory loads JSON files from one directory in
// deterministic path order. When recursive is false only direct children are
// loaded, which keeps runtime bundles from reading apidoc resources.
func (l ResourceLoader) loadFilesystemBundleDirectory(ctx context.Context, filesystem fs.FS, dir string, recursive bool) map[string]string {
	entries := make([]string, 0)
	if recursive {
		if err := fs.WalkDir(filesystem, dir, func(filePath string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				if filePath == dir {
					return nil
				}
				return walkErr
			}
			if entry == nil || entry.IsDir() || !strings.HasSuffix(filePath, ".json") {
				return nil
			}
			entries = append(entries, filePath)
			return nil
		}); err != nil {
			logger.Warningf(ctx, "scan i18n resource directory failed dir=%s err=%v", dir, err)
			return map[string]string{}
		}
	} else {
		dirEntries, err := fs.ReadDir(filesystem, dir)
		if err != nil {
			return map[string]string{}
		}
		for _, entry := range dirEntries {
			if entry == nil || entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
				continue
			}
			entries = append(entries, path.Join(dir, entry.Name()))
		}
	}
	sort.Strings(entries)
	bundle := make(map[string]string)
	for _, entryPath := range entries {
		mergeCatalog(bundle, l.loadFilesystemBundleFile(ctx, filesystem, entryPath))
	}
	return bundle
}

// loadFilesystemBundleFile loads one JSON file from a filesystem.
func (l ResourceLoader) loadFilesystemBundleFile(ctx context.Context, filesystem fs.FS, filePath string) map[string]string {
	content, err := fs.ReadFile(filesystem, filePath)
	if err != nil || len(content) == 0 {
		return map[string]string{}
	}
	bundle, err := ParseCatalog(content, l.valueMode())
	if err != nil {
		logger.Warningf(ctx, "parse i18n resource failed path=%s err=%v", filePath, err)
		return map[string]string{}
	}
	return l.filterCatalog(ctx, "", bundle)
}

// sourcePlugins returns the registered source plugins from the configured provider.
func (l ResourceLoader) sourcePlugins() []SourcePlugin {
	if l.SourcePlugins == nil {
		return []SourcePlugin{}
	}
	return l.SourcePlugins()
}
