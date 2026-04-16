package user

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	v1 "lina-core/api/user/v1"
)

// ImportTemplate downloads user import template
func (c *ControllerV1) ImportTemplate(ctx context.Context, req *v1.ImportTemplateReq) (res *v1.ImportTemplateRes, err error) {
	data, err := c.userSvc.GenerateImportTemplate()
	if err != nil {
		return nil, err
	}

	r := g.RequestFromCtx(ctx)
	r.Response.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	r.Response.Header().Set("Content-Disposition", "attachment; filename=user-import-template.xlsx")
	r.Response.WriteOver(data)
	r.ExitAll()
	return nil, nil
}
