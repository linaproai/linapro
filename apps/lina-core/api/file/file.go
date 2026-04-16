// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package file

import (
	"context"

	"lina-core/api/file/v1"
)

type IFileV1 interface {
	Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error)
	Detail(ctx context.Context, req *v1.DetailReq) (res *v1.DetailRes, err error)
	Download(ctx context.Context, req *v1.DownloadReq) (res *v1.DownloadRes, err error)
	InfoByIds(ctx context.Context, req *v1.InfoByIdsReq) (res *v1.InfoByIdsRes, err error)
	List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error)
	FileSuffixes(ctx context.Context, req *v1.FileSuffixesReq) (res *v1.FileSuffixesRes, err error)
	Upload(ctx context.Context, req *v1.UploadReq) (res *v1.UploadRes, err error)
	UsageScenes(ctx context.Context, req *v1.UsageScenesReq) (res *v1.UsageScenesRes, err error)
}
