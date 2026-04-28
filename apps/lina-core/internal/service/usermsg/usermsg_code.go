// This file defines user-message business error codes and their i18n metadata.

package usermsg

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeUserMsgNotAuthenticated reports that the current request has no authenticated user.
	CodeUserMsgNotAuthenticated = bizerr.MustDefine(
		"USERMSG_NOT_AUTHENTICATED",
		"Not signed in",
		gcode.CodeNotAuthorized,
	)
	// CodeUserMsgNotFound reports that the requested current-user message does not exist.
	CodeUserMsgNotFound = bizerr.MustDefine(
		"USERMSG_NOT_FOUND",
		"Message does not exist",
		gcode.CodeNotFound,
	)
)
