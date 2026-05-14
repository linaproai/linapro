// This file defines online-session business error codes.

package session

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeSessionStateUnavailable reports that shared online-session hot state
	// cannot be read or written.
	CodeSessionStateUnavailable = bizerr.MustDefineWithKey(
		"SESSION_STATE_UNAVAILABLE",
		"error.session.state.unavailable",
		"Online-session state is temporarily unavailable",
		gcode.CodeInternalError,
	)
)
