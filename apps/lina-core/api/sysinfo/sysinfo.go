// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package sysinfo

import (
	"context"

	"lina-core/api/sysinfo/v1"
)

type ISysinfoV1 interface {
	GetInfo(ctx context.Context, req *v1.GetInfoReq) (res *v1.GetInfoRes, err error)
}
