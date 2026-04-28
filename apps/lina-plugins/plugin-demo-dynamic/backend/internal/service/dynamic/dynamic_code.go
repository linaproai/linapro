// This file defines dynamic-demo business error codes and runtime i18n
// metadata.

package dynamicservice

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeDynamicDemoRecordInvalidInput reports a generic invalid demo-record request.
	CodeDynamicDemoRecordInvalidInput = bizerr.MustDefine(
		"PLUGIN_DEMO_DYNAMIC_RECORD_INVALID_INPUT",
		"Demo record request is invalid",
		gcode.CodeInvalidParameter,
	)
	// CodeDynamicDemoRecordRequestBodyRequired reports that a request body is required.
	CodeDynamicDemoRecordRequestBodyRequired = bizerr.MustDefine(
		"PLUGIN_DEMO_DYNAMIC_RECORD_REQUEST_BODY_REQUIRED",
		"Request body cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeDynamicDemoRecordTitleRequired reports that a title is required.
	CodeDynamicDemoRecordTitleRequired = bizerr.MustDefine(
		"PLUGIN_DEMO_DYNAMIC_RECORD_TITLE_REQUIRED",
		"Record title cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeDynamicDemoRecordTitleTooLong reports that a title exceeds the limit.
	CodeDynamicDemoRecordTitleTooLong = bizerr.MustDefine(
		"PLUGIN_DEMO_DYNAMIC_RECORD_TITLE_TOO_LONG",
		"Record title must not exceed {maxChars} characters",
		gcode.CodeInvalidParameter,
	)
	// CodeDynamicDemoRecordContentTooLong reports that content exceeds the limit.
	CodeDynamicDemoRecordContentTooLong = bizerr.MustDefine(
		"PLUGIN_DEMO_DYNAMIC_RECORD_CONTENT_TOO_LONG",
		"Record content must not exceed {maxChars} characters",
		gcode.CodeInvalidParameter,
	)
	// CodeDynamicDemoRecordAttachmentNameRequired reports a missing attachment filename.
	CodeDynamicDemoRecordAttachmentNameRequired = bizerr.MustDefine(
		"PLUGIN_DEMO_DYNAMIC_RECORD_ATTACHMENT_NAME_REQUIRED",
		"Attachment name is required when uploading an attachment",
		gcode.CodeInvalidParameter,
	)
	// CodeDynamicDemoRecordAttachmentBase64Invalid reports an invalid Base64 attachment payload.
	CodeDynamicDemoRecordAttachmentBase64Invalid = bizerr.MustDefine(
		"PLUGIN_DEMO_DYNAMIC_RECORD_ATTACHMENT_BASE64_INVALID",
		"Attachment content must be a valid Base64 string",
		gcode.CodeInvalidParameter,
	)
	// CodeDynamicDemoRecordAttachmentContentRequired reports an empty decoded attachment.
	CodeDynamicDemoRecordAttachmentContentRequired = bizerr.MustDefine(
		"PLUGIN_DEMO_DYNAMIC_RECORD_ATTACHMENT_CONTENT_REQUIRED",
		"Attachment content cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeDynamicDemoRecordAttachmentSizeTooLarge reports an uploaded attachment that exceeds the limit.
	CodeDynamicDemoRecordAttachmentSizeTooLarge = bizerr.MustDefine(
		"PLUGIN_DEMO_DYNAMIC_RECORD_ATTACHMENT_SIZE_TOO_LARGE",
		"Attachment size must not exceed {maxSizeMB}MB",
		gcode.CodeInvalidParameter,
	)
	// CodeDynamicDemoRecordIDRequired reports that a demo record ID is required.
	CodeDynamicDemoRecordIDRequired = bizerr.MustDefine(
		"PLUGIN_DEMO_DYNAMIC_RECORD_ID_REQUIRED",
		"Record ID cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeDynamicDemoRecordNotFound reports that the requested demo record does not exist.
	CodeDynamicDemoRecordNotFound = bizerr.MustDefine(
		"PLUGIN_DEMO_DYNAMIC_RECORD_NOT_FOUND",
		"Dynamic demo record does not exist",
		gcode.CodeNotFound,
	)
	// CodeDynamicDemoRecordAttachmentNotFound reports that the requested attachment is missing.
	CodeDynamicDemoRecordAttachmentNotFound = bizerr.MustDefine(
		"PLUGIN_DEMO_DYNAMIC_RECORD_ATTACHMENT_NOT_FOUND",
		"Dynamic demo record attachment does not exist",
		gcode.CodeNotFound,
	)
	// CodeDynamicDemoRecordAttachmentRollbackFailed reports cleanup failure after a mutation failed.
	CodeDynamicDemoRecordAttachmentRollbackFailed = bizerr.MustDefine(
		"PLUGIN_DEMO_DYNAMIC_RECORD_ATTACHMENT_ROLLBACK_FAILED",
		"Failed to roll back dynamic demo record attachment",
		gcode.CodeInternalError,
	)
)
