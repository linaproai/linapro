package dict

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	v1 "lina-core/api/dict/v1"
)

// TypeImportTemplate downloads the dictionary type import template.
func (c *ControllerV1) TypeImportTemplate(ctx context.Context, req *v1.TypeImportTemplateReq) (res *v1.TypeImportTemplateRes, err error) {
	data, err := c.dictSvc.GenerateTypeImportTemplate()
	if err != nil {
		return nil, err
	}

	r := g.RequestFromCtx(ctx)
	r.Response.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	r.Response.Header().Set("Content-Disposition", "attachment; filename=dict-type-import-template.xlsx")
	r.Response.WriteOver(data)
	r.ExitAll()
	return nil, nil
}