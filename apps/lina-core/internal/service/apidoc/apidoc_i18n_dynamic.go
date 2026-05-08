// This file loads API-documentation i18n assets from enabled dynamic plugin
// release artifacts so runtime extension routes can be localized without
// coupling plugin-owned translations into the host apidoc bundle.

package apidoc

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/dao"
	"lina-core/internal/model/entity"
	"lina-core/pkg/i18nresource"
	"lina-core/pkg/logger"
	bridgeartifact "lina-core/pkg/pluginbridge/artifact"
)

const (
	// openAPIDynamicPluginType identifies dynamic plugins in sys_plugin.
	openAPIDynamicPluginType = "dynamic"
	// openAPIDynamicPluginInstalledYes marks one plugin registry row as installed.
	openAPIDynamicPluginInstalledYes = 1
	// openAPIDynamicPluginStatusEnabled marks one plugin registry row as enabled.
	openAPIDynamicPluginStatusEnabled = 1
	// openAPIDynamicPluginReleaseStatusActive marks one release row as active.
	openAPIDynamicPluginReleaseStatusActive = "active"
)

// openAPIDynamicStorageConfigProvider exposes the dynamic plugin artifact root
// without forcing narrow apidoc tests to implement the full config service.
type openAPIDynamicStorageConfigProvider interface {
	// GetPluginDynamicStoragePath returns the configured dynamic plugin storage root.
	GetPluginDynamicStoragePath(ctx context.Context) string
}

// openAPIDynamicPluginI18NAsset stores one apidoc locale snapshot embedded in
// a dynamic plugin artifact.
type openAPIDynamicPluginI18NAsset struct {
	Locale  string `json:"locale"`
	Content string `json:"content"`
}

// loadOpenAPIDynamicPluginBundles loads enabled dynamic-plugin apidoc
// translations for one locale from the active release artifact custom sections.
func (s *serviceImpl) loadOpenAPIDynamicPluginBundles(ctx context.Context, locale string) map[string]string {
	bundle := make(map[string]string)
	defer func() {
		if recovered := recover(); recovered != nil {
			logger.Warningf(ctx, "load dynamic plugin apidoc i18n bundle panic locale=%s err=%v", locale, recovered)
		}
	}()

	releases, err := listOpenAPIEnabledDynamicPluginReleases(ctx)
	if err != nil {
		logger.Warningf(ctx, "load dynamic plugin apidoc i18n releases failed locale=%s err=%v", locale, err)
		return bundle
	}
	for _, release := range releases {
		if release == nil {
			continue
		}
		pluginID := strings.TrimSpace(release.PluginId)
		if pluginID == "" {
			continue
		}
		pluginBundle, loadErr := s.loadOpenAPIDynamicPluginBundle(ctx, pluginID, release.PackagePath, locale)
		if loadErr != nil {
			logger.Warningf(
				ctx,
				"load dynamic plugin apidoc i18n assets failed plugin=%s release=%s err=%v",
				release.PluginId,
				release.ReleaseVersion,
				loadErr,
			)
			continue
		}
		mergeOpenAPIPluginMessageCatalog(ctx, bundle, pluginID, pluginBundle)
	}
	return bundle
}

