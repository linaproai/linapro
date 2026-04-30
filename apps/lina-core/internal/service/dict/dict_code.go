// This file defines dictionary-service business error codes and their i18n
// metadata.

package dict

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeDictTypeExists reports that a dictionary type already exists.
	CodeDictTypeExists = bizerr.MustDefine(
		"DICT_TYPE_EXISTS",
		"Dictionary type already exists",
		gcode.CodeInvalidParameter,
	)
	// CodeDictTypeNotFound reports that the requested dictionary type does not exist.
	CodeDictTypeNotFound = bizerr.MustDefine(
		"DICT_TYPE_NOT_FOUND",
		"Dictionary type does not exist",
		gcode.CodeNotFound,
	)
	// CodeDictTypeBuiltinDeleteDenied reports that a built-in dictionary type cannot be deleted.
	CodeDictTypeBuiltinDeleteDenied = bizerr.MustDefine(
		"DICT_TYPE_BUILTIN_DELETE_DENIED",
		"Built-in dictionary types cannot be deleted",
		gcode.CodeNotAuthorized,
	)
	// CodeDictDataNotFound reports that the requested dictionary data entry does not exist.
	CodeDictDataNotFound = bizerr.MustDefine(
		"DICT_DATA_NOT_FOUND",
		"Dictionary data does not exist",
		gcode.CodeNotFound,
	)
	// CodeDictDataBuiltinDeleteDenied reports that built-in dictionary data cannot be deleted.
	CodeDictDataBuiltinDeleteDenied = bizerr.MustDefine(
		"DICT_DATA_BUILTIN_DELETE_DENIED",
		"Built-in dictionary data cannot be deleted",
		gcode.CodeNotAuthorized,
	)
	// CodeDictImportExcelParseFailed reports that an uploaded dictionary workbook cannot be parsed.
	CodeDictImportExcelParseFailed = bizerr.MustDefine(
		"DICT_IMPORT_EXCEL_PARSE_FAILED",
		"Failed to parse Excel file",
		gcode.CodeInvalidParameter,
	)
	// CodeDictImportExcelReadFailed reports that a dictionary workbook cannot be read.
	CodeDictImportExcelReadFailed = bizerr.MustDefine(
		"DICT_IMPORT_EXCEL_READ_FAILED",
		"Failed to read Excel file",
		gcode.CodeInvalidParameter,
	)
	// CodeDictImportFileRequired reports that no dictionary import file was uploaded.
	CodeDictImportFileRequired = bizerr.MustDefine(
		"DICT_IMPORT_FILE_REQUIRED",
		"Please select a file to import",
		gcode.CodeInvalidParameter,
	)
)
