// This file defines business error codes for pure plugin value validation.

package plugintypes

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodePluginIDInvalid reports that a manifest, dependency, or runtime
	// artifact plugin ID violates the runtime-safe plugin ID boundary.
	CodePluginIDInvalid = bizerr.MustDefine(
		"PLUGIN_ID_INVALID",
		"Plugin ID {pluginId} is invalid: {reason}",
		gcode.CodeInvalidParameter,
	)
)
