// Package loginlog declares the HTTP controller contract exposed by the
// monitor-loginlog source plugin.
package loginlog

import (
	"context"

	"lina-plugin-monitor-loginlog/backend/api/loginlog/v1"
)

// ILoginlogV1 defines the monitor-loginlog HTTP handlers.
type ILoginlogV1 interface {
	// Clean deletes login logs within one optional time range.
	Clean(ctx context.Context, req *v1.CleanReq) (res *v1.CleanRes, err error)
	// Delete deletes one or more login logs by ID list.
	Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error)
	// Export exports login logs into an Excel workbook.
	Export(ctx context.Context, req *v1.ExportReq) (res *v1.ExportRes, err error)
	// Get returns one login-log detail by ID.
	Get(ctx context.Context, req *v1.GetReq) (res *v1.GetRes, err error)
	// List returns one paginated login-log list.
	List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error)
}
