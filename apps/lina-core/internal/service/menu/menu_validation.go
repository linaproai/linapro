// menu_validation.go validates menu write-time invariants that affect sidebar
// navigation rendering and recognition.
package menu

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/dao"
	"lina-core/pkg/menutype"
)

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
		return gerror.Newf("菜单图标[%s]已被其他目录或菜单使用", normalizedIcon)
	}
	return nil
}
