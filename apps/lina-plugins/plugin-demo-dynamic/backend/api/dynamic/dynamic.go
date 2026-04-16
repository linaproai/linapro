package dynamicapi

import (
	"context"

	"lina-plugin-demo-dynamic/backend/api/dynamic/v1"
)

// IDynamicV1 defines the backend API contract for the dynamic sample plugin.
type IDynamicV1 interface {
	BackendSummary(ctx context.Context, req *v1.BackendSummaryReq) (res *v1.BackendSummaryRes, err error)
	DemoRecordList(ctx context.Context, req *v1.DemoRecordListReq) (res *v1.DemoRecordListRes, err error)
	DemoRecord(ctx context.Context, req *v1.DemoRecordReq) (res *v1.DemoRecordRes, err error)
	CreateDemoRecord(ctx context.Context, req *v1.CreateDemoRecordReq) (res *v1.CreateDemoRecordRes, err error)
	UpdateDemoRecord(ctx context.Context, req *v1.UpdateDemoRecordReq) (res *v1.UpdateDemoRecordRes, err error)
	DeleteDemoRecord(ctx context.Context, req *v1.DeleteDemoRecordReq) (res *v1.DeleteDemoRecordRes, err error)
	DownloadDemoRecordAttachment(ctx context.Context, req *v1.DownloadDemoRecordAttachmentReq) (res *v1.DownloadDemoRecordAttachmentRes, err error)
	HostCallDemo(ctx context.Context, req *v1.HostCallDemoReq) (res *v1.HostCallDemoRes, err error)
}
