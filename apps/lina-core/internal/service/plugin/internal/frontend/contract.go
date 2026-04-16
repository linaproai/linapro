// This file validates how dynamic plugin menus consume hosted frontend assets.
// The host serves these assets from WASM-backed in-memory bundles, and enable-time
// validation prevents broken runtime menus from entering the router.

package frontend

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/dao"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
)

const (
	// hostedAssetURLPrefix is the public URL prefix for all plugin-hosted assets.
	hostedAssetURLPrefix = "/plugin-assets/"
	// dynamicPageComponentPath is the frontend component used for embedded-mount plugin pages.
	dynamicPageComponentPath = "system/plugin/dynamic-page"
	// DynamicPageComponentPath is the exported form of dynamicPageComponentPath for cross-package access.
	DynamicPageComponentPath = dynamicPageComponentPath
	// menuQueryKeyAccessMode is the query parameter key that controls plugin page access mode.
	menuQueryKeyAccessMode = "pluginAccessMode"
	// accessModeEmbedded is the access mode for ESM-mounted plugin pages.
	accessModeEmbedded = "embedded-mount"
	// embeddedJSExtension is the allowed ESM entry extension.
	embeddedJSExtension = ".js"
	// embeddedMJSExtension is the allowed ESM module entry extension.
	embeddedMJSExtension = ".mjs"
)

// ValidateRuntimeFrontendMenuBindings verifies that dynamic plugin menus only reference
// hosted assets that exist in the plugin's in-memory bundle.
func (s *serviceImpl) ValidateRuntimeFrontendMenuBindings(ctx context.Context, manifest *catalog.Manifest) error {
	if manifest == nil || catalog.NormalizeType(manifest.Type) != catalog.TypeDynamic {
		return nil
	}

	menus, err := s.listPluginOwnedMenus(ctx, manifest.ID)
	if err != nil {
		return err
	}
	return s.validateHostedMenuBindings(ctx, manifest, menus)
}

func (s *serviceImpl) listPluginOwnedMenus(ctx context.Context, pluginID string) ([]*entity.SysMenu, error) {
	columns := dao.SysMenu.Columns()
	prefixPattern := catalog.MenuKeyPrefix + pluginID + ":%"
	remarkPattern := catalog.MenuRemarkPrefix + pluginID + "%"

	var menus []*entity.SysMenu
	if err := dao.SysMenu.Ctx(ctx).
		WhereLike(columns.MenuKey, prefixPattern).
		WhereOrLike(columns.Remark, remarkPattern).
		OrderAsc(columns.Id).
		Scan(&menus); err != nil {
		return nil, err
	}
	return menus, nil
}

func (s *serviceImpl) validateHostedMenuBindings(ctx context.Context, manifest *catalog.Manifest, menus []*entity.SysMenu) error {
	if manifest == nil || manifest.RuntimeArtifact == nil || len(menus) == 0 {
		return nil
	}

	var b *bundle
	for _, menu := range menus {
		if menu == nil || catalog.ParsePluginIDFromMenu(menu) != manifest.ID {
			continue
		}

		relativeAssetPath, usesHostedAsset, err := s.resolveHostedMenuAssetPath(manifest, menu.Path)
		if err != nil {
			return wrapMenuValidationError(menu, err)
		}
		if !usesHostedAsset {
			continue
		}

		if b == nil {
			b, err = s.ensureBundle(ctx, manifest)
			if err != nil {
				return wrapMenuValidationError(menu, err)
			}
		}
		if !b.HasAsset(relativeAssetPath) {
			return wrapMenuValidationError(
				menu,
				gerror.Newf("菜单引用的运行时前端资源不存在: %s", relativeAssetPath),
			)
		}

		queryParams, err := parseMenuQueryParams(menu.QueryParam)
		if err != nil {
			return wrapMenuValidationError(menu, err)
		}
		if err = validateHostedMenuMode(menu, queryParams, relativeAssetPath); err != nil {
			return wrapMenuValidationError(menu, err)
		}
	}
	return nil
}

