// This file defines online-session business error codes.

package session

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeSessionStateUnavailable reports that shared online-session hot state
	// cannot be read or written.
	// messageKey is derived as error.session.state.unavailable.
	CodeSessionStateUnavailable = bizerr.MustDefine(
		"SESSION_STATE_UNAVAILABLE",
		"Online-session state is temporarily unavailable",
		gcode.CodeInternalError,
	)
)
