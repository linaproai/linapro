// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package config

import (
	"context"

	"lina-core/api/config/v1"
)

type IConfigV1 interface {
	ByKey(ctx context.Context, req *v1.ByKeyReq) (res *v1.ByKeyRes, err error)
	Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error)
	Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error)
	Export(ctx context.Context, req *v1.ExportReq) (res *v1.ExportRes, err error)
	Get(ctx context.Context, req *v1.GetReq) (res *v1.GetRes, err error)
	ConfigImport(ctx context.Context, req *v1.ConfigImportReq) (res *v1.ConfigImportRes, err error)
	ConfigImportTemplate(ctx context.Context, req *v1.ConfigImportTemplateReq) (res *v1.ConfigImportTemplateRes, err error)
	List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error)
	Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error)
}
