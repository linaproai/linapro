// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package demo

import (
	"context"

	"lina-plugin-demo-source/backend/api/demo/v1"
)

type IDemoV1 interface {
	Ping(ctx context.Context, req *v1.PingReq) (res *v1.PingRes, err error)
	Summary(ctx context.Context, req *v1.SummaryReq) (res *v1.SummaryRes, err error)
}
