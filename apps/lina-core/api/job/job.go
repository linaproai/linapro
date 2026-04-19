// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package job

import (
	"context"

	"lina-core/api/job/v1"
)

type IJobV1 interface {
	Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error)
	CronPreview(ctx context.Context, req *v1.CronPreviewReq) (res *v1.CronPreviewRes, err error)
	Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error)
	Detail(ctx context.Context, req *v1.DetailReq) (res *v1.DetailRes, err error)
	List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error)
	Reset(ctx context.Context, req *v1.ResetReq) (res *v1.ResetRes, err error)
	UpdateStatus(ctx context.Context, req *v1.UpdateStatusReq) (res *v1.UpdateStatusRes, err error)
	Trigger(ctx context.Context, req *v1.TriggerReq) (res *v1.TriggerRes, err error)
	Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error)
}
