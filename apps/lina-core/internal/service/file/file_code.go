// This file defines file-service business error codes and their i18n metadata.

package file

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeFileUploadRequired reports that no upload file was provided.
	CodeFileUploadRequired = bizerr.MustDefine(
		"FILE_UPLOAD_REQUIRED",
		"Please upload a file",
		gcode.CodeInvalidParameter,
	)
	// CodeFileTooLarge reports that an upload exceeds the configured size limit.
	CodeFileTooLarge = bizerr.MustDefine(
		"FILE_TOO_LARGE",
		"File size must not exceed {maxSizeMB}MB",
		gcode.CodeInvalidParameter,
	)
	// CodeFileOpenFailed reports that the uploaded file cannot be opened.
	CodeFileOpenFailed = bizerr.MustDefine(
		"FILE_OPEN_FAILED",
		"Failed to open uploaded file",
		gcode.CodeInvalidParameter,
	)
	// CodeFileHashFailed reports that the uploaded file hash cannot be calculated.
	CodeFileHashFailed = bizerr.MustDefine(
		"FILE_HASH_FAILED",
		"Failed to calculate file hash",
		gcode.CodeInternalError,
	)
	// CodeFileHashQueryFailed reports that duplicate-file detection failed.
	CodeFileHashQueryFailed = bizerr.MustDefine(
		"FILE_HASH_QUERY_FAILED",
		"Failed to query file hash",
		gcode.CodeInternalError,
	)
	// CodeFileRecordSaveFailed reports that file metadata cannot be saved.
	CodeFileRecordSaveFailed = bizerr.MustDefine(
		"FILE_RECORD_SAVE_FAILED",
		"Failed to save file record",
		gcode.CodeInternalError,
	)
	// CodeFileRecordSaveCleanupFailed reports that both file metadata save and stored-file cleanup failed.
	CodeFileRecordSaveCleanupFailed = bizerr.MustDefine(
		"FILE_RECORD_SAVE_CLEANUP_FAILED",
		"Failed to save file record and cleanup stored file",
		gcode.CodeInternalError,
	)
	// CodeFileRecordIDReadFailed reports that the created file record ID cannot be read.
	CodeFileRecordIDReadFailed = bizerr.MustDefine(
		"FILE_RECORD_ID_READ_FAILED",
		"Failed to read created file record ID",
		gcode.CodeInternalError,
	)
	// CodeFileReadResetFailed reports that the upload reader cannot be reset before storage.
	CodeFileReadResetFailed = bizerr.MustDefine(
		"FILE_READ_RESET_FAILED",
		"Failed to reset uploaded file reader",
		gcode.CodeInternalError,
	)
	// CodeFileStoreFailed reports that the uploaded file cannot be stored.
	CodeFileStoreFailed = bizerr.MustDefine(
		"FILE_STORE_FAILED",
		"Failed to store uploaded file",
		gcode.CodeInternalError,
	)
	// CodeFileNotFound reports that the requested file record does not exist.
	CodeFileNotFound = bizerr.MustDefine(
		"FILE_NOT_FOUND",
		"File does not exist",
		gcode.CodeNotFound,
	)
	// CodeFileDeleteRequired reports that no file IDs were provided for deletion.
	CodeFileDeleteRequired = bizerr.MustDefine(
		"FILE_DELETE_REQUIRED",
		"Please select files to delete",
		gcode.CodeInvalidParameter,
	)
)
