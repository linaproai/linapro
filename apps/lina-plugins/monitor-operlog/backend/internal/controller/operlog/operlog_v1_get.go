package operlog

import (
	"context"

	v1 "lina-plugin-monitor-operlog/backend/api/operlog/v1"
)

// Get returns operation-log details by ID.
func (c *ControllerV1) Get(ctx context.Context, req *v1.GetReq) (res *v1.GetRes, err error) {
	record, err := c.operLogSvc.GetById(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.GetRes{OperLogEntity: toAPIOperLogEntity(record)}, nil
}
