// This file defines monitor-operlog business error codes and runtime i18n
// metadata.

package operlog

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeOperLogNotFound reports that the requested operation-log record does not exist.
	CodeOperLogNotFound = bizerr.MustDefine(
		"MONITOR_OPERLOG_NOT_FOUND",
		"Operation log does not exist",
		gcode.CodeNotFound,
	)
)
