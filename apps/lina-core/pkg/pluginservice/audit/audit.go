// Package audit exposes a narrowed host audit-recording contract to source
// plugins so they can emit request-audit events and read dynamic-route metadata
// without depending on host-internal service packages.
package audit

import (
	"context"

	"github.com/gogf/gf/v2/net/ghttp"

	internalplugin "lina-core/internal/service/plugin"
	"lina-core/pkg/audittype"
)

// RecordInput defines the request-audit payload published to source plugins.
type RecordInput struct {
	// Title is the audit title derived from handler metadata.
	Title string
	// OperSummary is the audit summary derived from handler metadata.
	OperSummary string
	// OperType is the normalized semantic audit operation type.
	OperType audittype.OperType
	// Method is the routed handler path or method marker.
	Method string
	// RequestMethod is the HTTP request method.
	RequestMethod string
	// OperName is the operator username.
	OperName string
	// OperURL is the full request URL.
	OperURL string
	// OperIP is the client IP captured by the audit event.
	OperIP string
	// OperParam is the sanitized request payload snippet.
	OperParam string
	// JSONResult is the serialized response snippet.
	JSONResult string
	// Status is the audit status code.
	Status int
	// ErrorMsg is the captured error summary.
	ErrorMsg string
	// CostTime is the request duration in milliseconds.
	CostTime int
}

// DynamicRouteMetadata is the published projection of dynamic-route audit metadata.
type DynamicRouteMetadata struct {
	// Title is the route-tag projection used as the audit title.
	Title string
	// Summary is the route summary projected into the audit record.
	Summary string
	// OperLogTag is the normalized operLog tag attached to the matched route.
	OperLogTag string
	// ResponseBody stores the raw bridge response body captured by the runtime dispatcher.
	ResponseBody string
	// ResponseContentType stores the resolved content type of the bridge response.
	ResponseContentType string
}

// Service defines the audit operations published to source plugins.
type Service interface {
	// Record dispatches one request-audit event through the host plugin runtime.
	Record(ctx context.Context, input RecordInput) error
	// DynamicRouteMetadata returns the dynamic-route audit metadata attached to the current request.
	DynamicRouteMetadata(request *ghttp.Request) *DynamicRouteMetadata
}

// serviceAdapter bridges the internal plugin service into the published audit contract.
type serviceAdapter struct {
	service internalplugin.Service
}

// New creates and returns the published audit service adapter.
func New() Service {
	return &serviceAdapter{service: internalplugin.New(nil)}
}

// Record dispatches one request-audit event through the host plugin runtime.
func (s *serviceAdapter) Record(ctx context.Context, input RecordInput) error {
	if s == nil || s.service == nil {
		return nil
	}
	return s.service.HandleAuditRecorded(ctx, internalplugin.AuditRecordedInput{
		Title:         input.Title,
		OperSummary:   input.OperSummary,
		OperType:      input.OperType,
		Method:        input.Method,
		RequestMethod: input.RequestMethod,
		OperName:      input.OperName,
		OperUrl:       input.OperURL,
		OperIp:        input.OperIP,
		OperParam:     input.OperParam,
		JsonResult:    input.JSONResult,
		Status:        input.Status,
		ErrorMsg:      input.ErrorMsg,
		CostTime:      input.CostTime,
	})
}

// DynamicRouteMetadata returns the dynamic-route audit metadata attached to the current request.
func (s *serviceAdapter) DynamicRouteMetadata(request *ghttp.Request) *DynamicRouteMetadata {
	metadata := internalplugin.GetDynamicRouteOperLogMetadata(request)
	if metadata == nil {
		return nil
	}
	return &DynamicRouteMetadata{
		Title:               metadata.Title,
		Summary:             metadata.Summary,
		OperLogTag:          metadata.OperLogTag,
		ResponseBody:        metadata.ResponseBody,
		ResponseContentType: metadata.ResponseContentType,
	}
}
