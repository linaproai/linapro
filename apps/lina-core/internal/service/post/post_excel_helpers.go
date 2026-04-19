// This file centralizes Excel generation helpers for post exports.

package post

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/xuri/excelize/v2"

	"lina-core/pkg/excelutil"
)

// closeExcelFile closes one Excel workbook while preserving the primary error path.
func closeExcelFile(file *excelize.File, errPtr *error) {
	excelutil.CloseFile(file, errPtr)
}

// setCellValue proxies cell writes through the shared Excel helper package.
func setCellValue(file *excelize.File, sheet string, col int, row int, value any) error {
	return excelutil.SetCellValue(file, sheet, col, row, value)
}

// setCellValueByName writes one cell identified by A1-style coordinates.
func setCellValueByName(file *excelize.File, sheet string, cell string, value any) error {
	return excelutil.SetCellValueByName(file, sheet, cell, value)
}

// cellName converts numeric coordinates into an Excel A1-style cell reference.
func cellName(col int, row int) string {
	name, err := excelize.CoordinatesToCellName(col, row)
	if err != nil {
		panic(gerror.Wrap(err, "生成Excel单元格名称失败"))
	}
	return name
}