// listOpenAPIEnabledDynamicPluginReleases returns active release rows for
// enabled dynamic plugins so the apidoc service can read plugin-owned resources.
func listOpenAPIEnabledDynamicPluginReleases(ctx context.Context) ([]*entity.SysPluginRelease, error) {
	var plugins []*entity.SysPlugin
	if err := dao.SysPlugin.Ctx(ctx).
		OrderAsc(dao.SysPlugin.Columns().PluginId).
		Scan(&plugins); err != nil {
		return nil, err
	}

	var allReleases []*entity.SysPluginRelease
	if err := dao.SysPluginRelease.Ctx(ctx).Scan(&allReleases); err != nil {
		return nil, err
	}
	releasesByID, activeReleasesByPluginID := buildOpenAPIReleaseIndexes(allReleases)

	releases := make([]*entity.SysPluginRelease, 0, len(plugins))
	for _, plugin := range plugins {
		if plugin == nil || strings.TrimSpace(plugin.PluginId) == "" {
			continue
		}
		if strings.TrimSpace(plugin.Type) != openAPIDynamicPluginType ||
			plugin.Installed != openAPIDynamicPluginInstalledYes ||
			plugin.Status != openAPIDynamicPluginStatusEnabled {
			continue
		}
		release := releasesByID[plugin.ReleaseId]
		if release == nil || strings.TrimSpace(release.PackagePath) == "" {
			release = activeReleasesByPluginID[strings.TrimSpace(plugin.PluginId)]
		}
		if release != nil && strings.TrimSpace(release.PackagePath) != "" {
			releases = append(releases, release)
		}
	}
	return releases, nil
}

// buildOpenAPIReleaseIndexes prepares release lookup maps from one full-table
// snapshot so apidoc startup loading does not issue one release query per plugin.
func buildOpenAPIReleaseIndexes(
	releases []*entity.SysPluginRelease,
) (map[int]*entity.SysPluginRelease, map[string]*entity.SysPluginRelease) {
	releasesByID := make(map[int]*entity.SysPluginRelease, len(releases))
	activeReleasesByPluginID := make(map[string]*entity.SysPluginRelease)
	for _, release := range releases {
		if release == nil {
			continue
		}
		releasesByID[release.Id] = release
		if strings.TrimSpace(release.Status) != openAPIDynamicPluginReleaseStatusActive {
			continue
		}
		pluginID := strings.TrimSpace(release.PluginId)
		if pluginID == "" {
			continue
		}
		if existing := activeReleasesByPluginID[pluginID]; existing == nil || release.Id > existing.Id {
			activeReleasesByPluginID[pluginID] = release
		}
	}
	return releasesByID, activeReleasesByPluginID
}

// loadOpenAPIDynamicPluginBundle reads one active dynamic-plugin artifact and
// returns the matching apidoc locale catalog.
func (s *serviceImpl) loadOpenAPIDynamicPluginBundle(ctx context.Context, pluginID string, packagePath string, locale string) (map[string]string, error) {
	absolutePath, err := s.resolveOpenAPIDynamicPluginPackagePath(ctx, packagePath)
	if err != nil {
		return nil, err
	}
	content, err := os.ReadFile(absolutePath)
	if err != nil {
		return nil, err
	}

	assets, err := parseOpenAPIDynamicPluginI18NAssets(content)
	if err != nil {
		return nil, err
	}
	localeAssets := make([]i18nresource.LocaleAsset, 0, len(assets))
	for _, asset := range assets {
		if asset == nil {
			continue
		}
		localeAssets = append(localeAssets, i18nresource.LocaleAsset{
			Locale:  asset.Locale,
			Content: asset.Content,
		})
	}
	pluginBundles := openAPIResourceLoader(i18nresource.ResourceLoader{
		PluginScope: i18nresource.PluginScopeRestrictedToPluginNamespace,
	}).LoadDynamicPluginBundles(ctx, locale, []i18nresource.ReleaseRef{
		{
			PluginID: pluginID,
			Assets:   localeAssets,
		},
	})
	if bundle := pluginBundles[pluginID]; len(bundle) > 0 {
		return bundle, nil
	}
	return map[string]string{}, nil
}

