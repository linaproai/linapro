// This file defines structured plugin storage capability error codes.

package storagecap

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeStoragePluginIDRequired reports a storage call before plugin scope binding.
	CodeStoragePluginIDRequired = bizerr.MustDefine(
		"PLUGIN_STORAGE_PLUGIN_ID_REQUIRED",
		"Plugin storage service requires a plugin ID",
		gcode.CodeInvalidParameter,
	)
	// CodeStoragePathRequired reports an empty object path.
	CodeStoragePathRequired = bizerr.MustDefine(
		"PLUGIN_STORAGE_PATH_REQUIRED",
		"Plugin storage path cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeStoragePathInvalid reports an unsafe logical path.
	CodeStoragePathInvalid = bizerr.MustDefine(
		"PLUGIN_STORAGE_PATH_INVALID",
		"Plugin storage path is invalid",
		gcode.CodeInvalidParameter,
	)
	// CodeStoragePathTooLong reports a logical path beyond MaxLogicalPathBytes.
	CodeStoragePathTooLong = bizerr.MustDefine(
		"PLUGIN_STORAGE_PATH_TOO_LONG",
		"Plugin storage path exceeds the limit of {maxBytes} bytes",
		gcode.CodeInvalidParameter,
	)
	// CodeStorageObjectExists reports a Put without overwrite for an existing object.
	CodeStorageObjectExists = bizerr.MustDefine(
		"PLUGIN_STORAGE_OBJECT_EXISTS",
		"Plugin storage object already exists",
		gcode.CodeInvalidParameter,
	)
	// CodeStorageListLimitInvalid reports an invalid list limit.
	CodeStorageListLimitInvalid = bizerr.MustDefine(
		"PLUGIN_STORAGE_LIST_LIMIT_INVALID",
		"Plugin storage list limit is invalid",
		gcode.CodeInvalidParameter,
	)
	// CodeStorageProviderUnavailable reports that no active provider can serve requests.
	CodeStorageProviderUnavailable = bizerr.MustDefine(
		"PLUGIN_STORAGE_PROVIDER_UNAVAILABLE",
		"Plugin storage provider is unavailable",
		gcode.CodeInvalidOperation,
	)
	// CodeStorageProviderConflict reports that multiple provider plugins can serve requests.
	CodeStorageProviderConflict = bizerr.MustDefine(
		"PLUGIN_STORAGE_PROVIDER_CONFLICT",
		"Multiple plugin storage providers are available: {providerIds}",
		gcode.CodeInvalidOperation,
	)
	// CodeStorageProviderAlreadyRegistered reports duplicate provider registration.
	CodeStorageProviderAlreadyRegistered = bizerr.MustDefine(
		"PLUGIN_STORAGE_PROVIDER_ALREADY_REGISTERED",
		"Plugin storage provider {providerId} is already registered",
		gcode.CodeInvalidParameter,
	)
	// CodeStorageProviderIDRequired reports an empty provider identifier.
	CodeStorageProviderIDRequired = bizerr.MustDefine(
		"PLUGIN_STORAGE_PROVIDER_ID_REQUIRED",
		"Plugin storage provider ID cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeStorageProviderFactoryRequired reports a nil provider factory.
	CodeStorageProviderFactoryRequired = bizerr.MustDefine(
		"PLUGIN_STORAGE_PROVIDER_FACTORY_REQUIRED",
		"Plugin storage provider factory is required",
		gcode.CodeInvalidParameter,
	)
	// CodeStorageDirectAccessUnsupported reports that the active provider cannot
	// issue client direct access for the requested operation.
	CodeStorageDirectAccessUnsupported = bizerr.MustDefine(
		"PLUGIN_STORAGE_DIRECT_ACCESS_UNSUPPORTED",
		"Plugin storage direct access is not supported for this operation",
		gcode.CodeInvalidOperation,
	)
	// CodeStorageDirectAccessOperationInvalid reports an unknown direct-access operation.
	CodeStorageDirectAccessOperationInvalid = bizerr.MustDefine(
		"PLUGIN_STORAGE_DIRECT_ACCESS_OPERATION_INVALID",
		"Plugin storage direct access operation is invalid",
		gcode.CodeInvalidParameter,
	)
	// CodeStorageDirectAccessIssueFailed reports provider failure while signing access.
	CodeStorageDirectAccessIssueFailed = bizerr.MustDefine(
		"PLUGIN_STORAGE_DIRECT_ACCESS_ISSUE_FAILED",
		"Failed to issue plugin storage direct access",
		gcode.CodeInternalError,
	)
	// CodeStorageDirectSessionInvalid reports a missing or foreign direct-upload session.
	CodeStorageDirectSessionInvalid = bizerr.MustDefine(
		"PLUGIN_STORAGE_DIRECT_SESSION_INVALID",
		"Plugin storage direct access session is invalid",
		gcode.CodeInvalidParameter,
	)
	// CodeStorageDirectSessionExpired reports an expired direct-upload session.
	CodeStorageDirectSessionExpired = bizerr.MustDefine(
		"PLUGIN_STORAGE_DIRECT_SESSION_EXPIRED",
		"Plugin storage direct access session has expired",
		gcode.CodeInvalidParameter,
	)
	// CodeStorageDirectCompleteFailed reports object validation failure on complete.
	CodeStorageDirectCompleteFailed = bizerr.MustDefine(
		"PLUGIN_STORAGE_DIRECT_COMPLETE_FAILED",
		"Plugin storage direct upload complete validation failed",
		gcode.CodeInvalidParameter,
	)
	// CodeStorageMultipartUnsupported reports that the active provider cannot
	// run cloud multipart uploads.
	CodeStorageMultipartUnsupported = bizerr.MustDefine(
		"PLUGIN_STORAGE_MULTIPART_UNSUPPORTED",
		"Plugin storage multipart upload is not supported by the active provider",
		gcode.CodeInvalidOperation,
	)
	// CodeStorageMultipartSessionInvalid reports a missing or mismatched multipart session.
	CodeStorageMultipartSessionInvalid = bizerr.MustDefine(
		"PLUGIN_STORAGE_MULTIPART_SESSION_INVALID",
		"Plugin storage multipart upload session is invalid",
		gcode.CodeInvalidParameter,
	)
	// CodeStorageMultipartPartInvalid reports an invalid part number, etag, or body.
	CodeStorageMultipartPartInvalid = bizerr.MustDefine(
		"PLUGIN_STORAGE_MULTIPART_PART_INVALID",
		"Plugin storage multipart part is invalid",
		gcode.CodeInvalidParameter,
	)
	// CodeStorageMultipartCompleteFailed reports CompleteMultipart failure.
	CodeStorageMultipartCompleteFailed = bizerr.MustDefine(
		"PLUGIN_STORAGE_MULTIPART_COMPLETE_FAILED",
		"Plugin storage multipart complete failed",
		gcode.CodeInvalidParameter,
	)
)

// NewInvalidDirectAccessOperationError returns a stable invalid-operation error.
func NewInvalidDirectAccessOperationError() error {
	return bizerr.NewCode(CodeStorageDirectAccessOperationInvalid)
}

// NewDirectAccessIssueFailedError returns a stable issue-failed error.
func NewDirectAccessIssueFailedError() error {
	return bizerr.NewCode(CodeStorageDirectAccessIssueFailed)
}
