// This file implements the combined v1 dictionary import-template HTTP handler.

package dict

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	v1 "lina-core/api/dict/v1"
)

// ImportTemplate downloads the combined import template.
func (c *ControllerV1) ImportTemplate(ctx context.Context, req *v1.ImportTemplateReq) (res *v1.ImportTemplateRes, err error) {
	data, err := c.dictSvc.CombinedImportTemplate(ctx)
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
