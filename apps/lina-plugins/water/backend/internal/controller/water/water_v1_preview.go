package water

import (
	"context"

	"lina-plugin-water/backend/api/water/v1"
	watersvc "lina-plugin-water/backend/internal/service/water"
)

// Preview synchronously renders one watermark preview.
func (c *ControllerV1) Preview(ctx context.Context, req *v1.PreviewReq) (res *v1.PreviewRes, err error) {
	out, err := c.waterSvc.Preview(ctx, watersvc.PreviewInput{
		Tenant:      req.Tenant,
		DeviceId:    req.DeviceId,
		DeviceCode:  req.DeviceCode,
		ChannelCode: req.ChannelCode,
		Image:       req.Image,
	})
	if err != nil {
		return nil, err
	}
	return &v1.PreviewRes{
		Success:      out.Success,
		Status:       string(out.Status),
		Message:      out.Message,
		Image:        out.Image,
		StrategyId:   out.StrategyId,
		StrategyName: out.StrategyName,
		Source:       string(out.Source),
		SourceLabel:  out.SourceLabel,
		DurationMs:   out.DurationMs,
	}, nil
}
