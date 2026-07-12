// This file defines middleware business error codes and their i18n metadata.

package middleware

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeMiddlewareHTTPUnauthorized reports that a protected request has no valid authentication.
	CodeMiddlewareHTTPUnauthorized = bizerr.MustDefine(
		"HTTP_UNAUTHORIZED",
		"Authentication required",
		gcode.CodeNotAuthorized,
	)
	// CodeMiddlewareHTTPForbidden reports that a request is forbidden by middleware.
	CodeMiddlewareHTTPForbidden = bizerr.MustDefine(
		"HTTP_FORBIDDEN",
		"Permission denied",
		gcode.CodeNotAuthorized,
	)
	// CodeMiddlewareHTTPNotFound reports that the requested route or resource was not found.
	CodeMiddlewareHTTPNotFound = bizerr.MustDefine(
		"HTTP_NOT_FOUND",
		"Resource not found",
		gcode.CodeNotFound,
	)
	// CodeMiddlewareHTTPError reports that middleware only has a generic HTTP failure status.
	CodeMiddlewareHTTPError = bizerr.MustDefine(
		"HTTP_ERROR",
		"Request failed",
		gcode.CodeUnknown,
	)
	// CodeMiddlewarePermissionCurrentUserMissing reports that permission middleware has no user context.
	// messageKey is derived as error.permission.current.user.missing.
	CodeMiddlewarePermissionCurrentUserMissing = bizerr.MustDefine(
		"PERMISSION_CURRENT_USER_MISSING",
		"Current authenticated user is unavailable",
		gcode.CodeNotAuthorized,
	)
	// CodeMiddlewarePermissionContextLoadFailed reports that permission context loading failed.
	// messageKey is derived as error.permission.context.load.failed.
	CodeMiddlewarePermissionContextLoadFailed = bizerr.MustDefine(
		"PERMISSION_CONTEXT_LOAD_FAILED",
		"Failed to load API permission context",
		gcode.CodeInternalError,
	)
	// CodeMiddlewarePermissionDeniedRequired reports that the user lacks declared permissions.
	// messageKey is derived as error.permission.denied.required.
	CodeMiddlewarePermissionDeniedRequired = bizerr.MustDefine(
		"PERMISSION_DENIED_REQUIRED",
		"Current user lacks required API permissions: {permissions}",
		gcode.CodeNotAuthorized,
	)
	// CodeMiddlewareUploadFileTooLarge reports that an uploaded file exceeds the configured limit.
	CodeMiddlewareUploadFileTooLarge = bizerr.MustDefine(
		"UPLOAD_FILE_TOO_LARGE",
		"File size must not exceed {maxSizeMB}MB",
		gcode.CodeInvalidParameter,
	)
	// CodeMiddlewareUploadRequestBodyTooLarge reports that a request body exceeds the transport limit.
	CodeMiddlewareUploadRequestBodyTooLarge = bizerr.MustDefine(
		"UPLOAD_REQUEST_BODY_TOO_LARGE",
		"Uploaded file is too large",
		gcode.CodeInvalidParameter,
	)
)
