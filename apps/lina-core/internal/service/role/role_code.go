// This file defines role-service business error codes and their i18n metadata.

package role

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeRoleNotFound reports that the requested role does not exist.
	CodeRoleNotFound = bizerr.MustDefine(
		"ROLE_NOT_FOUND",
		"Role does not exist",
		gcode.CodeNotFound,
	)
	// CodeRoleNameExists reports that a role name already exists.
	CodeRoleNameExists = bizerr.MustDefine(
		"ROLE_NAME_EXISTS",
		"Role name already exists",
		gcode.CodeInvalidParameter,
	)
	// CodeRoleKeyExists reports that a role permission key already exists.
	CodeRoleKeyExists = bizerr.MustDefine(
		"ROLE_KEY_EXISTS",
		"Role permission key already exists",
		gcode.CodeInvalidParameter,
	)
	// CodeRoleBuiltinDeleteDenied reports that the built-in administrator role cannot be deleted.
	CodeRoleBuiltinDeleteDenied = bizerr.MustDefine(
		"ROLE_BUILTIN_DELETE_DENIED",
		"Cannot delete the built-in administrator role",
		gcode.CodeNotAuthorized,
	)
	// CodeRoleDeleteIdsRequired reports that a batch delete request has no role IDs.
	CodeRoleDeleteIdsRequired = bizerr.MustDefine(
		"ROLE_DELETE_IDS_REQUIRED",
		"Please select roles to delete",
		gcode.CodeInvalidParameter,
	)
	// CodeRoleDataScopeDeptUnavailable reports that department data scope requires org-center.
	CodeRoleDataScopeDeptUnavailable = bizerr.MustDefine(
		"ROLE_DATA_SCOPE_DEPT_UNAVAILABLE",
		"Department data scope requires the organization management plugin to be enabled",
		gcode.CodeInvalidParameter,
	)
	// CodeRoleDataScopeUnsupported reports that a submitted role data-scope
	// value is outside the supported host contract.
	CodeRoleDataScopeUnsupported = bizerr.MustDefine(
		"ROLE_DATA_SCOPE_UNSUPPORTED",
		"Unsupported role data scope: {scope}",
		gcode.CodeInvalidParameter,
	)
	// CodeTenantRoleAllDataScopeForbidden reports that tenant roles cannot
	// receive cross-tenant all-data scope.
	CodeTenantRoleAllDataScopeForbidden = bizerr.MustDefine(
		"TENANT_ROLE_ALL_DATA_SCOPE_FORBIDDEN",
		"Tenant roles cannot use all-data scope",
		gcode.CodeInvalidParameter,
	)
	// CodeRoleTenantMismatch reports that a role or role relation does not
	// belong to the current tenant boundary.
	CodeRoleTenantMismatch = bizerr.MustDefine(
		"ROLE_TENANT_MISMATCH",
		"Role does not belong to the current tenant",
		gcode.CodeNotAuthorized,
	)
	// CodePlatformRoleAssignmentForbidden reports a platform role assignment to
	// a non-platform user or tenant context.
	CodePlatformRoleAssignmentForbidden = bizerr.MustDefine(
		"PLATFORM_ROLE_ASSIGNMENT_FORBIDDEN",
		"Platform roles can only be assigned in platform context",
		gcode.CodeNotAuthorized,
	)
	// CodeTenantRoleAssignmentForbidden reports a tenant role assignment to a
	// platform user or a user without active membership in the role tenant.
	CodeTenantRoleAssignmentForbidden = bizerr.MustDefine(
		"TENANT_ROLE_ASSIGNMENT_FORBIDDEN",
		"Tenant roles can only be assigned to active users in the same tenant",
		gcode.CodeNotAuthorized,
	)
)
