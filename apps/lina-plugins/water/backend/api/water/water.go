// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package water

import (
	"context"

	"lina-plugin-water/backend/api/water/v1"
)

type IWaterV1 interface {
	Preview(ctx context.Context, req *v1.PreviewReq) (res *v1.PreviewRes, err error)
	SubmitSnap(ctx context.Context, req *v1.SubmitSnapReq) (res *v1.SubmitSnapRes, err error)
	GetTask(ctx context.Context, req *v1.GetTaskReq) (res *v1.GetTaskRes, err error)
}
