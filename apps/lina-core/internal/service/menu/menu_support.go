// This file contains supporting menu service rules: runtime localization,
// sidebar icon validation, and platform-context checks for global menu writes.

package menu

import (
	"context"
	"strings"

	"lina-core/internal/dao"
	"lina-core/internal/model/entity"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/menutype"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
)

// localizeMenuEntities localizes one menu-entity list in place.
func (s *serviceImpl) localizeMenuEntities(ctx context.Context, menus []*entity.SysMenu) {
	for _, menu := range menus {
		s.localizeMenuEntity(ctx, menu)
	}
}

// localizeMenuEntity localizes one menu entity in place.
func (s *serviceImpl) localizeMenuEntity(ctx context.Context, menu *entity.SysMenu) {
	if s == nil || s.i18nSvc == nil || menu == nil {
		return
	}
	translationKey := buildMenuTitleKey(menu.MenuKey, menu.Name)
	if translationKey == "" {
		return
	}
	menu.Name = s.i18nSvc.Translate(ctx, translationKey, menu.Name)
}

// buildMenuTitleKey derives the runtime translation key for one menu title.
func buildMenuTitleKey(menuKey string, name string) string {
	trimmedMenuKey := strings.TrimSpace(menuKey)
	if trimmedMenuKey != "" {
		return "menu." + trimmedMenuKey + ".title"
	}

	trimmedName := strings.TrimSpace(name)
	if strings.Contains(trimmedName, ".") {
		return trimmedName
	}
	return ""
}

// menuIconPlaceholder marks entries without a rendered icon.
const menuIconPlaceholder = "#"

// normalizeMenuIcon trims menu icon input before validation or persistence.
func normalizeMenuIcon(icon string) string {
	return strings.TrimSpace(icon)
}

// shouldValidateMenuIcon reports whether the current menu state participates in
// sidebar icon uniqueness checks.
func shouldValidateMenuIcon(menuType, icon string) bool {
	normalizedIcon := normalizeMenuIcon(icon)
	if normalizedIcon == "" || normalizedIcon == menuIconPlaceholder {
		return false
	}

	return menuType == menutype.Directory.String() || menuType == menutype.Menu.String()
}

// checkIconUnique ensures directory and menu icons remain globally unique so
// the sidebar never renders repeated iconography.
func (s *serviceImpl) checkIconUnique(ctx context.Context, menuType, icon string, excludeID int) error {
	normalizedIcon := normalizeMenuIcon(icon)
	if !shouldValidateMenuIcon(menuType, normalizedIcon) {
		return nil
	}

	cols := dao.SysMenu.Columns()
	m := dao.SysMenu.Ctx(ctx).
		Where(cols.Icon, normalizedIcon).
		WhereIn(cols.Type, []string{menutype.Directory.String(), menutype.Menu.String()})
	if excludeID > 0 {
		m = m.WhereNot(cols.Id, excludeID)
	}

	count, err := m.Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return bizerr.NewCode(CodeMenuIconExists, bizerr.P("icon", normalizedIcon))
	}
	return nil
}

// ensurePlatformMenuGovernance verifies the current request can mutate the
// global menu topology.
func (s *serviceImpl) ensurePlatformMenuGovernance(ctx context.Context) error {
	if s == nil {
		return nil
	}
	return ensurePlatformMenuGovernanceContext(ctx, s.tenantSvc)
}

// ensurePlatformMenuGovernanceContext applies platform-menu checks to one
// tenant service.
func ensurePlatformMenuGovernanceContext(ctx context.Context, tenantSvc tenantspi.Service) error {
	if tenantSvc == nil || !tenantSvc.Available(ctx) || tenantSvc.PlatformBypass(ctx) {
		return nil
	}
	return bizerr.NewCode(tenantcap.CodePlatformPermissionRequired)
}
