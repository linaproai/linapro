package dict

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	v1 "lina-core/api/dict/v1"
)

// DataImportTemplate downloads the dictionary data import template.
func (c *ControllerV1) DataImportTemplate(ctx context.Context, req *v1.DataImportTemplateReq) (res *v1.DataImportTemplateRes, err error) {
	data, err := c.dictSvc.GenerateDataImportTemplate()
	if err != nil {
		return nil, err
	}

	r := g.RequestFromCtx(ctx)
	r.Response.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	r.Response.Header().Set("Content-Disposition", "attachment; filename=dict-data-import-template.xlsx")
	r.Response.WriteOver(data)
	r.ExitAll()
	return nil, nil
}