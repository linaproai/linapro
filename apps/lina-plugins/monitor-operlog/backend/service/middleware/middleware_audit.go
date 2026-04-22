// This file implements the plugin-owned operation-audit middleware and event
// dispatch flow.

package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/grpool"

	"lina-core/pkg/logger"
	hostaudit "lina-core/pkg/pluginservice/audit"
	operlogsvc "lina-plugin-monitor-operlog/backend/service/operlog"
)

// maxParamLen bounds serialized request and response snippets captured by operation logs.
const maxParamLen = 2000

// Sensitive request-field masking tokens used by operation-log sanitization.
const (
	operLogMaskedPassword = "***"
	operLogRedactedValue  = "[REDACTED]"
	operLogBinaryContent  = "[二进制内容]"
)

// operLogTagToType maps semantic operLog tags to the published audit operation codes.
var operLogTagToType = map[string]int{
	"create": operlogsvc.OperTypeCreate,
	"update": operlogsvc.OperTypeUpdate,
	"delete": operlogsvc.OperTypeDelete,
	"export": operlogsvc.OperTypeExport,
	"import": operlogsvc.OperTypeImport,
	"other":  operlogsvc.OperTypeOther,
}

// auditRouteMetadata stores the route-level audit metadata collected from the
// static handler declaration and the dynamic-route runtime projection.
type auditRouteMetadata struct {
	operLogTag          string
	title               string
	operSummary         string
	responseBody        string
	responseContentType string
}

// Audit captures one completed request and dispatches the normalized audit event.
func (s *serviceImpl) Audit(request *ghttp.Request) {
	if request == nil {
		return
	}

	startTime := time.Now()
	request.Middleware.Next()

	operName := s.currentUsername(request.Context())
	if strings.TrimSpace(operName) == "" {
		return
	}

	metadata := s.resolveAuditRouteMetadata(request)
	if !shouldRecordAuditRequest(request.Method, metadata.operLogTag) {
		return
	}

	input := buildAuditRecordInput(request, metadata, operName, startTime)
	dispatchCtx := request.GetNeverDoneCtx()
	if dispatchCtx == nil {
		dispatchCtx = request.Context()
	}
	s.dispatchAuditEvent(dispatchCtx, input)
}

// currentUsername reads the authenticated operator username from the request context.
func (s *serviceImpl) currentUsername(ctx context.Context) string {
	if s == nil || s.bizCtxSvc == nil {
		return ""
	}
	return s.bizCtxSvc.CurrentUsername(ctx)
}

// resolveAuditRouteMetadata loads audit tags from the static handler metadata
// and lets the dynamic-route projection override them when available.
func (s *serviceImpl) resolveAuditRouteMetadata(request *ghttp.Request) auditRouteMetadata {
	metadata := auditRouteMetadata{}
	if request == nil {
		return metadata
	}

	handler := request.GetServeHandler()
	if handler != nil {
		metadata.operLogTag = handler.GetMetaTag("operLog")
		metadata.title = handler.GetMetaTag("tags")
		metadata.operSummary = handler.GetMetaTag("summary")
	}

	if s == nil || s.auditSvc == nil {
		return metadata
	}
	dynamicMetadata := s.auditSvc.DynamicRouteMetadata(request)
	if dynamicMetadata == nil {
		return metadata
	}
	if strings.TrimSpace(dynamicMetadata.OperLogTag) != "" {
		metadata.operLogTag = dynamicMetadata.OperLogTag
	}
	if strings.TrimSpace(dynamicMetadata.Title) != "" {
		metadata.title = dynamicMetadata.Title
	}
	if strings.TrimSpace(dynamicMetadata.Summary) != "" {
		metadata.operSummary = dynamicMetadata.Summary
	}
	if dynamicMetadata.ResponseBody != "" {
		metadata.responseBody = dynamicMetadata.ResponseBody
	}
	if dynamicMetadata.ResponseContentType != "" {
		metadata.responseContentType = dynamicMetadata.ResponseContentType
	}
	return metadata
}

