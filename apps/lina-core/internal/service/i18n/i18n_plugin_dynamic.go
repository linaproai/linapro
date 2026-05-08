// This file loads runtime i18n assets from enabled dynamic plugin release
// artifacts so plugin lifecycle changes participate in host translation aggregation.

package i18n

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/pkg/i18nresource"
	"lina-core/pkg/logger"
	bridgeartifact "lina-core/pkg/pluginbridge/artifact"
)

const (
	// dynamicPluginType identifies dynamic plugins in sys_plugin.
	dynamicPluginType = "dynamic"
	// dynamicPluginInstalledYes marks one plugin registry row as installed.
	dynamicPluginInstalledYes = 1
	// dynamicPluginStatusEnabled marks one plugin registry row as enabled.
	dynamicPluginStatusEnabled = 1
	// dynamicPluginReleaseStatusActive marks one release row as active.
	dynamicPluginReleaseStatusActive = "active"
)

// dynamicPluginI18NAsset stores one locale snapshot embedded in a dynamic plugin artifact.
type dynamicPluginI18NAsset struct {
	Locale  string `json:"locale"`
	Content string `json:"content"`
}

// dynamicPluginI18NArtifactCandidate describes one artifact path that may
// contain source-text translations for an inactive dynamic plugin.
type dynamicPluginI18NArtifactCandidate struct {
	label       string
	packagePath string
}

// loadDynamicPluginLocaleBundles loads enabled dynamic-plugin translations for
// one locale, returning a per-plugin map. The cache stores each plugin entry
// separately so a single plugin lifecycle change can invalidate only its slice.
func (s *serviceImpl) loadDynamicPluginLocaleBundles(ctx context.Context, locale string) map[string]map[string]string {
	resolvedLocale := s.ResolveLocale(ctx, locale)
	bundles := make(map[string]map[string]string)
	defer func() {
		if recovered := recover(); recovered != nil {
			logger.Warningf(ctx, "load dynamic plugin i18n bundle panic locale=%s err=%v", resolvedLocale, recovered)
		}
	}()

	releases, err := s.listEnabledDynamicPluginReleases(ctx)
	if err != nil {
		logger.Warningf(ctx, "load enabled dynamic plugin i18n releases failed locale=%s err=%v", resolvedLocale, err)
		return bundles
	}

	releaseRefs := make([]i18nresource.ReleaseRef, 0, len(releases))
	for _, release := range releases {
		if release == nil {
			continue
		}
		assets, loadErr := s.readDynamicPluginI18NAssets(ctx, release.PackagePath)
		if loadErr != nil {
			logger.Warningf(
				ctx,
				"load dynamic plugin i18n assets failed plugin=%s release=%s err=%v",
				release.PluginId,
				release.ReleaseVersion,
				loadErr,
			)
			continue
		}
		pluginID := strings.TrimSpace(release.PluginId)
		if pluginID == "" {
			continue
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
		if len(localeAssets) == 0 {
			continue
		}
		releaseRefs = append(releaseRefs, i18nresource.ReleaseRef{
			PluginID: pluginID,
			Assets:   localeAssets,
		})
	}
	return i18nresource.ResourceLoader{
		PluginScope: i18nresource.PluginScopeOpen,
		ValueMode:   i18nresource.ValueModeStringifyScalars,
	}.LoadDynamicPluginBundles(ctx, resolvedLocale, releaseRefs)
}

// loadDynamicPluginLocaleBundle reloads one dynamic plugin's runtime i18n
// bundle after a plugin-scoped lifecycle invalidation.
func (s *serviceImpl) loadDynamicPluginLocaleBundle(ctx context.Context, locale string, pluginID string) map[string]string {
	resolvedLocale := s.ResolveLocale(ctx, locale)
	trimmedPluginID := strings.TrimSpace(pluginID)
	if trimmedPluginID == "" {
		return map[string]string{}
	}

	var plugin *entity.SysPlugin
	if err := dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{
			PluginId:  trimmedPluginID,
			Type:      dynamicPluginType,
			Installed: dynamicPluginInstalledYes,
			Status:    dynamicPluginStatusEnabled,
		}).
		Scan(&plugin); err != nil {
		logger.Warningf(ctx, "load dynamic plugin i18n plugin row failed plugin=%s locale=%s err=%v", trimmedPluginID, resolvedLocale, err)
		return map[string]string{}
	}
	if plugin == nil {
		return map[string]string{}
	}

	release, err := s.getEnabledDynamicPluginRelease(ctx, plugin)
	if err != nil {
		logger.Warningf(ctx, "load dynamic plugin i18n release failed plugin=%s locale=%s err=%v", trimmedPluginID, resolvedLocale, err)
		return map[string]string{}
	}
	if release == nil || strings.TrimSpace(release.PackagePath) == "" {
		return map[string]string{}
	}

	bundle, err := s.loadDynamicPluginLocaleBundleFromRelease(ctx, resolvedLocale, trimmedPluginID, release)
	if err != nil {
		logger.Warningf(ctx, "load dynamic plugin i18n assets failed plugin=%s locale=%s err=%v", trimmedPluginID, resolvedLocale, err)
		return map[string]string{}
	}
	return bundle
}

