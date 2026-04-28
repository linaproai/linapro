// This file defines menu-service business error codes and their i18n metadata.

package menu

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeMenuNotFound reports that the requested menu record does not exist.
	CodeMenuNotFound = bizerr.MustDefine(
		"MENU_NOT_FOUND",
		"Menu does not exist",
		gcode.CodeNotFound,
	)
	// CodeMenuMoveToSelfDenied reports that a menu cannot be moved under itself.
	CodeMenuMoveToSelfDenied = bizerr.MustDefine(
		"MENU_MOVE_TO_SELF_DENIED",
		"Cannot move a menu under itself",
		gcode.CodeInvalidParameter,
	)
	// CodeMenuMoveToDescendantDenied reports that a menu cannot be moved under its descendant.
	CodeMenuMoveToDescendantDenied = bizerr.MustDefine(
		"MENU_MOVE_TO_DESCENDANT_DENIED",
		"Cannot move a menu under its child menu",
		gcode.CodeInvalidParameter,
	)
	// CodeMenuHasChildrenDeleteDenied reports that a menu with children cannot be deleted without cascade.
	CodeMenuHasChildrenDeleteDenied = bizerr.MustDefine(
		"MENU_HAS_CHILDREN_DELETE_DENIED",
		"Menu has child menus and cannot be deleted",
		gcode.CodeInvalidParameter,
	)
	// CodeMenuNameExists reports that a sibling menu already uses the same name.
	CodeMenuNameExists = bizerr.MustDefine(
		"MENU_NAME_EXISTS",
		"Menu name already exists under the same parent",
		gcode.CodeInvalidParameter,
	)
	// CodeMenuIconExists reports that another menu already uses the same sidebar icon.
	CodeMenuIconExists = bizerr.MustDefine(
		"MENU_ICON_EXISTS",
		"Menu icon {icon} is already used by another directory or menu",
		gcode.CodeInvalidParameter,
	)
)
