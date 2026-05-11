// This file defines resolver business error codes and runtime i18n metadata.

package resolver

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeTenantRequired reports that the resolver chain could not choose a tenant.
	CodeTenantRequired = bizerr.MustDefine(
		"MULTI_TENANT_TENANT_REQUIRED",
		"Tenant selection is required",
		gcode.CodeNotAuthorized,
	)
	// CodeTenantForbidden reports that the resolved tenant is not visible to the user.
	CodeTenantForbidden = bizerr.MustDefine(
		"MULTI_TENANT_TENANT_FORBIDDEN",
		"User is not allowed to access the tenant",
		gcode.CodeNotAuthorized,
	)
	// CodePlatformPermissionRequired reports that a platform-only resolver was used by a non-platform user.
	CodePlatformPermissionRequired = bizerr.MustDefine(
		"MULTI_TENANT_PLATFORM_PERMISSION_REQUIRED",
		"Platform administrator permission is required",
		gcode.CodeNotAuthorized,
	)
	// CodeTenantOverrideInvalid reports that the tenant override header is malformed.
	CodeTenantOverrideInvalid = bizerr.MustDefine(
		"MULTI_TENANT_TENANT_OVERRIDE_INVALID",
		"Tenant override header must be a valid tenant ID",
		gcode.CodeInvalidParameter,
	)
)
