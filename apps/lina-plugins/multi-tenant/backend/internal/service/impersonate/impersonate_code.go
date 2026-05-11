// This file defines impersonation business error codes and runtime i18n metadata.

package impersonate

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeImpersonationPermissionDenied reports that the current user cannot impersonate tenants.
	CodeImpersonationPermissionDenied = bizerr.MustDefine(
		"MULTI_TENANT_IMPERSONATION_PERMISSION_DENIED",
		"Only platform administrators can start tenant impersonation",
		gcode.CodeNotAuthorized,
	)
	// CodeImpersonationTenantUnavailable reports that the target tenant cannot be impersonated.
	CodeImpersonationTenantUnavailable = bizerr.MustDefine(
		"MULTI_TENANT_IMPERSONATION_TENANT_UNAVAILABLE",
		"Target tenant cannot be impersonated",
		gcode.CodeInvalidOperation,
	)
	// CodeImpersonationTokenInvalid reports that an impersonation token is malformed or not impersonated.
	CodeImpersonationTokenInvalid = bizerr.MustDefine(
		"MULTI_TENANT_IMPERSONATION_TOKEN_INVALID",
		"Impersonation token is invalid",
		gcode.CodeNotAuthorized,
	)
	// CodeImpersonationTokenUnavailable reports that token signing is unavailable.
	CodeImpersonationTokenUnavailable = bizerr.MustDefine(
		"MULTI_TENANT_IMPERSONATION_TOKEN_UNAVAILABLE",
		"Impersonation token signing is not available",
		gcode.CodeInvalidOperation,
	)
)
