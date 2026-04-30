// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package health

import (
	"context"

	"lina-core/api/health/v1"
)

type IHealthV1 interface {
	Get(ctx context.Context, req *v1.GetReq) (res *v1.GetRes, err error)
}
