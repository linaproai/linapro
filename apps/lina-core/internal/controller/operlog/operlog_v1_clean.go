package operlog

import (
	"context"

	v1 "lina-core/api/operlog/v1"
	operlogsvc "lina-core/internal/service/operlog"
)

// Clean clears operation logs
func (c *ControllerV1) Clean(ctx context.Context, req *v1.CleanReq) (res *v1.CleanRes, err error) {
	deleted, err := c.operLogSvc.Clean(ctx, operlogsvc.CleanInput{
		BeginTime: req.BeginTime,
		EndTime:   req.EndTime,
	})
	if err != nil {
		return nil, err
	}
	return &v1.CleanRes{Deleted: deleted}, nil
}
