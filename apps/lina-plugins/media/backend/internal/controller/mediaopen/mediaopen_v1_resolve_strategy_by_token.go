// This file implements the public Tieta-token strategy resolution endpoint.

package mediaopen

import (
	"context"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-plugin-media/backend/api/mediaopen/v1"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// ResolveStrategyByToken resolves one media strategy after validating a Tieta token.
func (c *ControllerV1) ResolveStrategyByToken(
	ctx context.Context,
	req *v1.ResolveStrategyByTokenReq,
) (res *v1.ResolveStrategyByTokenRes, err error) {
	authorization := ""
	if request := ghttp.RequestFromCtx(ctx); request != nil {
		authorization = request.GetHeader("Authorization")
	}
	out, err := c.mediaSvc.ResolveStrategyByToken(ctx, mediasvc.ResolveStrategyByTokenInput{
		Token:         req.Token,
		Authorization: authorization,
		TenantId:      req.TenantId,
		DeviceId:      req.DeviceId,
	})
	if err != nil {
		return nil, err
	}
	return &v1.ResolveStrategyByTokenRes{
		UserId:       out.UserId,
		Username:     out.Username,
		RealName:     out.RealName,
		Mobile:       out.Mobile,
		TenantId:     out.TenantId,
		DeviceId:     out.DeviceId,
		HasAccess:    out.HasAccess,
		Matched:      out.Matched,
		Source:       out.Source,
		SourceLabel:  out.SourceLabel,
		StrategyId:   out.StrategyId,
		StrategyName: out.StrategyName,
		Strategy:     out.Strategy,
	}, nil
}
