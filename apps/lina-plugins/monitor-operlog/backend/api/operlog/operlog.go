// Package operlog declares the HTTP controller contract exposed by the
// monitor-operlog source plugin.
package operlog

import (
	"context"

	"lina-plugin-monitor-operlog/backend/api/operlog/v1"
)

// IOperlogV1 defines the monitor-operlog HTTP handlers.
type IOperlogV1 interface {
	// Clean deletes operation logs within one optional time range.
	Clean(ctx context.Context, req *v1.CleanReq) (res *v1.CleanRes, err error)
	// Delete deletes one or more operation logs by ID list.
	Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error)
	// Export exports operation logs into an Excel workbook.
	Export(ctx context.Context, req *v1.ExportReq) (res *v1.ExportRes, err error)
	// Get returns one operation-log detail by ID.
	Get(ctx context.Context, req *v1.GetReq) (res *v1.GetRes, err error)
	// List returns one paginated operation-log list.
	List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error)
}
