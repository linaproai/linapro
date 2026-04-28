// This file defines organization department business error codes and runtime
// i18n metadata.

package dept

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeDeptNotFound reports that the requested department record does not exist.
	CodeDeptNotFound = bizerr.MustDefine(
		"ORG_DEPT_NOT_FOUND",
		"Department does not exist",
		gcode.CodeNotFound,
	)
	// CodeDeptHasChildrenDeleteDenied reports that a department still has child departments.
	CodeDeptHasChildrenDeleteDenied = bizerr.MustDefine(
		"ORG_DEPT_HAS_CHILDREN_DELETE_DENIED",
		"Department has child departments and cannot be deleted",
		gcode.CodeInvalidParameter,
	)
	// CodeDeptHasUsersDeleteDenied reports that a department still has assigned users.
	CodeDeptHasUsersDeleteDenied = bizerr.MustDefine(
		"ORG_DEPT_HAS_USERS_DELETE_DENIED",
		"Department has assigned users and cannot be deleted",
		gcode.CodeInvalidParameter,
	)
	// CodeDeptCodeExists reports that a department code is already used.
	CodeDeptCodeExists = bizerr.MustDefine(
		"ORG_DEPT_CODE_EXISTS",
		"Department code already exists",
		gcode.CodeInvalidParameter,
	)
)
