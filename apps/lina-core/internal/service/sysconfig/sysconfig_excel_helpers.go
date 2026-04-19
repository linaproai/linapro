// This file centralizes Excel generation helpers for system-config imports and exports.

package sysconfig

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/xuri/excelize/v2"

	"lina-core/pkg/excelutil"
)

// closeExcelFile closes the workbook and folds any close failure into the
// caller-managed error pointer.
func closeExcelFile(file *excelize.File, errPtr *error) {
	excelutil.CloseFile(file, errPtr)
}

// setCellValue writes one value by row and column coordinates.
func setCellValue(file *excelize.File, sheet string, col int, row int, value any) error {
	return excelutil.SetCellValue(file, sheet, col, row, value)
}

// setCellValueByName writes one value by Excel-style cell name.
func setCellValueByName(file *excelize.File, sheet string, cell string, value any) error {
	return excelutil.SetCellValueByName(file, sheet, cell, value)
}

// cellName converts one row and column coordinate pair into an Excel cell
// identifier.
func cellName(col int, row int) string {
	name, err := excelize.CoordinatesToCellName(col, row)
	if err != nil {
		panic(gerror.Wrap(err, "生成Excel单元格名称失败"))
	}
	return name
}