// loadDynamicPluginLocaleBundleFromRelease loads one plugin locale bundle from
// an already resolved release row.
func (s *serviceImpl) loadDynamicPluginLocaleBundleFromRelease(
	ctx context.Context,
	resolvedLocale string,
	pluginID string,
	release *entity.SysPluginRelease,
) (map[string]string, error) {
	if release == nil || strings.TrimSpace(release.PackagePath) == "" {
		return map[string]string{}, nil
	}

	return s.loadDynamicPluginLocaleBundleFromPackagePath(ctx, resolvedLocale, pluginID, release.PackagePath)
}

// loadDynamicPluginLocaleBundleFromPackagePath loads one plugin locale bundle
// from a resolved dynamic artifact path.
func (s *serviceImpl) loadDynamicPluginLocaleBundleFromPackagePath(
	ctx context.Context,
	resolvedLocale string,
	pluginID string,
	packagePath string,
) (map[string]string, error) {
	if strings.TrimSpace(packagePath) == "" {
		return map[string]string{}, nil
	}

	assets, err := s.readDynamicPluginI18NAssets(ctx, packagePath)
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
	if len(localeAssets) == 0 {
		return map[string]string{}, nil
	}

	bundles := i18nresource.ResourceLoader{
		PluginScope: i18nresource.PluginScopeOpen,
		ValueMode:   i18nresource.ValueModeStringifyScalars,
	}.LoadDynamicPluginBundles(ctx, resolvedLocale, []i18nresource.ReleaseRef{
		{
			PluginID: pluginID,
			Assets:   localeAssets,
		},
	})
	return bundles[pluginID], nil
}

// TranslateDynamicPluginSourceText resolves source-owned text from the latest
// dynamic-plugin release artifact without adding inactive plugin resources to
// the process-wide runtime bundle cache.
func (s *serviceImpl) TranslateDynamicPluginSourceText(ctx context.Context, pluginID string, key string, sourceText string) string {
	trimmedPluginID := strings.TrimSpace(pluginID)
	trimmedKey := strings.TrimSpace(key)
	if trimmedPluginID == "" || trimmedKey == "" {
		return sourceText
	}

	bundle, err := s.loadDynamicPluginReleaseLocaleBundle(ctx, s.GetLocale(ctx), trimmedPluginID)
	if err != nil {
		logger.Warningf(ctx, "load dynamic plugin source-text i18n failed plugin=%s err=%v", trimmedPluginID, err)
		return sourceText
	}
	if value := strings.TrimSpace(bundle[trimmedKey]); value != "" {
		return value
	}
	return sourceText
}

