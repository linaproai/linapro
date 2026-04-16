package config

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	v1 "lina-core/api/config/v1"
	"lina-core/internal/service/sysconfig"
)

// Export exports config items to an Excel file.
func (c *ControllerV1) Export(ctx context.Context, req *v1.ExportReq) (res *v1.ExportRes, err error) {
	data, err := c.svc.Export(ctx, sysconfig.ExportInput{
		Name:      req.Name,
		Key:       req.Key,
		BeginTime: req.BeginTime,
		EndTime:   req.EndTime,
		Ids:       req.Ids,
	})
	if err != nil {
		return nil, err
	}
	r := g.RequestFromCtx(ctx)
	// Use RFC 5987 format for Chinese filename support
	r.Response.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	r.Response.Header().Set("Content-Disposition", "attachment; filename*=UTF-8''%E5%8F%82%E6%95%B0%E8%AE%BE%E7%BD%AE%E5%AF%BC%E5%87%BA.xlsx")
	r.Response.WriteOver(data)
	r.ExitAll()
	return nil, nil
}
