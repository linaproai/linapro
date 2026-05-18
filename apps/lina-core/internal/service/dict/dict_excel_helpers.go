// This file centralizes Excel generation helpers for dictionary imports and exports.

package dict

import (
	"context"

	"github.com/xuri/excelize/v2"

	"lina-core/pkg/excelutil"
)

// closeExcelFile closes one Excel workbook while preserving the primary error path.
func closeExcelFile(ctx context.Context, file *excelize.File, errPtr *error) {
	excelutil.CloseFile(ctx, file, errPtr)
}

// setCellValue proxies cell writes through the shared Excel helper package.
func setCellValue(file *excelize.File, sheet string, col int, row int, value any) error {
	return excelutil.SetCellValue(file, sheet, col, row, value)
}

// setCellValueByName writes one cell identified by A1-style coordinates.
func setCellValueByName(file *excelize.File, sheet string, cell string, value any) error {
	return excelutil.SetCellValueByName(file, sheet, cell, value)
}

// setSheetName renames one worksheet through the shared Excel helper package.
func setSheetName(file *excelize.File, source string, target string) error {
	return excelutil.SetSheetName(file, source, target)
}

// newSheet creates one worksheet through the shared Excel helper package.
func newSheet(file *excelize.File, name string) error {
	return excelutil.NewSheet(file, name)
}
