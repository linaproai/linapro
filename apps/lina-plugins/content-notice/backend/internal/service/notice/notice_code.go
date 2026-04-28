// This file defines content-notice business error codes and runtime i18n
// metadata.

package notice

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeNoticeNotFound reports that the requested notice record does not exist.
	CodeNoticeNotFound = bizerr.MustDefine(
		"CONTENT_NOTICE_NOT_FOUND",
		"Notice does not exist",
		gcode.CodeNotFound,
	)
	// CodeNoticeDeleteRequired reports that a delete operation received no notice IDs.
	CodeNoticeDeleteRequired = bizerr.MustDefine(
		"CONTENT_NOTICE_DELETE_REQUIRED",
		"Select at least one notice to delete",
		gcode.CodeInvalidParameter,
	)
)
