package demo

import (
	"context"

	"lina-plugin-demo-source/backend/api/demo/v1"
)

// Summary returns one concise plugin summary for page rendering.
func (c *ControllerV1) Summary(ctx context.Context, _ *v1.SummaryReq) (res *v1.SummaryRes, err error) {
	out, err := c.demoSvc.Summary(ctx)
	if err != nil {
		return nil, err
	}

	return &v1.SummaryRes{
		Message: out.Message,
	}, nil
}
