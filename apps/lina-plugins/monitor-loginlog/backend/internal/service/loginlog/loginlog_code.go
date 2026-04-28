// This file defines monitor-loginlog business error codes and runtime i18n
// metadata.

package loginlog

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeLoginLogNotFound reports that the requested login-log record does not exist.
	CodeLoginLogNotFound = bizerr.MustDefine(
		"MONITOR_LOGINLOG_NOT_FOUND",
		"Login log does not exist",
		gcode.CodeNotFound,
	)
)
