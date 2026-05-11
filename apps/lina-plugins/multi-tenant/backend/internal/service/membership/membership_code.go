// This file defines membership business error codes and runtime i18n metadata.

package membership

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeMembershipNotFound reports that the requested membership does not exist.
	CodeMembershipNotFound = bizerr.MustDefine(
		"MULTI_TENANT_MEMBERSHIP_NOT_FOUND",
		"Tenant membership does not exist",
		gcode.CodeNotFound,
	)
	// CodeMembershipExists reports that a user already belongs to the tenant.
	CodeMembershipExists = bizerr.MustDefine(
		"MULTI_TENANT_MEMBERSHIP_EXISTS",
		"User is already a member of the tenant",
		gcode.CodeInvalidParameter,
	)
	// CodePlatformMembershipForbidden reports that platform users cannot be added to tenant memberships.
	CodePlatformMembershipForbidden = bizerr.MustDefine(
		"MULTI_TENANT_PLATFORM_MEMBERSHIP_FORBIDDEN",
		"Platform users cannot be added to tenant memberships",
		gcode.CodeInvalidOperation,
	)
	// CodeSingleCardinalityViolation reports that single-cardinality mode rejected a second membership.
	CodeSingleCardinalityViolation = bizerr.MustDefine(
		"MULTI_TENANT_SINGLE_CARDINALITY_VIOLATION",
		"User already belongs to another tenant",
		gcode.CodeInvalidOperation,
	)
	// CodeTenantUnavailable reports that a membership targets a non-active tenant.
	CodeTenantUnavailable = bizerr.MustDefine(
		"MULTI_TENANT_TENANT_UNAVAILABLE",
		"Tenant is not available",
		gcode.CodeInvalidOperation,
	)
)
