package user

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	v1 "lina-core/api/user/v1"
	usersvc "lina-core/internal/service/user"
)

// Export exports users
func (c *ControllerV1) Export(ctx context.Context, req *v1.ExportReq) (res *v1.ExportRes, err error) {
	data, err := c.userSvc.Export(ctx, usersvc.ExportInput{
		Ids: req.Ids,
	})
	if err != nil {
		return nil, err
	}

	r := g.RequestFromCtx(ctx)
	// Use RFC 5987 format for Chinese filename support
	r.Response.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	r.Response.Header().Set("Content-Disposition", "attachment; filename*=UTF-8''%E7%94%A8%E6%88%B7%E6%95%B0%E6%8D%AE%E5%AF%BC%E5%87%BA.xlsx")
	r.Response.WriteOver(data)
	r.ExitAll()
	return nil, nil
}
