// This file validates how dynamic plugin menus consume hosted public assets.
// The host only serves assets declared by plugin.yaml public_assets, and
// enable-time validation prevents broken runtime menus from entering the router.

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
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/pkg/plugin/pluginhost"
)

const (
	// embeddedJSExtension is the allowed ESM entry extension.
	embeddedJSExtension = ".js"
	// embeddedMJSExtension is the allowed ESM module entry extension.
	embeddedMJSExtension = ".mjs"
)

// ValidateRuntimeFrontendMenuBindings verifies that dynamic plugin menus only
// reference declared public assets that exist in the plugin's in-memory bundle.
func (s *serviceImpl) ValidateRuntimeFrontendMenuBindings(ctx context.Context, manifest *catalog.Manifest) error {
	if manifest == nil || plugintypes.NormalizeType(manifest.Type) != plugintypes.TypeDynamic {
		return nil
	}

	menus, err := s.listPluginOwnedMenus(ctx, manifest.ID)
	if err != nil {
		return err
	}
	return s.validateHostedMenuBindings(ctx, manifest, menus)
}

// listPluginOwnedMenus loads menus owned by the target plugin so hosted asset
// validation can inspect persisted menu bindings.
func (s *serviceImpl) listPluginOwnedMenus(ctx context.Context, pluginID string) ([]*entity.SysMenu, error) {
	columns := dao.SysMenu.Columns()
	prefixPattern := plugintypes.MenuKeyPrefix + pluginID + ":%"

	var menus []*entity.SysMenu
	if err := dao.SysMenu.Ctx(ctx).
		WhereLike(columns.MenuKey, prefixPattern).
		OrderAsc(columns.Id).
		Scan(&menus); err != nil {
		return nil, err
	}
	return menus, nil
}

// validateHostedMenuBindings enforces that plugin menus only point at declared
// hosted public assets that exist and satisfy the embedded-mount contract.
func (s *serviceImpl) validateHostedMenuBindings(ctx context.Context, manifest *catalog.Manifest, menus []*entity.SysMenu) error {
	if manifest == nil || manifest.RuntimeArtifact == nil || len(menus) == 0 {
		return nil
	}

	var b *bundle
	for _, menu := range menus {
		if menu == nil || plugintypes.ParsePluginIDFromMenuKey(menu.MenuKey) != manifest.ID {
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
		resolvedAssetPath, assetErr := resolvePublicAssetDeclaration(manifest.PublicAssets, relativeAssetPath)
		if assetErr != nil {
			return wrapMenuValidationError(menu, assetErr)
		}
		if !b.HasAsset(resolvedAssetPath) {
			return wrapMenuValidationError(
				menu,
				gerror.Newf("menu references missing runtime public asset: %s", resolvedAssetPath),
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

// resolveHostedMenuAssetPath extracts the public URL-relative asset path when
// a menu points at one host-served plugin public asset.
func (s *serviceImpl) resolveHostedMenuAssetPath(
	manifest *catalog.Manifest,
	menuPath string,
) (string, bool, error) {
	normalizedPath := normalizeHostedMenuPath(menuPath)
	if !strings.HasPrefix(normalizedPath, pluginhost.HostedAssetURLPrefix) {
		return "", false, nil
	}

	expectedPrefix := s.BuildRuntimeFrontendPublicBaseURL(manifest.ID, manifest.Version)
	if !strings.HasPrefix(normalizedPath, expectedPrefix) {
		return "", true, gerror.Newf(
			"menu must reference hosted assets from the current plugin version: expected prefix %s",
			expectedPrefix,
		)
	}

	relativeAssetPath := strings.TrimPrefix(normalizedPath, expectedPrefix)
	return NormalizeAssetPath(relativeAssetPath), true, nil
}

// ValidateHostedMenuBindings is the exported form of validateHostedMenuBindings for cross-package access.
func (s *serviceImpl) ValidateHostedMenuBindings(ctx context.Context, manifest *catalog.Manifest, menus []*entity.SysMenu) error {
	return s.validateHostedMenuBindings(ctx, manifest, menus)
}

// wrapMenuValidationError enriches hosted-menu validation errors with menu identity.
func wrapMenuValidationError(menu *entity.SysMenu, err error) error {
	if menu == nil {
		return err
	}
	return gerror.Wrapf(err, "plugin menu validation failed [%s/%s]", strings.TrimSpace(menu.Name), strings.TrimSpace(menu.MenuKey))
}

// normalizeHostedMenuPath normalizes menu paths into absolute-style paths.
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

// parseMenuQueryParams decodes stored menu query JSON into a string map used
// by hosted-menu contract validation.
func parseMenuQueryParams(rawQuery string) (map[string]string, error) {
	trimmedQuery := strings.TrimSpace(rawQuery)
	if trimmedQuery == "" {
		return map[string]string{}, nil
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal([]byte(trimmedQuery), &decoded); err != nil {
		return nil, gerror.Wrap(err, "menu query_param is not valid JSON")
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

// validateHostedMenuMode enforces the extra constraints required by embedded
// mount menus that load runtime ESM entry assets inside the host shell.
func validateHostedMenuMode(
	menu *entity.SysMenu,
	queryParams map[string]string,
	relativeAssetPath string,
) error {
	componentPath := strings.TrimSpace(menu.Component)
	accessMode := strings.TrimSpace(queryParams[pluginhost.DynamicAccessModeQueryKey])
	isEmbeddedComponent := componentPath == pluginhost.DynamicPageComponentPath

	if accessMode == pluginhost.DynamicAccessModeEmbeddedMount {
		if !isEmbeddedComponent {
			return gerror.Newf("host embedded mount menus must use component %s", pluginhost.DynamicPageComponentPath)
		}
		if menu.IsFrame != 0 {
			return gerror.New("host embedded mount menus cannot be declared as external links")
		}
		extension := strings.ToLower(filepath.Ext(relativeAssetPath))
		if extension != embeddedJSExtension && extension != embeddedMJSExtension {
			return gerror.New("host embedded mount entry must point to a .js or .mjs ESM asset")
		}
		return nil
	}

	if isEmbeddedComponent {
		return gerror.Newf(
			"hosted asset menus using component %s must declare query_param.%s=%s",
			pluginhost.DynamicPageComponentPath,
			pluginhost.DynamicAccessModeQueryKey,
			pluginhost.DynamicAccessModeEmbeddedMount,
		)
	}
	return nil
}
