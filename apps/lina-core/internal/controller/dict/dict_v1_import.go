// This file implements the shared v1 dictionary import helper used by import
// endpoints.

package dict

import (
	"context"
	"io"

	"github.com/gogf/gf/v2/frame/g"

	v1 "lina-core/api/dict/v1"
	dictsvc "lina-core/internal/service/dict"
	"lina-core/pkg/closeutil"
)

// Import imports dictionary types and data together from an Excel file.
func (c *ControllerV1) Import(ctx context.Context, req *v1.ImportReq) (res *v1.ImportRes, err error) {
	// Get uploaded file
	r := g.RequestFromCtx(ctx)
	file, _, err := r.FormFile("file")
	if err != nil {
		return nil, err
	}
	defer closeutil.Close(file, &err, "关闭字典导入文件失败")

	fileData, err := io.ReadAll(file)
	if err != nil {
		return nil, err
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

// ImportTemplate downloads the combined import template.
func (c *ControllerV1) ImportTemplate(ctx context.Context, req *v1.ImportTemplateReq) (res *v1.ImportTemplateRes, err error) {
	data, err := c.dictSvc.CombinedImportTemplate()
	if err != nil {
		return nil, err
	}

	r := g.RequestFromCtx(ctx)
	r.Response.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	r.Response.Header().Set("Content-Disposition", "attachment; filename*=UTF-8''%E5%AD%97%E5%85%B8%E7%AE%A1%E7%90%86%E5%AF%BC%E5%85%A5%E6%A8%A1%E6%9D%BF.xlsx")
	r.Response.WriteOver(data)
	r.ExitAll()
	return nil, nil
}

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
