// menu_filter.go declares the narrow menu-filter capability that allows menu
// callers to inject plugin-aware visibility filtering without coupling the menu
// component to the full plugin service facade.
package menu

import (
	"context"

	"lina-core/internal/model/entity"
)

// MenuFilter defines the narrow dependency required by the menu service to hide
// menus that should not be exposed for the current host state.
type MenuFilter interface {
	// FilterMenus returns only the menus that should remain visible.
	FilterMenus(ctx context.Context, menus []*entity.SysMenu) []*entity.SysMenu
}

// noopMenuFilter leaves the menu list unchanged when no external filter is injected.
type noopMenuFilter struct{}

// FilterMenus returns the original menu list unchanged.
func (noopMenuFilter) FilterMenus(_ context.Context, menus []*entity.SysMenu) []*entity.SysMenu {
	return menus
}
