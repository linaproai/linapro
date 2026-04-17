// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package publicconfig

import (
	"context"

	"lina-core/api/publicconfig/v1"
)

// IConfigPublicV1 defines the public config API contract.
type IConfigPublicV1 interface {
	Frontend(ctx context.Context, req *v1.FrontendReq) (res *v1.FrontendRes, err error)
}
