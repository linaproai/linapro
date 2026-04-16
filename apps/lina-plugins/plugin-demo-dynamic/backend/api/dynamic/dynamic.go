package dynamicapi

import (
	"context"

	"lina-plugin-demo-dynamic/backend/api/dynamic/v1"
)

// IDynamicV1 defines the backend API contract for the dynamic sample plugin.
type IDynamicV1 interface {
	BackendSummary(ctx context.Context, req *v1.BackendSummaryReq) (res *v1.BackendSummaryRes, err error)
	HostCallDemo(ctx context.Context, req *v1.HostCallDemoReq) (res *v1.HostCallDemoRes, err error)
}
