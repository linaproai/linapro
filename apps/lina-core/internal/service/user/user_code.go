// This file defines user-service business error codes and their i18n metadata.

package user

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeUserUsernameExists reports that the requested username is already used.
	CodeUserUsernameExists = bizerr.MustDefine(
		"USER_USERNAME_EXISTS",
		"Username already exists",
		gcode.CodeInvalidParameter,
	)
	// CodeUserNotFound reports that the requested user record does not exist.
	CodeUserNotFound = bizerr.MustDefine(
		"USER_NOT_FOUND",
		"User does not exist",
		gcode.CodeNotFound,
	)
	// CodeUserCurrentEditDenied reports that the current user cannot edit itself through admin management.
	CodeUserCurrentEditDenied = bizerr.MustDefine(
		"USER_CURRENT_EDIT_DENIED",
		"Cannot edit the current signed-in user",
		gcode.CodeNotAuthorized,
	)
	// CodeUserBuiltinAdminDeleteDenied reports that the built-in administrator cannot be deleted.
	CodeUserBuiltinAdminDeleteDenied = bizerr.MustDefine(
		"USER_BUILTIN_ADMIN_DELETE_DENIED",
		"Cannot delete the built-in administrator",
		gcode.CodeNotAuthorized,
	)
	// CodeUserCurrentDeleteDenied reports that the current user cannot delete itself.
	CodeUserCurrentDeleteDenied = bizerr.MustDefine(
		"USER_CURRENT_DELETE_DENIED",
		"Cannot delete the current signed-in user",
		gcode.CodeNotAuthorized,
	)
	// CodeUserDeleteIdsRequired reports that a batch delete request has no user IDs.
	CodeUserDeleteIdsRequired = bizerr.MustDefine(
		"USER_DELETE_IDS_REQUIRED",
		"Please select users to delete",
		gcode.CodeInvalidParameter,
	)
	// CodeUserCurrentDisableDenied reports that the current user cannot disable itself.
	CodeUserCurrentDisableDenied = bizerr.MustDefine(
		"USER_CURRENT_DISABLE_DENIED",
		"Cannot disable the current signed-in user",
		gcode.CodeNotAuthorized,
	)
	// CodeUserNotAuthenticated reports that the current request has no authenticated user context.
	CodeUserNotAuthenticated = bizerr.MustDefine(
		"USER_NOT_AUTHENTICATED",
		"Not signed in",
		gcode.CodeNotAuthorized,
	)
	// CodeUserDataScopeDenied reports that the requested user row is outside the current data scope.
	CodeUserDataScopeDenied = bizerr.MustDefine(
		"USER_DATA_SCOPE_DENIED",
		"User data is outside the current data permission scope",
		gcode.CodeNotAuthorized,
	)
	// CodeUserDataScopeUnsupported reports that an enabled role has an unsupported data scope.
	CodeUserDataScopeUnsupported = bizerr.MustDefine(
		"USER_DATA_SCOPE_UNSUPPORTED",
		"Unsupported user data permission scope: {scope}",
		gcode.CodeInvalidParameter,
	)
	// CodeUserTenantMembershipQueryFailed reports failure while checking tenant membership visibility.
	CodeUserTenantMembershipQueryFailed = bizerr.MustDefine(
		"USER_TENANT_MEMBERSHIP_QUERY_FAILED",
		"Failed to query tenant membership visibility",
		gcode.CodeInternalError,
	)
	// CodeUserTenantMembershipReplaceFailed reports failure while replacing tenant membership.
	CodeUserTenantMembershipReplaceFailed = bizerr.MustDefine(
		"USER_TENANT_MEMBERSHIP_REPLACE_FAILED",
		"Failed to update tenant membership",
		gcode.CodeInternalError,
	)
	// CodeUserTenantMembershipCrossTenantDenied reports cross-tenant membership writes.
	CodeUserTenantMembershipCrossTenantDenied = bizerr.MustDefine(
		"USER_TENANT_MEMBERSHIP_CROSS_TENANT_DENIED",
		"Cannot assign users to another tenant in the current context",
		gcode.CodeNotAuthorized,
	)
	// CodeUserTenantMembershipTenantUnavailable reports unavailable tenant assignment.
	CodeUserTenantMembershipTenantUnavailable = bizerr.MustDefine(
		"USER_TENANT_MEMBERSHIP_TENANT_UNAVAILABLE",
		"Selected tenant is unavailable",
		gcode.CodeInvalidParameter,
	)
	// CodeUserTenantMembershipCardinalityExceeded reports single-cardinality membership violations.
	CodeUserTenantMembershipCardinalityExceeded = bizerr.MustDefine(
		"USER_TENANT_MEMBERSHIP_CARDINALITY_EXCEEDED",
		"User can only belong to one tenant in the current configuration",
		gcode.CodeInvalidParameter,
	)
	// CodeUserImportExcelParseFailed reports that the uploaded user workbook cannot be parsed.
	CodeUserImportExcelParseFailed = bizerr.MustDefine(
		"USER_IMPORT_EXCEL_PARSE_FAILED",
		"Failed to parse Excel file",
		gcode.CodeInvalidParameter,
	)
	// CodeUserImportFileRequired reports that no user import file was uploaded.
	CodeUserImportFileRequired = bizerr.MustDefine(
		"USER_IMPORT_FILE_REQUIRED",
		"Please select a file to import",
		gcode.CodeInvalidParameter,
	)
	// CodeUserImportSheetReadFailed reports that the expected worksheet cannot be read.
	CodeUserImportSheetReadFailed = bizerr.MustDefine(
		"USER_IMPORT_SHEET_READ_FAILED",
		"Failed to read worksheet {sheet}",
		gcode.CodeInvalidParameter,
	)
)
