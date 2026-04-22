// post_v1_export.go implements the controller method that exports posts.

package post

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	v1 "lina-plugin-org-center/backend/api/post/v1"
	postsvc "lina-plugin-org-center/backend/internal/service/post"
)

// Export exports posts.
func (c *ControllerV1) Export(ctx context.Context, req *v1.ExportReq) (res *v1.ExportRes, err error) {
	data, err := c.postSvc.Export(ctx, postsvc.ExportInput{
		DeptId: req.DeptId,
		Code:   req.Code,
		Name:   req.Name,
		Status: req.Status,
	})
	if err != nil {
		return nil, err
	}
	r := g.RequestFromCtx(ctx)
	r.Response.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	r.Response.Header().Set("Content-Disposition", "attachment; filename*=UTF-8''%E5%B2%97%E4%BD%8D%E6%95%B0%E6%8D%AE%E5%AF%BC%E5%87%BA.xlsx")
	r.Response.WriteOver(data)
	r.ExitAll()
	return nil, nil
}