// loadDynamicPluginReleaseLocaleBundle loads one locale bundle from the latest
// known dynamic-plugin artifacts regardless of install or enable state.
func (s *serviceImpl) loadDynamicPluginReleaseLocaleBundle(ctx context.Context, locale string, pluginID string) (map[string]string, error) {
	candidates, err := s.listDynamicPluginI18NArtifactCandidates(ctx, pluginID)
	if err != nil {
		return map[string]string{}, err
	}

	resolvedLocale := s.ResolveLocale(ctx, locale)
	var firstErr error
	mergedBundle := make(map[string]string)
	for _, candidate := range candidates {
		bundle, loadErr := s.loadDynamicPluginLocaleBundleFromPackagePath(
			ctx,
			resolvedLocale,
			pluginID,
			candidate.packagePath,
		)
		if loadErr != nil {
			if firstErr == nil {
				firstErr = gerror.Wrapf(loadErr, "load %s", candidate.label)
			}
			continue
		}
		for key, value := range bundle {
			if _, exists := mergedBundle[key]; exists {
				continue
			}
			mergedBundle[key] = value
		}
	}
	if len(mergedBundle) > 0 {
		return mergedBundle, nil
	}
	if firstErr != nil {
		return map[string]string{}, firstErr
	}
	return map[string]string{}, nil
}

// listDynamicPluginI18NArtifactCandidates returns artifact paths that can
// provide pre-enable dynamic-plugin metadata translation, ordered from most
// authoritative to broader fallbacks.
func (s *serviceImpl) listDynamicPluginI18NArtifactCandidates(
	ctx context.Context,
	pluginID string,
) ([]dynamicPluginI18NArtifactCandidate, error) {
	trimmedPluginID := strings.TrimSpace(pluginID)
	if trimmedPluginID == "" {
		return nil, nil
	}

	candidates := make([]dynamicPluginI18NArtifactCandidate, 0, 4)
	seenPackagePaths := make(map[string]struct{})

	var plugin *entity.SysPlugin
	if err := dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{
			PluginId: trimmedPluginID,
			Type:     dynamicPluginType,
		}).
		Scan(&plugin); err != nil {
		return nil, err
	}
	if plugin != nil && plugin.ReleaseId > 0 {
		var release *entity.SysPluginRelease
		if err := dao.SysPluginRelease.Ctx(ctx).
			Where(do.SysPluginRelease{Id: plugin.ReleaseId}).
			Scan(&release); err != nil {
			return nil, err
		}
		if release != nil && strings.TrimSpace(release.PackagePath) != "" {
			candidates = appendDynamicPluginI18NArtifactCandidate(
				candidates,
				seenPackagePaths,
				"dynamic plugin registry release",
				release.PackagePath,
			)
		}
	}

	var releases []*entity.SysPluginRelease
	err := dao.SysPluginRelease.Ctx(ctx).
		Where(do.SysPluginRelease{
			PluginId: trimmedPluginID,
			Type:     dynamicPluginType,
		}).
		OrderDesc(dao.SysPluginRelease.Columns().Id).
		Scan(&releases)
	if err != nil {
		return nil, err
	}
	for _, release := range releases {
		if release == nil || strings.TrimSpace(release.PackagePath) == "" {
			continue
		}
		candidates = appendDynamicPluginI18NArtifactCandidate(
			candidates,
			seenPackagePaths,
			"dynamic plugin release",
			release.PackagePath,
		)
	}

	candidates = s.appendDynamicPluginStagingI18NArtifactCandidate(ctx, trimmedPluginID, candidates, seenPackagePaths)
	return candidates, nil
}

