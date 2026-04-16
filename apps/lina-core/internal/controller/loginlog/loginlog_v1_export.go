package loginlog

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	v1 "lina-core/api/loginlog/v1"
	loginlogsvc "lina-core/internal/service/loginlog"
)

// Export exports login logs
func (c *ControllerV1) Export(ctx context.Context, req *v1.ExportReq) (res *v1.ExportRes, err error) {
	data, err := c.loginLogSvc.Export(ctx, loginlogsvc.ExportInput{
		UserName:       req.UserName,
		Ip:             req.Ip,
		Status:         req.Status,
		BeginTime:      req.BeginTime,
		EndTime:        req.EndTime,
		OrderBy:        req.OrderBy,
		OrderDirection: req.OrderDirection,
		Ids:            req.Ids,
	})
	if err != nil {
		return nil, err
	}

	r := g.RequestFromCtx(ctx)
	// Use RFC 5987 format for Chinese filename support
	r.Response.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	r.Response.Header().Set("Content-Disposition", "attachment; filename*=UTF-8''%E7%99%BB%E5%BD%95%E6%97%A5%E5%BF%97%E5%AF%BC%E5%87%BA.xlsx")
	r.Response.WriteOver(data)
	r.ExitAll()
	return nil, nil
}
