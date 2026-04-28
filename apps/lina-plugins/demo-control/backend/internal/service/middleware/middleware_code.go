// This file defines demo-control business error codes and runtime i18n
// metadata.

package middleware

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeDemoControlWriteDenied reports that demo mode rejected a write request.
	CodeDemoControlWriteDenied = bizerr.MustDefine(
		"DEMO_CONTROL_WRITE_DENIED",
		"Demo mode is enabled; write operations are disabled",
		gcode.CodeNotAuthorized,
	)
)
