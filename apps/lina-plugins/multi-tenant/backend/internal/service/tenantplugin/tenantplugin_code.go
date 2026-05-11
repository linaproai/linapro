// This file defines tenant plugin-governance business error codes.

package tenantplugin

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeTenantRequired reports that tenant plugin governance needs a tenant context.
	CodeTenantRequired = bizerr.MustDefine(
		"MULTI_TENANT_PLUGIN_TENANT_REQUIRED",
		"Tenant plugin governance requires a tenant context",
		gcode.CodeInvalidOperation,
	)
	// CodePluginNotFound reports that the requested tenant plugin is unavailable.
	CodePluginNotFound = bizerr.MustDefine(
		"MULTI_TENANT_PLUGIN_NOT_FOUND",
		"Tenant plugin does not exist or is not tenant scoped",
		gcode.CodeNotFound,
	)
	// CodePluginRuntimeRevisionUnavailable reports missing plugin-runtime cache coordination state.
	CodePluginRuntimeRevisionUnavailable = bizerr.MustDefine(
		"MULTI_TENANT_PLUGIN_RUNTIME_REVISION_UNAVAILABLE",
		"Plugin runtime cache revision is unavailable",
		gcode.CodeInternalError,
	)
)
