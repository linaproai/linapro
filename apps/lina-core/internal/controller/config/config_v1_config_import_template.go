package config

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	v1 "lina-core/api/config/v1"
)

// ConfigImportTemplate downloads the config import template.
func (c *ControllerV1) ConfigImportTemplate(ctx context.Context, req *v1.ConfigImportTemplateReq) (res *v1.ConfigImportTemplateRes, err error) {
	data, err := c.svc.GenerateImportTemplate()
	if err != nil {
		return nil, err
	}

	r := g.RequestFromCtx(ctx)
	r.Response.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	r.Response.Header().Set("Content-Disposition", "attachment; filename=config-import-template.xlsx")
	r.Response.WriteOver(data)
	r.ExitAll()
	return nil, nil
}