func (s *serviceImpl) resolveHostedMenuAssetPath(
	manifest *catalog.Manifest,
	menuPath string,
) (string, bool, error) {
	normalizedPath := normalizeHostedMenuPath(menuPath)
	if !strings.HasPrefix(normalizedPath, hostedAssetURLPrefix) {
		return "", false, nil
	}

	expectedPrefix := s.BuildRuntimeFrontendPublicBaseURL(manifest.ID, manifest.Version)
	if !strings.HasPrefix(normalizedPath, expectedPrefix) {
		return "", true, gerror.Newf(
			"菜单必须引用当前插件版本的托管资源: expected prefix %s",
			expectedPrefix,
		)
	}

	relativeAssetPath := strings.TrimPrefix(normalizedPath, expectedPrefix)
	if strings.TrimSpace(relativeAssetPath) == "" {
		relativeAssetPath = "index.html"
	}
	return NormalizeAssetPath(relativeAssetPath), true, nil
}

// ValidateHostedMenuBindings is the exported form of validateHostedMenuBindings for cross-package access.
func (s *serviceImpl) ValidateHostedMenuBindings(ctx context.Context, manifest *catalog.Manifest, menus []*entity.SysMenu) error {
	return s.validateHostedMenuBindings(ctx, manifest, menus)
}

func wrapMenuValidationError(menu *entity.SysMenu, err error) error {
	if menu == nil {
		return err
	}
	return gerror.Wrapf(err, "插件菜单校验失败[%s/%s]", strings.TrimSpace(menu.Name), strings.TrimSpace(menu.MenuKey))
}

func normalizeHostedMenuPath(menuPath string) string {
	trimmedPath := strings.TrimSpace(menuPath)
	if trimmedPath == "" {
		return ""
	}
	if strings.HasPrefix(trimmedPath, "/") {
		return trimmedPath
	}
	return "/" + trimmedPath
}

func parseMenuQueryParams(rawQuery string) (map[string]string, error) {
	trimmedQuery := strings.TrimSpace(rawQuery)
	if trimmedQuery == "" {
		return map[string]string{}, nil
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal([]byte(trimmedQuery), &decoded); err != nil {
		return nil, gerror.Wrap(err, "菜单 query_param 不是合法 JSON")
	}

	queryParams := make(map[string]string, len(decoded))
	for key, value := range decoded {
		if strings.TrimSpace(key) == "" {
			continue
		}
		queryParams[key] = fmt.Sprint(value)
	}
	return queryParams, nil
}

func validateHostedMenuMode(
	menu *entity.SysMenu,
	queryParams map[string]string,
	relativeAssetPath string,
) error {
	componentPath := strings.TrimSpace(menu.Component)
	accessMode := strings.TrimSpace(queryParams[menuQueryKeyAccessMode])
	isEmbeddedComponent := componentPath == dynamicPageComponentPath

	if accessMode == accessModeEmbedded {
		if !isEmbeddedComponent {
			return gerror.Newf("宿主内嵌挂载菜单必须使用组件 %s", dynamicPageComponentPath)
		}
		if menu.IsFrame != 0 {
			return gerror.New("宿主内嵌挂载菜单不能声明为外链")
		}
		extension := strings.ToLower(filepath.Ext(relativeAssetPath))
		if extension != embeddedJSExtension && extension != embeddedMJSExtension {
			return gerror.New("宿主内嵌挂载入口必须指向 .js 或 .mjs ESM 资源")
		}
		return nil
	}

	if isEmbeddedComponent {
		return gerror.Newf(
			"使用组件 %s 的托管资源菜单必须声明 query_param.%s=%s",
			dynamicPageComponentPath,
			menuQueryKeyAccessMode,
			accessModeEmbedded,
		)
	}
	return nil
}
