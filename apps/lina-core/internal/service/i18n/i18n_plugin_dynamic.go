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
	"lina-core/pkg/pluginbridge"
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

// listEnabledDynamicPluginReleases returns active release rows for plugins that are currently enabled.
func (s *serviceImpl) listEnabledDynamicPluginReleases(ctx context.Context) ([]*entity.SysPluginRelease, error) {
	var plugins []*entity.SysPlugin
	if err := dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{
			Type:      dynamicPluginType,
			Installed: dynamicPluginInstalledYes,
			Status:    dynamicPluginStatusEnabled,
		}).
		OrderAsc(dao.SysPlugin.Columns().PluginId).
		Scan(&plugins); err != nil {
		return nil, err
	}

	releases := make([]*entity.SysPluginRelease, 0, len(plugins))
	for _, plugin := range plugins {
		if plugin == nil || strings.TrimSpace(plugin.PluginId) == "" {
			continue
		}
		release, err := s.getEnabledDynamicPluginRelease(ctx, plugin)
		if err != nil {
			return nil, err
		}
		if release == nil || strings.TrimSpace(release.PackagePath) == "" {
			continue
		}
		releases = append(releases, release)
	}
	return releases, nil
}

// getEnabledDynamicPluginRelease resolves the active release row for one enabled dynamic plugin.
func (s *serviceImpl) getEnabledDynamicPluginRelease(ctx context.Context, plugin *entity.SysPlugin) (*entity.SysPluginRelease, error) {
	if plugin == nil {
		return nil, nil
	}

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
			PluginId: strings.TrimSpace(plugin.PluginId),
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

	sectionContent, ok, err := pluginbridge.ReadCustomSection(content, pluginbridge.WasmSectionI18NAssets)
	if err != nil {
		return nil, err
	}
	if !ok {
		return []*dynamicPluginI18NAsset{}, nil
	}

	assets := make([]*dynamicPluginI18NAsset, 0)
	if err = json.Unmarshal(sectionContent, &assets); err != nil {
		return nil, gerror.Wrap(err, "解析动态插件国际化自定义节失败")
	}
	for _, asset := range assets {
		if asset == nil {
			return nil, gerror.New("动态插件国际化自定义节存在空项")
		}
		asset.Locale = normalizeLocale(asset.Locale)
		asset.Content = strings.TrimSpace(asset.Content)
		if asset.Locale == "" || asset.Content == "" {
			return nil, gerror.New("动态插件国际化自定义节缺少 locale 或 content")
		}
	}
	return assets, nil
}

// resolveDynamicPluginPackagePath converts a release package path into an absolute filesystem path.
func (s *serviceImpl) resolveDynamicPluginPackagePath(ctx context.Context, packagePath string) (string, error) {
	trimmedPath := strings.TrimSpace(packagePath)
	if trimmedPath == "" {
		return "", gerror.New("动态插件 release package_path 不能为空")
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
	return "", gerror.Newf("未找到仓库根目录: %s", startDir)
}
