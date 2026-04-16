package loginlog

import (
	"context"

	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"

	v1 "lina-core/api/loginlog/v1"
)

// Delete deletes login logs
func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error) {
	idStrs := gstr.SplitAndTrim(req.Ids, ",")
	ids := make([]int, 0, len(idStrs))
	for _, s := range idStrs {
		ids = append(ids, gconv.Int(s))
	}
	deleted, err := c.loginLogSvc.DeleteByIds(ctx, ids)
	if err != nil {
		return nil, err
	}
	return &v1.DeleteRes{Deleted: deleted}, nil
}
