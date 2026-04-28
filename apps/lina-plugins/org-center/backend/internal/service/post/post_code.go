// This file defines organization post business error codes and runtime i18n
// metadata.

package post

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodePostNotFound reports that the requested post record does not exist.
	CodePostNotFound = bizerr.MustDefine(
		"ORG_POST_NOT_FOUND",
		"Post does not exist",
		gcode.CodeNotFound,
	)
	// CodePostDeleteRequired reports that a delete operation received no post IDs.
	CodePostDeleteRequired = bizerr.MustDefine(
		"ORG_POST_DELETE_REQUIRED",
		"Select at least one post to delete",
		gcode.CodeInvalidParameter,
	)
	// CodePostAssignedDeleteDenied reports that a post is assigned to users.
	CodePostAssignedDeleteDenied = bizerr.MustDefine(
		"ORG_POST_ASSIGNED_DELETE_DENIED",
		"Post {id} has assigned users and cannot be deleted",
		gcode.CodeInvalidParameter,
	)
	// CodePostValidIDRequired reports that no valid post IDs remain after parsing.
	CodePostValidIDRequired = bizerr.MustDefine(
		"ORG_POST_VALID_ID_REQUIRED",
		"No valid post IDs were provided",
		gcode.CodeInvalidParameter,
	)
	// CodePostCodeExists reports that a post code is already used.
	CodePostCodeExists = bizerr.MustDefine(
		"ORG_POST_CODE_EXISTS",
		"Post code already exists",
		gcode.CodeInvalidParameter,
	)
)
