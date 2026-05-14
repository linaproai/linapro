package water

import (
	"context"

	"lina-plugin-water/backend/api/water/v1"
)

// GetTask returns one recent watermark task status.
func (c *ControllerV1) GetTask(ctx context.Context, req *v1.GetTaskReq) (res *v1.GetTaskRes, err error) {
	out, err := c.waterSvc.GetTask(ctx, req.TaskId)
	if err != nil {
		return nil, err
	}
	return &v1.GetTaskRes{
		TaskId:       out.TaskId,
		Status:       string(out.Status),
		Success:      out.Success,
		Message:      out.Message,
		Error:        out.Error,
		Tenant:       out.Tenant,
		DeviceId:     out.DeviceId,
		StrategyId:   out.StrategyId,
		StrategyName: out.StrategyName,
		Source:       string(out.Source),
		SourceLabel:  out.SourceLabel,
		Image:        out.Image,
		CreatedAt:    out.CreatedAt,
		UpdatedAt:    out.UpdatedAt,
		DurationMs:   out.DurationMs,
	}, nil
}
