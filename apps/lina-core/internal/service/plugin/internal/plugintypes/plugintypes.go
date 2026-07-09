// Package plugintypes defines pure plugin value objects shared by the host
// plugin sub-components. It must remain a leaf package: no catalog, store,
// runtime, integration, DAO, DO, or Entity dependencies are allowed here.
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
