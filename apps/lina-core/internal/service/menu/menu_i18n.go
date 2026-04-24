// This file localizes menu display fields using stable menu keys and runtime i18n resources.

package menu

import (
	"context"
	"strings"

	"lina-core/internal/model/entity"
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
	translationKey := buildMenuTranslationKey(menu.MenuKey, menu.Name)
	if translationKey == "" {
		return
	}
	menu.Name = s.i18nSvc.Translate(ctx, translationKey, menu.Name)
}

// buildMenuTranslationKey derives the runtime translation key for one menu.
func buildMenuTranslationKey(menuKey string, name string) string {
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
