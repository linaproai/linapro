// This file implements the v1 dictionary-data import HTTP handler.

package dict

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	v1 "lina-core/api/dict/v1"
	dictsvc "lina-core/internal/service/dict"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/closeutil"
)

// DataImport imports dictionary data from an Excel file.
func (c *ControllerV1) DataImport(ctx context.Context, req *v1.DataImportReq) (res *v1.DataImportRes, err error) {
	r := g.RequestFromCtx(ctx)
	file := r.GetUploadFile("file")
	if file == nil {
		return nil, bizerr.NewCode(dictsvc.CodeDictImportFileRequired)
	}

	// Get updateSupport flag
	updateSupport := r.Get("updateSupport").String() == "1" || r.Get("updateSupport").String() == "true"

	f, err := file.Open()
	if err != nil {
		return nil, bizerr.WrapCode(err, dictsvc.CodeDictImportExcelReadFailed)
	}
	defer closeutil.Close(ctx, f, &err, "close dictionary data import file failed")

	result, err := c.dictSvc.DataImport(ctx, f, updateSupport)
	if err != nil {
		return nil, err
	}

	failList := make([]v1.DataImportFailItem, 0, len(result.FailList))
	for _, item := range result.FailList {
		failList = append(failList, v1.DataImportFailItem{
			Row:    item.Row,
			Reason: item.Reason,
		})
	}

	return &v1.DataImportRes{
		Success:  result.Success,
		Fail:     result.Fail,
		FailList: failList,
	}, nil
}
