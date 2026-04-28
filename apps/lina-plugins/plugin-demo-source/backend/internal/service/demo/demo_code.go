// This file defines source-demo business error codes and runtime i18n metadata.

package demo

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeDemoRecordCountQueryFailed reports that the demo record count query failed.
	CodeDemoRecordCountQueryFailed = bizerr.MustDefine(
		"PLUGIN_DEMO_SOURCE_RECORD_COUNT_QUERY_FAILED",
		"Failed to query demo record count",
		gcode.CodeInternalError,
	)
	// CodeDemoRecordListQueryFailed reports that the demo record list query failed.
	CodeDemoRecordListQueryFailed = bizerr.MustDefine(
		"PLUGIN_DEMO_SOURCE_RECORD_LIST_QUERY_FAILED",
		"Failed to query demo record list",
		gcode.CodeInternalError,
	)
	// CodeDemoRecordCreateFailed reports that creating a demo record failed.
	CodeDemoRecordCreateFailed = bizerr.MustDefine(
		"PLUGIN_DEMO_SOURCE_RECORD_CREATE_FAILED",
		"Failed to create demo record",
		gcode.CodeInternalError,
	)
	// CodeDemoRecordUpdateFailed reports that updating a demo record failed.
	CodeDemoRecordUpdateFailed = bizerr.MustDefine(
		"PLUGIN_DEMO_SOURCE_RECORD_UPDATE_FAILED",
		"Failed to update demo record",
		gcode.CodeInternalError,
	)
	// CodeDemoRecordDeleteFailed reports that deleting a demo record failed.
	CodeDemoRecordDeleteFailed = bizerr.MustDefine(
		"PLUGIN_DEMO_SOURCE_RECORD_DELETE_FAILED",
		"Failed to delete demo record",
		gcode.CodeInternalError,
	)
	// CodeDemoRecordAttachmentRequired reports that the current record has no attachment.
	CodeDemoRecordAttachmentRequired = bizerr.MustDefine(
		"PLUGIN_DEMO_SOURCE_RECORD_ATTACHMENT_REQUIRED",
		"Current record has no attachment",
		gcode.CodeNotFound,
	)
	// CodeDemoRecordAttachmentFileNotFound reports that the stored attachment file is missing.
	CodeDemoRecordAttachmentFileNotFound = bizerr.MustDefine(
		"PLUGIN_DEMO_SOURCE_RECORD_ATTACHMENT_FILE_NOT_FOUND",
		"Attachment file does not exist",
		gcode.CodeNotFound,
	)
	// CodeDemoRecordIDRequired reports that a demo record ID is required.
	CodeDemoRecordIDRequired = bizerr.MustDefine(
		"PLUGIN_DEMO_SOURCE_RECORD_ID_REQUIRED",
		"Record ID cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeDemoRecordDetailQueryFailed reports that querying one demo record failed.
	CodeDemoRecordDetailQueryFailed = bizerr.MustDefine(
		"PLUGIN_DEMO_SOURCE_RECORD_DETAIL_QUERY_FAILED",
		"Failed to query demo record detail",
		gcode.CodeInternalError,
	)
	// CodeDemoRecordNotFound reports that the requested demo record does not exist.
	CodeDemoRecordNotFound = bizerr.MustDefine(
		"PLUGIN_DEMO_SOURCE_RECORD_NOT_FOUND",
		"Demo record does not exist",
		gcode.CodeNotFound,
	)
	// CodeDemoRecordTitleRequired reports that a demo record title is required.
	CodeDemoRecordTitleRequired = bizerr.MustDefine(
		"PLUGIN_DEMO_SOURCE_RECORD_TITLE_REQUIRED",
		"Record title cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeDemoRecordAttachmentPathQueryFailed reports that querying attachment paths failed.
	CodeDemoRecordAttachmentPathQueryFailed = bizerr.MustDefine(
		"PLUGIN_DEMO_SOURCE_RECORD_ATTACHMENT_PATH_QUERY_FAILED",
		"Failed to query demo attachment paths",
		gcode.CodeInternalError,
	)
	// CodeDemoRecordTableCheckFailed reports that checking the demo table failed.
	CodeDemoRecordTableCheckFailed = bizerr.MustDefine(
		"PLUGIN_DEMO_SOURCE_TABLE_CHECK_FAILED",
		"Failed to inspect the demo record table",
		gcode.CodeInternalError,
	)
	// CodeDemoRecordTableNotInstalled reports that the demo table is missing.
	CodeDemoRecordTableNotInstalled = bizerr.MustDefine(
		"PLUGIN_DEMO_SOURCE_TABLE_NOT_INSTALLED",
		"Demo record table does not exist. Install the plugin first",
		gcode.CodeNotFound,
	)
	// CodeDemoAttachmentStoragePurgeFailed reports that removing attachment storage failed.
	CodeDemoAttachmentStoragePurgeFailed = bizerr.MustDefine(
		"PLUGIN_DEMO_SOURCE_ATTACHMENT_STORAGE_PURGE_FAILED",
		"Failed to remove demo attachment storage",
		gcode.CodeInternalError,
	)
	// CodeDemoAttachmentOpenFailed reports that opening an uploaded attachment failed.
	CodeDemoAttachmentOpenFailed = bizerr.MustDefine(
		"PLUGIN_DEMO_SOURCE_ATTACHMENT_OPEN_FAILED",
		"Failed to open demo attachment",
		gcode.CodeInvalidParameter,
	)
	// CodeDemoAttachmentSourceCloseFailed reports that closing an uploaded attachment failed.
	CodeDemoAttachmentSourceCloseFailed = bizerr.MustDefine(
		"PLUGIN_DEMO_SOURCE_ATTACHMENT_SOURCE_CLOSE_FAILED",
		"Failed to close demo attachment source",
		gcode.CodeInternalError,
	)
	// CodeDemoAttachmentDirCreateFailed reports that creating the attachment directory failed.
	CodeDemoAttachmentDirCreateFailed = bizerr.MustDefine(
		"PLUGIN_DEMO_SOURCE_ATTACHMENT_DIR_CREATE_FAILED",
		"Failed to create demo attachment directory",
		gcode.CodeInternalError,
	)
	// CodeDemoAttachmentCreateFailed reports that creating the attachment file failed.
	CodeDemoAttachmentCreateFailed = bizerr.MustDefine(
		"PLUGIN_DEMO_SOURCE_ATTACHMENT_CREATE_FAILED",
		"Failed to create demo attachment file",
		gcode.CodeInternalError,
	)
	// CodeDemoAttachmentTargetCloseFailed reports that closing the stored attachment failed.
	CodeDemoAttachmentTargetCloseFailed = bizerr.MustDefine(
		"PLUGIN_DEMO_SOURCE_ATTACHMENT_TARGET_CLOSE_FAILED",
		"Failed to close demo attachment file",
		gcode.CodeInternalError,
	)
	// CodeDemoAttachmentWriteFailed reports that writing an attachment failed.
	CodeDemoAttachmentWriteFailed = bizerr.MustDefine(
		"PLUGIN_DEMO_SOURCE_ATTACHMENT_WRITE_FAILED",
		"Failed to write demo attachment",
		gcode.CodeInternalError,
	)
	// CodeDemoAttachmentDeleteFailed reports that deleting an attachment failed.
	CodeDemoAttachmentDeleteFailed = bizerr.MustDefine(
		"PLUGIN_DEMO_SOURCE_ATTACHMENT_DELETE_FAILED",
		"Failed to delete demo attachment file",
		gcode.CodeInternalError,
	)
	// CodeDemoAttachmentSizeTooLarge reports that an uploaded attachment is too large.
	CodeDemoAttachmentSizeTooLarge = bizerr.MustDefine(
		"PLUGIN_DEMO_SOURCE_ATTACHMENT_SIZE_TOO_LARGE",
		"Attachment size must not exceed {maxSizeMB}MB",
		gcode.CodeInvalidParameter,
	)
	// CodeDemoAttachmentPartialCleanupFailed reports that cleanup after a failed write failed.
	CodeDemoAttachmentPartialCleanupFailed = bizerr.MustDefine(
		"PLUGIN_DEMO_SOURCE_ATTACHMENT_PARTIAL_CLEANUP_FAILED",
		"Failed to remove incomplete demo attachment",
		gcode.CodeInternalError,
	)
)
