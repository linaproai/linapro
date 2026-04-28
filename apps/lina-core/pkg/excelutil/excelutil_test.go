// This file verifies Excel helper errors remain explicit return values.

package excelutil

import (
	"context"
	"testing"

	"github.com/xuri/excelize/v2"
)

// TestSetCellValueReturnsCoordinateError verifies invalid coordinates return an
// error instead of panicking inside export helpers.
func TestSetCellValueReturnsCoordinateError(t *testing.T) {
	file := excelize.NewFile()
	defer CloseFile(context.Background(), file, nil)

	if err := SetCellValue(file, "Sheet1", 0, 1, "bad"); err == nil {
		t.Fatal("expected invalid Excel column coordinate to return an error")
	}
}
