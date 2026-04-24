// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package i18n

import (
	"context"

	"lina-core/api/i18n/v1"
)

type II18NV1 interface {
	DiagnoseMessages(ctx context.Context, req *v1.DiagnoseMessagesReq) (res *v1.DiagnoseMessagesRes, err error)
	ExportMessages(ctx context.Context, req *v1.ExportMessagesReq) (res *v1.ExportMessagesRes, err error)
	ImportMessages(ctx context.Context, req *v1.ImportMessagesReq) (res *v1.ImportMessagesRes, err error)
	MissingMessages(ctx context.Context, req *v1.MissingMessagesReq) (res *v1.MissingMessagesRes, err error)
	RuntimeLocales(ctx context.Context, req *v1.RuntimeLocalesReq) (res *v1.RuntimeLocalesRes, err error)
	RuntimeMessages(ctx context.Context, req *v1.RuntimeMessagesReq) (res *v1.RuntimeMessagesRes, err error)
}