// appendDynamicPluginI18NArtifactCandidate appends a non-empty artifact path
// while avoiding duplicate reads within one source-text lookup.
func appendDynamicPluginI18NArtifactCandidate(
	candidates []dynamicPluginI18NArtifactCandidate,
	seenPackagePaths map[string]struct{},
	label string,
	packagePath string,
) []dynamicPluginI18NArtifactCandidate {
	trimmedPackagePath := strings.TrimSpace(packagePath)
	if trimmedPackagePath == "" {
		return candidates
	}
	normalizedPackagePath := filepath.Clean(filepath.FromSlash(trimmedPackagePath))
	if _, ok := seenPackagePaths[normalizedPackagePath]; ok {
		return candidates
	}
	seenPackagePaths[normalizedPackagePath] = struct{}{}
	return append(candidates, dynamicPluginI18NArtifactCandidate{
		label:       label,
		packagePath: trimmedPackagePath,
	})
}

// appendDynamicPluginStagingI18NArtifactCandidate appends the current upload
// artifact when it exists in plugin.dynamic.storagePath.
func (s *serviceImpl) appendDynamicPluginStagingI18NArtifactCandidate(
	ctx context.Context,
	pluginID string,
	candidates []dynamicPluginI18NArtifactCandidate,
	seenPackagePaths map[string]struct{},
) []dynamicPluginI18NArtifactCandidate {
	stagingPackagePath := dynamicPluginStagingArtifactPackagePath(pluginID)
	absolutePath, err := s.resolveDynamicPluginPackagePath(ctx, stagingPackagePath)
	if err == nil {
		if _, statErr := os.Stat(absolutePath); statErr != nil && os.IsNotExist(statErr) {
			return candidates
		}
	}
	return appendDynamicPluginI18NArtifactCandidate(
		candidates,
		seenPackagePaths,
		"dynamic plugin staging artifact",
		stagingPackagePath,
	)
}

// dynamicPluginStagingArtifactPackagePath returns the flat upload artifact name
// used for dynamic plugin discovery before a release package is selected.
func dynamicPluginStagingArtifactPackagePath(pluginID string) string {
	return strings.TrimSpace(pluginID) + ".wasm"
}

