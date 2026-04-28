// This file defines system-configuration business error codes and their i18n
// metadata.

package sysconfig

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeSysConfigNotFound reports that the requested configuration record does not exist.
	CodeSysConfigNotFound = bizerr.MustDefine(
		"SYSCONFIG_NOT_FOUND",
		"System configuration does not exist",
		gcode.CodeNotFound,
	)
	// CodeSysConfigKeyExists reports that the requested configuration key already exists.
	CodeSysConfigKeyExists = bizerr.MustDefine(
		"SYSCONFIG_KEY_EXISTS",
		"System configuration key {key} already exists",
		gcode.CodeInvalidParameter,
	)
	// CodeSysConfigKeyNotFound reports that the requested configuration key does not exist.
	CodeSysConfigKeyNotFound = bizerr.MustDefine(
		"SYSCONFIG_KEY_NOT_FOUND",
		"System configuration key does not exist",
		gcode.CodeNotFound,
	)
	// CodeSysConfigBuiltinKeyRenameDenied reports that a protected runtime key cannot be renamed.
	CodeSysConfigBuiltinKeyRenameDenied = bizerr.MustDefine(
		"SYSCONFIG_BUILTIN_KEY_RENAME_DENIED",
		"Built-in runtime configuration keys cannot be renamed",
		gcode.CodeNotAuthorized,
	)
	// CodeSysConfigBuiltinDeleteDenied reports that a protected runtime parameter cannot be deleted.
	CodeSysConfigBuiltinDeleteDenied = bizerr.MustDefine(
		"SYSCONFIG_BUILTIN_DELETE_DENIED",
		"Built-in runtime configurations cannot be deleted",
		gcode.CodeNotAuthorized,
	)
	// CodeSysConfigProtectedValueInvalid reports that a protected runtime parameter value is invalid.
	CodeSysConfigProtectedValueInvalid = bizerr.MustDefine(
		"SYSCONFIG_PROTECTED_VALUE_INVALID",
		"Built-in system configuration value is invalid",
		gcode.CodeInvalidParameter,
	)
	// CodeSysConfigImportFileRequired reports that no configuration import file was uploaded.
	CodeSysConfigImportFileRequired = bizerr.MustDefine(
		"SYSCONFIG_IMPORT_FILE_REQUIRED",
		"Please select a file to import",
		gcode.CodeInvalidParameter,
	)
	// CodeSysConfigImportFileReadFailed reports that an uploaded configuration file cannot be read.
	CodeSysConfigImportFileReadFailed = bizerr.MustDefine(
		"SYSCONFIG_IMPORT_FILE_READ_FAILED",
		"Failed to read import file",
		gcode.CodeInvalidParameter,
	)
	// CodeSysConfigImportExcelParseFailed reports that the uploaded configuration workbook cannot be parsed.
	CodeSysConfigImportExcelParseFailed = bizerr.MustDefine(
		"SYSCONFIG_IMPORT_EXCEL_PARSE_FAILED",
		"Failed to parse Excel file",
		gcode.CodeInvalidParameter,
	)
	// CodeSysConfigImportSheetReadFailed reports that the configuration import worksheet cannot be read.
	CodeSysConfigImportSheetReadFailed = bizerr.MustDefine(
		"SYSCONFIG_IMPORT_SHEET_READ_FAILED",
		"Failed to read worksheet {sheet}",
		gcode.CodeInvalidParameter,
	)
	// CodeSysConfigImportQueryFailed reports that import failed while querying existing configuration records.
	CodeSysConfigImportQueryFailed = bizerr.MustDefine(
		"SYSCONFIG_IMPORT_QUERY_FAILED",
		"Failed to query system configuration during import",
		gcode.CodeInternalError,
	)
	// CodeSysConfigImportUpdateFailed reports that import failed while updating an existing record.
	CodeSysConfigImportUpdateFailed = bizerr.MustDefine(
		"SYSCONFIG_IMPORT_UPDATE_FAILED",
		"Failed to update system configuration during import",
		gcode.CodeInternalError,
	)
	// CodeSysConfigImportInsertFailed reports that import failed while creating a new record.
	CodeSysConfigImportInsertFailed = bizerr.MustDefine(
		"SYSCONFIG_IMPORT_INSERT_FAILED",
		"Failed to insert system configuration during import",
		gcode.CodeInternalError,
	)
)
