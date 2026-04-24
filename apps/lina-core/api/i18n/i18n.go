// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package i18n

import (
	"context"

	"lina-core/api/i18n/v1"
)

type II18NV1 interface {
	RuntimeLocales(ctx context.Context, req *v1.RuntimeLocalesReq) (res *v1.RuntimeLocalesRes, err error)
	RuntimeMessages(ctx context.Context, req *v1.RuntimeMessagesReq) (res *v1.RuntimeMessagesRes, err error)
}
