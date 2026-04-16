// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package usermsg

import (
	"context"

	"lina-core/api/usermsg/v1"
)

type IUsermsgV1 interface {
	Clear(ctx context.Context, req *v1.ClearReq) (res *v1.ClearRes, err error)
	Count(ctx context.Context, req *v1.CountReq) (res *v1.CountRes, err error)
	Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error)
	List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error)
	Read(ctx context.Context, req *v1.ReadReq) (res *v1.ReadRes, err error)
	ReadAll(ctx context.Context, req *v1.ReadAllReq) (res *v1.ReadAllRes, err error)
}
