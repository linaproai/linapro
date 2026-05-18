// Package excelutil provides shared helpers for generating Excel files with
// explicit error handling.
package excelutil

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/xuri/excelize/v2"

	"lina-core/pkg/logger"
)

// CloseFile closes one Excel file and folds any close failure into errPtr.
func CloseFile(ctx context.Context, file *excelize.File, errPtr *error) {
	if file == nil {
		return
	}
	closeErr := file.Close()
	if closeErr == nil {
		return
	}
	wrapped := gerror.Wrap(closeErr, "close Excel file failed")
	if errPtr == nil {
		// A nil error pointer means the caller misused this helper by omitting
		// the named return error path, so log the close failure instead of
		// panicking or silently dropping it.
		logger.Warningf(ctx, "excel close failed without error return path err=%v", wrapped)
		return
	}
	if *errPtr == nil {
		// Keep any earlier business error intact and only report the close error
		// when it is the first failure in the call chain.
		*errPtr = wrapped
	}
}

// SetSheetName renames one Excel worksheet and returns a wrapped error when it fails.
func SetSheetName(file *excelize.File, source string, target string) error {
	if err := file.SetSheetName(source, target); err != nil {
		return gerror.Wrapf(err, "rename Excel worksheet failed source=%s target=%s", source, target)
	}
	return nil
}

// NewSheet creates one Excel worksheet and returns a wrapped error when it fails.
func NewSheet(file *excelize.File, name string) error {
	if _, err := file.NewSheet(name); err != nil {
		return gerror.Wrapf(err, "create Excel worksheet failed sheet=%s", name)
	}
	return nil
}

// SetCellValue writes one Excel cell addressed by coordinates.
func SetCellValue(file *excelize.File, sheet string, col int, row int, value any) error {
	cell, err := excelize.CoordinatesToCellName(col, row)
	if err != nil {
		return gerror.Wrap(err, "build Excel cell name failed")
	}
	return SetCellValueByName(file, sheet, cell, value)
}

// SetCellValueByName writes one Excel cell addressed by A1 notation.
func SetCellValueByName(file *excelize.File, sheet string, cell string, value any) error {
	if err := file.SetCellValue(sheet, cell, value); err != nil {
		return gerror.Wrapf(err, "write Excel cell failed sheet=%s cell=%s", sheet, cell)
	}
	return nil
}
