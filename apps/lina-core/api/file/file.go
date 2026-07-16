// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package file

import (
	"context"

	"lina-core/api/file/v1"
)

type IFileV1 interface {
	Access(ctx context.Context, req *v1.AccessReq) (res *v1.AccessRes, err error)
	Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error)
	Detail(ctx context.Context, req *v1.DetailReq) (res *v1.DetailRes, err error)
	DirectUploadInit(ctx context.Context, req *v1.DirectUploadInitReq) (res *v1.DirectUploadInitRes, err error)
	DirectUploadComplete(ctx context.Context, req *v1.DirectUploadCompleteReq) (res *v1.DirectUploadCompleteRes, err error)
	DirectUploadAbort(ctx context.Context, req *v1.DirectUploadAbortReq) (res *v1.DirectUploadAbortRes, err error)
	DirectUploadPartURL(ctx context.Context, req *v1.DirectUploadPartURLReq) (res *v1.DirectUploadPartURLRes, err error)
	DirectDownload(ctx context.Context, req *v1.DirectDownloadReq) (res *v1.DirectDownloadRes, err error)
	Download(ctx context.Context, req *v1.DownloadReq) (res *v1.DownloadRes, err error)
	InfoByIds(ctx context.Context, req *v1.InfoByIdsReq) (res *v1.InfoByIdsRes, err error)
	List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error)
	FileSuffixes(ctx context.Context, req *v1.FileSuffixesReq) (res *v1.FileSuffixesRes, err error)
	Upload(ctx context.Context, req *v1.UploadReq) (res *v1.UploadRes, err error)
	ChunkedUploadInit(ctx context.Context, req *v1.ChunkedUploadInitReq) (res *v1.ChunkedUploadInitRes, err error)
	ChunkedUploadPart(ctx context.Context, req *v1.ChunkedUploadPartReq) (res *v1.ChunkedUploadPartRes, err error)
	ChunkedUploadComplete(ctx context.Context, req *v1.ChunkedUploadCompleteReq) (res *v1.ChunkedUploadCompleteRes, err error)
	ChunkedUploadAbort(ctx context.Context, req *v1.ChunkedUploadAbortReq) (res *v1.ChunkedUploadAbortRes, err error)
	UsageScenes(ctx context.Context, req *v1.UsageScenesReq) (res *v1.UsageScenesRes, err error)
}
