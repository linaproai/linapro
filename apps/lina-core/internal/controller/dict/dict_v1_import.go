// This file implements the combined v1 dictionary import HTTP handler.

package dict

import (
	"context"
	"io"

	"github.com/gogf/gf/v2/frame/g"

	v1 "lina-core/api/dict/v1"
	dictsvc "lina-core/internal/service/dict"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/closeutil"
)

// Import imports dictionary types and data together from an Excel file.
func (c *ControllerV1) Import(ctx context.Context, req *v1.ImportReq) (res *v1.ImportRes, err error) {
	// Get uploaded file
	r := g.RequestFromCtx(ctx)
	uploadFile := r.GetUploadFile("file")
	if uploadFile == nil {
		return nil, bizerr.NewCode(dictsvc.CodeDictImportFileRequired)
	}

	file, err := uploadFile.Open()
	if err != nil {
		return nil, bizerr.WrapCode(err, dictsvc.CodeDictImportExcelReadFailed)
	}
	defer closeutil.Close(ctx, file, &err, "close dictionary import file failed")

	fileData, err := io.ReadAll(file)
	if err != nil {
		return nil, bizerr.WrapCode(err, dictsvc.CodeDictImportExcelReadFailed)
	}

	// Get updateSupport flag
	updateSupport := r.Get("updateSupport").String() == "1" || r.Get("updateSupport").String() == "true"

	result, err := c.dictSvc.CombinedImport(ctx, fileData, updateSupport)
	if err != nil {
		return nil, err
	}

	return &v1.ImportRes{
		TypeSuccess: result.TypeSuccess,
		TypeFail:    result.TypeFail,
		DataSuccess: result.DataSuccess,
		DataFail:    result.DataFail,
		FailList:    convertFailList(result.FailList),
	}, nil
}

// convertFailList converts service-layer import failures into the API response
// shape returned by dictionary import endpoints.
func convertFailList(items []dictsvc.ImportFailItem) []v1.ImportFailItem {
	result := make([]v1.ImportFailItem, len(items))
	for i, item := range items {
		result[i] = v1.ImportFailItem{
			Sheet:  item.Sheet,
			Row:    item.Row,
			Reason: item.Reason,
		}
	}
	return result
}