// buildAuditRecordInput normalizes one completed request into the published audit payload.
func buildAuditRecordInput(
	request *ghttp.Request,
	metadata auditRouteMetadata,
	operName string,
	startTime time.Time,
) hostaudit.RecordInput {
	var (
		operParam        = buildAuditRequestParam(request)
		jsonResult       = buildAuditResponseResult(request, metadata)
		status, errorMsg = resolveAuditStatus(request)
	)
	return hostaudit.RecordInput{
		Title:         metadata.title,
		OperSummary:   metadata.operSummary,
		OperType:      inferOperType(request.Method, request.URL.Path, metadata.operLogTag),
		Method:        request.URL.Path,
		RequestMethod: request.Method,
		OperName:      operName,
		OperURL:       request.URL.String(),
		OperIP:        request.GetClientIp(),
		OperParam:     operParam,
		JSONResult:    jsonResult,
		Status:        status,
		ErrorMsg:      errorMsg,
		CostTime:      int(time.Since(startTime).Milliseconds()),
	}
}

// shouldRecordAuditRequest reports whether the current request matches audit logging rules.
func shouldRecordAuditRequest(method string, operLogTag string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodDelete:
		return true
	case http.MethodGet:
		return strings.TrimSpace(operLogTag) != ""
	default:
		return false
	}
}

// dispatchAuditEvent schedules the audit hook dispatch without blocking the request path.
func (s *serviceImpl) dispatchAuditEvent(ctx context.Context, input hostaudit.RecordInput) {
	if err := grpool.AddWithRecover(ctx, func(taskCtx context.Context) {
		if recordErr := s.recordAuditEvent(taskCtx, input); recordErr != nil {
			logger.Warningf(taskCtx, "dispatch operation audit hook failed err=%v", recordErr)
		}
	}, func(taskCtx context.Context, panicErr error) {
		logger.Errorf(taskCtx, "monitor-operlog middleware panic: %v", panicErr)
	}); err != nil {
		logger.Warningf(ctx, "schedule operation audit task failed err=%v", err)
		if recordErr := s.recordAuditEvent(ctx, input); recordErr != nil {
			logger.Warningf(ctx, "fallback dispatch operation audit hook failed err=%v", recordErr)
		}
	}
}

// recordAuditEvent publishes one normalized audit event through the host audit service.
func (s *serviceImpl) recordAuditEvent(ctx context.Context, input hostaudit.RecordInput) error {
	if s == nil || s.auditSvc == nil {
		return nil
	}
	return s.auditSvc.Record(ctx, input)
}

// inferOperType determines the audit operation type from HTTP method, path, and operLog tag.
func inferOperType(method string, path string, operLogTag string) int {
	if strings.TrimSpace(operLogTag) != "" {
		if operType, ok := resolveOperLogTag(operLogTag); ok {
			return operType
		}
		return operlogsvc.OperTypeOther
	}

	switch method {
	case http.MethodPost:
		if strings.Contains(strings.ToLower(path), "import") {
			return operlogsvc.OperTypeImport
		}
		return operlogsvc.OperTypeCreate
	case http.MethodPut:
		return operlogsvc.OperTypeUpdate
	case http.MethodDelete:
		return operlogsvc.OperTypeDelete
	default:
		return operlogsvc.OperTypeOther
	}
}

// resolveOperLogTag converts a semantic operLog tag to the published audit operation type code.
func resolveOperLogTag(tag string) (int, bool) {
	operType, ok := operLogTagToType[strings.TrimSpace(tag)]
	return operType, ok
}

// buildAuditRequestParam extracts the request payload snippet suitable for operation logging.
func buildAuditRequestParam(request *ghttp.Request) string {
	if isBinaryContentType(request.GetHeader("Content-Type")) {
		return operLogBinaryContent
	}
	return truncate(sanitizeOperLogParam(getRequestParam(request)), maxParamLen)
}

// buildAuditResponseResult extracts the response snippet suitable for operation logging.
func buildAuditResponseResult(request *ghttp.Request, metadata auditRouteMetadata) string {
	responseContentType := request.Response.Header().Get("Content-Type")
	responseBody := request.Response.BufferString()
	if responseContentType == "" {
		responseContentType = metadata.responseContentType
	}
	if responseBody == "" {
		responseBody = metadata.responseBody
	}
	if isBinaryContentType(responseContentType) {
		return operLogBinaryContent
	}
	return truncate(responseBody, maxParamLen)
}

// resolveAuditStatus derives the normalized audit status and error message from the request result.
func resolveAuditStatus(request *ghttp.Request) (int, string) {
	if request.Response.Status >= http.StatusBadRequest || request.GetError() != nil {
		if err := request.GetError(); err != nil {
			return operlogsvc.OperStatusFail, err.Error()
		}
		return operlogsvc.OperStatusFail, ""
	}
	return operlogsvc.OperStatusSuccess, ""
}
