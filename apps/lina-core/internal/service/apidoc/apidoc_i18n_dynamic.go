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
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/pkg/logger"
	"lina-core/pkg/pluginbridge"
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
		pluginBundle, loadErr := s.loadOpenAPIDynamicPluginBundle(ctx, release.PackagePath, locale)
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
		mergeOpenAPIPluginMessageCatalog(ctx, bundle, release.PluginId, pluginBundle)
	}
	return bundle
}

// loadOpenAPIWorkspacePluginBundles reads plugin-owned apidoc resources from
// the local monorepo source tree when it is available. Source-plugin embedded
// files and dynamic release artifacts remain the production discovery paths;
// this fallback keeps development and tests aligned with plugin-owned files.
func loadOpenAPIWorkspacePluginBundles(ctx context.Context, locale string) map[string]string {
	bundle := make(map[string]string)
	workingDir, err := os.Getwd()
	if err != nil {
		return bundle
	}
	repoRoot, err := findRepoRootForOpenAPIDynamicPlugin(workingDir)
	if err != nil {
		return bundle
	}
	pluginsRoot := filepath.Join(repoRoot, "apps", "lina-plugins")
	entries, err := os.ReadDir(pluginsRoot)
	if err != nil {
		return bundle
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		bundlePath := filepath.Join(pluginsRoot, entry.Name(), filepath.FromSlash(openAPIPluginI18nDir), locale+".json")
		content, readErr := os.ReadFile(bundlePath)
		if readErr != nil || len(content) == 0 {
			continue
		}
		pluginBundle := make(map[string]string)
		if unmarshalErr := json.Unmarshal(content, &pluginBundle); unmarshalErr != nil {
			logger.Warningf(ctx, "parse workspace plugin apidoc i18n bundle failed path=%s err=%v", bundlePath, unmarshalErr)
			continue
		}
		mergeOpenAPIPluginMessageCatalog(ctx, bundle, entry.Name(), pluginBundle)
	}
	return bundle
}

// listOpenAPIEnabledDynamicPluginReleases returns active release rows for
// enabled dynamic plugins so the apidoc service can read plugin-owned resources.
func listOpenAPIEnabledDynamicPluginReleases(ctx context.Context) ([]*entity.SysPluginRelease, error) {
	var plugins []*entity.SysPlugin
	if err := dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{
			Type:      openAPIDynamicPluginType,
			Installed: openAPIDynamicPluginInstalledYes,
			Status:    openAPIDynamicPluginStatusEnabled,
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
		release, err := getOpenAPIEnabledDynamicPluginRelease(ctx, plugin)
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

// getOpenAPIEnabledDynamicPluginRelease resolves the active release row for one
// enabled dynamic plugin.
func getOpenAPIEnabledDynamicPluginRelease(ctx context.Context, plugin *entity.SysPlugin) (*entity.SysPluginRelease, error) {
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
			Status:   openAPIDynamicPluginReleaseStatusActive,
		}).
		OrderDesc(dao.SysPluginRelease.Columns().Id).
		Scan(&release); err != nil {
		return nil, err
	}
	return release, nil
}

// loadOpenAPIDynamicPluginBundle reads one active dynamic-plugin artifact and
// returns the matching apidoc locale catalog.
func (s *serviceImpl) loadOpenAPIDynamicPluginBundle(ctx context.Context, packagePath string, locale string) (map[string]string, error) {
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
	bundle := make(map[string]string)
	for _, asset := range assets {
		if asset == nil || normalizeOpenAPILocale(asset.Locale) != locale {
			continue
		}
		assetBundle := make(map[string]string)
		if err = json.Unmarshal([]byte(asset.Content), &assetBundle); err != nil {
			return nil, gerror.Wrap(err, "parse dynamic plugin apidoc i18n asset failed")
		}
		mergeOpenAPIMessageCatalog(bundle, assetBundle)
	}
	return bundle, nil
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
	sections, err := parseOpenAPIWasmCustomSections(content)
	if err != nil {
		return nil, err
	}
	sectionContent, ok := sections[pluginbridge.WasmSectionAPIDocI18NAssets]
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

// parseOpenAPIWasmCustomSections extracts wasm custom sections by name.
func parseOpenAPIWasmCustomSections(content []byte) (map[string][]byte, error) {
	if len(content) < 8 {
		return nil, gerror.New("wasm file is too short")
	}
	if string(content[:4]) != "\x00asm" {
		return nil, gerror.New("invalid wasm header")
	}
	if content[4] != 0x01 || content[5] != 0x00 || content[6] != 0x00 || content[7] != 0x00 {
		return nil, gerror.New("invalid wasm version")
	}

	sections := make(map[string][]byte)
	cursor := 8
	for cursor < len(content) {
		sectionID := content[cursor]
		cursor++

		sectionSize, nextCursor, err := readOpenAPIWasmULEB128(content, cursor)
		if err != nil {
			return nil, err
		}
		cursor = nextCursor

		end := cursor + int(sectionSize)
		if end > len(content) {
			return nil, gerror.New("wasm section length exceeds content")
		}
		if sectionID == 0 {
			nameLength, nameCursor, err := readOpenAPIWasmULEB128(content, cursor)
			if err != nil {
				return nil, err
			}
			nameEnd := nameCursor + int(nameLength)
			if nameEnd > end {
				return nil, gerror.New("wasm custom section name exceeds content")
			}
			sectionName := string(content[nameCursor:nameEnd])
			sectionPayload := make([]byte, end-nameEnd)
			copy(sectionPayload, content[nameEnd:end])
			sections[sectionName] = sectionPayload
		}

		cursor = end
	}
	return sections, nil
}

// readOpenAPIWasmULEB128 decodes one unsigned LEB128 integer from a wasm byte stream.
func readOpenAPIWasmULEB128(content []byte, start int) (uint32, int, error) {
	var (
		value uint32
		shift uint
	)

	cursor := start
	for {
		if cursor >= len(content) {
			return 0, cursor, gerror.New("wasm ULEB128 data exceeds content")
		}
		current := content[cursor]
		cursor++

		value |= uint32(current&0x7f) << shift
		if current&0x80 == 0 {
			return value, cursor, nil
		}

		shift += 7
		if shift >= 32 {
			return 0, cursor, gerror.New("wasm ULEB128 value is too large")
		}
	}
}
