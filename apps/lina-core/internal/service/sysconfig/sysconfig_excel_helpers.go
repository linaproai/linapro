// This file centralizes Excel generation helpers for system-config imports and exports.

package sysconfig

import (
	"context"

	"github.com/xuri/excelize/v2"

	"lina-core/pkg/excelutil"
)

// closeExcelFile closes the workbook and folds any close failure into the
// caller-managed error pointer.
func closeExcelFile(ctx context.Context, file *excelize.File, errPtr *error) {
	excelutil.CloseFile(ctx, file, errPtr)
}

// setCellValue writes one value by row and column coordinates.
func setCellValue(file *excelize.File, sheet string, col int, row int, value any) error {
	return excelutil.SetCellValue(file, sheet, col, row, value)
}

// setCellValueByName writes one value by Excel-style cell name.
func setCellValueByName(file *excelize.File, sheet string, cell string, value any) error {
	return excelutil.SetCellValueByName(file, sheet, cell, value)
}