// resolveOpenAPIDynamicPluginPackagePath converts a release package path into
// an absolute filesystem path using the configured dynamic plugin storage root.
func (s *serviceImpl) resolveOpenAPIDynamicPluginPackagePath(ctx context.Context, packagePath string) (string, error) {
	trimmedPath := strings.TrimSpace(packagePath)
	if trimmedPath == "" {
		return "", gerror.New("dynamic plugin release package_path is empty")
	}
	if filepath.IsAbs(trimmedPath) {
		return filepath.Clean(trimmedPath), nil
	}

	configProvider, ok := s.configSvc.(openAPIDynamicStorageConfigProvider)
	if !ok || configProvider == nil {
		return filepath.Clean(trimmedPath), nil
	}
	storagePath := strings.TrimSpace(configProvider.GetPluginDynamicStoragePath(ctx))
	if storagePath == "" {
		return filepath.Clean(trimmedPath), nil
	}
	if filepath.IsAbs(storagePath) {
		return filepath.Clean(filepath.Join(storagePath, filepath.FromSlash(trimmedPath))), nil
	}

	workingDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	storageRoot := resolveOpenAPIDynamicPluginStorageRoot(workingDir, storagePath)
	return filepath.Clean(filepath.Join(storageRoot, filepath.FromSlash(trimmedPath))), nil
}

// resolveOpenAPIDynamicPluginStorageRoot resolves relative dynamic-plugin
// storage paths against the repository root when the backend runs from a
// subdirectory such as apps/lina-core.
func resolveOpenAPIDynamicPluginStorageRoot(workingDir string, storagePath string) string {
	trimmedStoragePath := strings.TrimSpace(storagePath)
	if trimmedStoragePath == "" {
		return filepath.Clean(workingDir)
	}
	if filepath.IsAbs(trimmedStoragePath) {
		return filepath.Clean(trimmedStoragePath)
	}

	candidates := make([]string, 0, 4)
	if repoRoot, err := findRepoRootForOpenAPIDynamicPlugin(workingDir); err == nil {
		candidates = append(candidates, filepath.Join(repoRoot, trimmedStoragePath))
	}
	candidates = append(
		candidates,
		filepath.Join(workingDir, trimmedStoragePath),
		filepath.Join(workingDir, "..", trimmedStoragePath),
		filepath.Join(workingDir, "..", "..", trimmedStoragePath),
	)
	for _, candidate := range candidates {
		cleanPath := filepath.Clean(candidate)
		if _, err := os.Stat(cleanPath); err == nil {
			return cleanPath
		}
	}
	return filepath.Clean(candidates[0])
}

// findRepoRootForOpenAPIDynamicPlugin walks upward until it finds the go.work
// marker used by the monorepo root.
func findRepoRootForOpenAPIDynamicPlugin(startDir string) (string, error) {
	currentDir, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}
	for {
		if _, statErr := os.Stat(filepath.Join(currentDir, "go.work")); statErr == nil {
			return currentDir, nil
		}
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			break
		}
		currentDir = parentDir
	}
	return "", gerror.Newf("repository root not found: %s", startDir)
}

// parseOpenAPIDynamicPluginI18NAssets extracts apidoc i18n asset snapshots from
// one dynamic plugin wasm artifact.
func parseOpenAPIDynamicPluginI18NAssets(content []byte) ([]*openAPIDynamicPluginI18NAsset, error) {
	sectionContent, ok, err := bridgeartifact.ReadCustomSection(content, bridgeartifact.WasmSectionAPIDocI18NAssets)
	if err != nil {
		return nil, err
	}
	if !ok {
		return []*openAPIDynamicPluginI18NAsset{}, nil
	}

	assets := make([]*openAPIDynamicPluginI18NAsset, 0)
	if err = json.Unmarshal(sectionContent, &assets); err != nil {
		return nil, gerror.Wrap(err, "parse dynamic plugin apidoc i18n custom section failed")
	}
	for _, asset := range assets {
		if asset == nil {
			return nil, gerror.New("dynamic plugin apidoc i18n custom section contains nil asset")
		}
		asset.Locale = normalizeOpenAPILocale(asset.Locale)
		asset.Content = strings.TrimSpace(asset.Content)
		if asset.Locale == "" || asset.Content == "" {
			return nil, gerror.New("dynamic plugin apidoc i18n custom section misses locale or content")
		}
	}
	return assets, nil
}
