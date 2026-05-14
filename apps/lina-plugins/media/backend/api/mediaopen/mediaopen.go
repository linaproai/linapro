// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package mediaopen

import (
	"context"

	"lina-plugin-media/backend/api/mediaopen/v1"
)

type IMediaopenV1 interface {
	ResolveStrategyByToken(ctx context.Context, req *v1.ResolveStrategyByTokenReq) (res *v1.ResolveStrategyByTokenRes, err error)
}
