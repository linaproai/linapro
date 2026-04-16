package demo

import (
	"context"

	"lina-plugin-demo-source/backend/api/demo/v1"
)

// Ping returns one anonymous ping payload for public route-group verification.
func (c *ControllerV1) Ping(ctx context.Context, _ *v1.PingReq) (res *v1.PingRes, err error) {
	out, err := c.demoSvc.Ping(ctx)
	if err != nil {
		return nil, err
	}

	return &v1.PingRes{
		Message: out.Message,
	}, nil
}
