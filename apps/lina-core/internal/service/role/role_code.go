// This file defines role-service business error codes and their i18n metadata.

package role

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeRoleNotFound reports that the requested role does not exist.
	CodeRoleNotFound = bizerr.MustDefine(
		"ROLE_NOT_FOUND",
		"Role does not exist",
		gcode.CodeNotFound,
	)
	// CodeRoleNameExists reports that a role name already exists.
	CodeRoleNameExists = bizerr.MustDefine(
		"ROLE_NAME_EXISTS",
		"Role name already exists",
		gcode.CodeInvalidParameter,
	)
	// CodeRoleKeyExists reports that a role permission key already exists.
	CodeRoleKeyExists = bizerr.MustDefine(
		"ROLE_KEY_EXISTS",
		"Role permission key already exists",
		gcode.CodeInvalidParameter,
	)
	// CodeRoleBuiltinDeleteDenied reports that the built-in administrator role cannot be deleted.
	CodeRoleBuiltinDeleteDenied = bizerr.MustDefine(
		"ROLE_BUILTIN_DELETE_DENIED",
		"Cannot delete the built-in administrator role",
		gcode.CodeNotAuthorized,
	)
	// CodeRoleDeleteIdsRequired reports that a batch delete request has no role IDs.
	CodeRoleDeleteIdsRequired = bizerr.MustDefine(
		"ROLE_DELETE_IDS_REQUIRED",
		"Please select roles to delete",
		gcode.CodeInvalidParameter,
	)
)
