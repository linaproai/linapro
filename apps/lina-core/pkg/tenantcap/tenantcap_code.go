// This file defines shared tenant capability business error codes.

package tenantcap

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeTenantRequired reports that a tenant must be selected before continuing.
	CodeTenantRequired = bizerr.MustDefine(
		"TENANT_REQUIRED",
		"Tenant selection is required",
		gcode.CodeNotAuthorized,
	)
	// CodeTenantForbidden reports that the current user cannot access the requested tenant.
	CodeTenantForbidden = bizerr.MustDefine(
		"TENANT_FORBIDDEN",
		"Current user cannot access tenant {tenantId}",
		gcode.CodeNotAuthorized,
	)
	// CodeCrossTenantNotAllowed reports an implicit cross-tenant operation.
	CodeCrossTenantNotAllowed = bizerr.MustDefine(
		"CROSS_TENANT_NOT_ALLOWED",
		"Cross-tenant access is not allowed in this context",
		gcode.CodeNotAuthorized,
	)
	// CodePlatformPermissionRequired reports that platform permission is required.
	CodePlatformPermissionRequired = bizerr.MustDefine(
		"PLATFORM_PERMISSION_REQUIRED",
		"Platform permission is required",
		gcode.CodeNotAuthorized,
	)
	// CodeTenantSuspended reports access to a suspended tenant.
	CodeTenantSuspended = bizerr.MustDefine(
		"TENANT_SUSPENDED",
		"Tenant {tenantId} is suspended",
		gcode.CodeNotAuthorized,
	)
)
