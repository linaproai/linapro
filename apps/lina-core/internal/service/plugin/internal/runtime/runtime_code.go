// This file defines dynamic-plugin runtime business error codes and their i18n
// metadata.

package runtime

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeDynamicPluginManifestRequired reports that a dynamic-plugin manifest is required.
	CodeDynamicPluginManifestRequired = bizerr.MustDefine(
		"PLUGIN_DYNAMIC_RUNTIME_MANIFEST_REQUIRED",
		"Dynamic plugin manifest cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeDynamicPluginArtifactMissing reports that a lifecycle action requires a missing Wasm artifact.
	CodeDynamicPluginArtifactMissing = bizerr.MustDefine(
		"PLUGIN_DYNAMIC_ARTIFACT_MISSING",
		"Dynamic plugin is missing {artifactPath} and cannot {action}. Run make wasm p={pluginId} first",
		gcode.CodeInvalidParameter,
	)
	// CodeDynamicPluginArtifactValidateFailed reports artifact validation failure before a lifecycle action.
	CodeDynamicPluginArtifactValidateFailed = bizerr.MustDefine(
		"PLUGIN_DYNAMIC_ARTIFACT_VALIDATE_FAILED",
		"Dynamic plugin artifact validation failed and cannot {action}",
		gcode.CodeInvalidParameter,
	)
)
