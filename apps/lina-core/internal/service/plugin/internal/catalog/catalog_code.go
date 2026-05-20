// This file defines catalog-level business error codes returned while plugin
// manifests and runtime artifacts are validated before they enter lifecycle
// storage or HTTP-visible plugin governance flows.

package catalog

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
