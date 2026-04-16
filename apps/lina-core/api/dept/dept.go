// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package dept

import (
	"context"

	"lina-core/api/dept/v1"
)

type IDeptV1 interface {
	Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error)
	Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error)
	Exclude(ctx context.Context, req *v1.ExcludeReq) (res *v1.ExcludeRes, err error)
	Get(ctx context.Context, req *v1.GetReq) (res *v1.GetRes, err error)
	List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error)
	Tree(ctx context.Context, req *v1.TreeReq) (res *v1.TreeRes, err error)
	Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error)
	Users(ctx context.Context, req *v1.UsersReq) (res *v1.UsersRes, err error)
}
