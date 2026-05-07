// This file defines shared data-scope business error codes.

package datascope

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeDataScopeDenied reports that the requested row is outside the current data scope.
	CodeDataScopeDenied = bizerr.MustDefine(
		"DATASCOPE_DENIED",
		"Data is outside the current data permission scope",
		gcode.CodeNotAuthorized,
	)
	// CodeDataScopeNotAuthenticated reports that no authenticated user context exists.
	CodeDataScopeNotAuthenticated = bizerr.MustDefine(
		"DATASCOPE_NOT_AUTHENTICATED",
		"Not signed in",
		gcode.CodeNotAuthorized,
	)
	// CodeDataScopeUnsupported reports that a role carries an unsupported data scope.
	CodeDataScopeUnsupported = bizerr.MustDefine(
		"DATASCOPE_UNSUPPORTED",
		"Unsupported data permission scope: {scope}",
		gcode.CodeInvalidParameter,
	)
)
