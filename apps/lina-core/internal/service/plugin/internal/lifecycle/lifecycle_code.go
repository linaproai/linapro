// This file defines plugin lifecycle business error codes and their i18n
// metadata.

package lifecycle

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeSourcePluginInstallUnsupported reports that source plugins cannot be installed by dynamic lifecycle.
	CodeSourcePluginInstallUnsupported = bizerr.MustDefine(
		"PLUGIN_SOURCE_INSTALL_UNSUPPORTED",
		"Source plugins are compiled into the host and cannot be installed by the dynamic lifecycle",
		gcode.CodeInvalidParameter,
	)
	// CodeDynamicPluginDowngradeUnsupported reports that dynamic release rollback is unsupported here.
	CodeDynamicPluginDowngradeUnsupported = bizerr.MustDefine(
		"PLUGIN_DYNAMIC_VERSION_DOWNGRADE_UNSUPPORTED",
		"Downgrading to an older dynamic plugin version is not supported. Use host rollback output or upload a newer version",
		gcode.CodeInvalidParameter,
	)
	// CodeSourcePluginUninstallUnsupported reports that source plugins cannot be uninstalled by dynamic lifecycle.
	CodeSourcePluginUninstallUnsupported = bizerr.MustDefine(
		"PLUGIN_SOURCE_UNINSTALL_UNSUPPORTED",
		"Source plugins are compiled into the host and cannot be uninstalled by the dynamic lifecycle",
		gcode.CodeInvalidParameter,
	)
)
