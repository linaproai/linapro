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
)
