// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package joblog

import (
	"context"

	"lina-core/api/joblog/v1"
)

type IJoblogV1 interface {
	Cancel(ctx context.Context, req *v1.CancelReq) (res *v1.CancelRes, err error)
	Clear(ctx context.Context, req *v1.ClearReq) (res *v1.ClearRes, err error)
	Detail(ctx context.Context, req *v1.DetailReq) (res *v1.DetailRes, err error)
	List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error)
}