// listEnabledDynamicPluginReleases returns active release rows for plugins that are currently enabled.
func (s *serviceImpl) listEnabledDynamicPluginReleases(ctx context.Context) ([]*entity.SysPluginRelease, error) {
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
	releasesByID, activeReleasesByPluginID := buildDynamicPluginReleaseIndexes(allReleases)

	releases := make([]*entity.SysPluginRelease, 0, len(plugins))
	for _, plugin := range plugins {
		if plugin == nil || strings.TrimSpace(plugin.PluginId) == "" {
			continue
		}
		if strings.TrimSpace(plugin.Type) != dynamicPluginType ||
			plugin.Installed != dynamicPluginInstalledYes ||
			plugin.Status != dynamicPluginStatusEnabled {
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

// buildDynamicPluginReleaseIndexes prepares release lookup maps from one
// full-table snapshot for runtime i18n startup bundle loading.
func buildDynamicPluginReleaseIndexes(
	releases []*entity.SysPluginRelease,
) (map[int]*entity.SysPluginRelease, map[string]*entity.SysPluginRelease) {
	releasesByID := make(map[int]*entity.SysPluginRelease, len(releases))
	activeReleasesByPluginID := make(map[string]*entity.SysPluginRelease)
	for _, release := range releases {
		if release == nil {
			continue
		}
		releasesByID[release.Id] = release
		if strings.TrimSpace(release.Status) != dynamicPluginReleaseStatusActive {
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

// getEnabledDynamicPluginRelease resolves the active release row for one enabled dynamic plugin.
func (s *serviceImpl) getEnabledDynamicPluginRelease(ctx context.Context, plugin *entity.SysPlugin) (*entity.SysPluginRelease, error) {
	if plugin == nil {
		return nil, nil
	}
	trimmedPluginID := strings.TrimSpace(plugin.PluginId)

	var release *entity.SysPluginRelease
	if plugin.ReleaseId > 0 {
		if err := dao.SysPluginRelease.Ctx(ctx).
			Where(do.SysPluginRelease{Id: plugin.ReleaseId}).
			Scan(&release); err != nil {
			return nil, err
		}
		if release != nil {
			return release, nil
		}
	}

	if err := dao.SysPluginRelease.Ctx(ctx).
		Where(do.SysPluginRelease{
			PluginId: trimmedPluginID,
			Status:   dynamicPluginReleaseStatusActive,
		}).
		OrderDesc(dao.SysPluginRelease.Columns().Id).
		Scan(&release); err != nil {
		return nil, err
	}
	return release, nil
}

// readDynamicPluginI18NAssets reads one dynamic plugin release artifact and restores its embedded i18n snapshots.
func (s *serviceImpl) readDynamicPluginI18NAssets(ctx context.Context, packagePath string) ([]*dynamicPluginI18NAsset, error) {
	absolutePath, err := s.resolveDynamicPluginPackagePath(ctx, packagePath)
	if err != nil {
		return nil, err
	}
	content, err := os.ReadFile(absolutePath)
	if err != nil {
		return nil, err
	}

	sectionContent, ok, err := bridgeartifact.ReadCustomSection(content, bridgeartifact.WasmSectionI18NAssets)
	if err != nil {
		return nil, err
	}
	if !ok {
		return []*dynamicPluginI18NAsset{}, nil
	}

	assets := make([]*dynamicPluginI18NAsset, 0)
	if err = json.Unmarshal(sectionContent, &assets); err != nil {
		return nil, gerror.Wrap(err, "parse dynamic plugin i18n custom section failed")
	}
	for _, asset := range assets {
		if asset == nil {
			return nil, gerror.New("dynamic plugin i18n custom section contains a nil item")
		}
		asset.Locale = normalizeLocale(asset.Locale)
		asset.Content = strings.TrimSpace(asset.Content)
		if asset.Locale == "" || asset.Content == "" {
			return nil, gerror.New("dynamic plugin i18n custom section is missing locale or content")
		}
	}
	return assets, nil
}

// resolveDynamicPluginPackagePath converts a release package path into an absolute filesystem path.
func (s *serviceImpl) resolveDynamicPluginPackagePath(ctx context.Context, packagePath string) (string, error) {
	trimmedPath := strings.TrimSpace(packagePath)
	if trimmedPath == "" {
		return "", gerror.New("dynamic plugin release package_path cannot be empty")
	}
	if filepath.IsAbs(trimmedPath) {
		return filepath.Clean(trimmedPath), nil
	}
	if s == nil || s.configSvc == nil {
		return filepath.Clean(trimmedPath), nil
	}
	storagePath := strings.TrimSpace(s.configSvc.GetPluginDynamicStoragePath(ctx))
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
	storageRoot := resolveDynamicPluginStorageRoot(workingDir, storagePath)
	return filepath.Clean(filepath.Join(storageRoot, filepath.FromSlash(trimmedPath))), nil
}

// resolveDynamicPluginStorageRoot resolves the configured dynamic-plugin
// storage root. Relative storage paths prefer the repository root when the
// backend is started from a subdirectory such as apps/lina-core.
func resolveDynamicPluginStorageRoot(workingDir string, storagePath string) string {
	trimmedStoragePath := strings.TrimSpace(storagePath)
	if trimmedStoragePath == "" {
		return filepath.Clean(workingDir)
	}
	if filepath.IsAbs(trimmedStoragePath) {
		return filepath.Clean(trimmedStoragePath)
	}

	candidates := make([]string, 0, 4)
	if repoRoot, err := findRepoRootForDynamicPluginI18N(workingDir); err == nil {
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

// findRepoRootForDynamicPluginI18N walks upward until it finds the repository
// go.work marker so relative runtime storage paths can be anchored consistently.
func findRepoRootForDynamicPluginI18N(startDir string) (string, error) {
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
	return "", gerror.Newf("repository root was not found: %s", startDir)
}
