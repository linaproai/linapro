// This file captures request audit payloads for the monitor-operlog source
// plugin by wrapping the host HTTP chain through the published global
// middleware registrar.

package backend

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/grpool"

	"lina-core/pkg/logger"
	hostaudit "lina-core/pkg/pluginservice/audit"
	hostbizctx "lina-core/pkg/pluginservice/bizctx"
)

// Shared host service adapters reused by the plugin-owned audit middleware.
var (
	// auditSvc dispatches request-audit events and reads dynamic-route metadata.
	auditSvc = hostaudit.New()
	// bizCtxSvc reads the authenticated operator identity from the current request context.
	bizCtxSvc = hostbizctx.New()
)

// maxParamLen bounds serialized request and response snippets captured by operation logs.
const maxParamLen = 2000

// Sensitive request-field masking tokens used by operation-log sanitization.
const (
	operLogMaskedPassword = "***"
	operLogRedactedValue  = "[REDACTED]"
	operLogBinaryContent  = "[二进制内容]"
)

// Local audit operation semantic constants mirror the published dictionary values.
const (
	operLogTypeCreate    = 1
	operLogTypeUpdate    = 2
	operLogTypeDelete    = 3
	operLogTypeExport    = 4
	operLogTypeImport    = 5
	operLogTypeOther     = 6
	operLogStatusSuccess = 0
	operLogStatusFail    = 1
)

// operLogTagToType maps semantic operLog tags to published audit operation codes.
var operLogTagToType = map[string]int{
	"create": operLogTypeCreate,
	"update": operLogTypeUpdate,
	"delete": operLogTypeDelete,
	"export": operLogTypeExport,
	"import": operLogTypeImport,
	"other":  operLogTypeOther,
}

// auditMiddleware captures one completed request and dispatches the normalized audit event.
func auditMiddleware(request *ghttp.Request) {
	startTime := time.Now()
	request.Middleware.Next()

	operName := bizCtxSvc.CurrentUsername(request.Context())
	if strings.TrimSpace(operName) == "" {
		return
	}

	var (
		handler     = request.GetServeHandler()
		operLogTag  = ""
		title       = ""
		operSummary = ""
	)
	if handler != nil {
		operLogTag = handler.GetMetaTag("operLog")
		title = handler.GetMetaTag("tags")
		operSummary = handler.GetMetaTag("summary")
	}
	if metadata := auditSvc.DynamicRouteMetadata(request); metadata != nil {
		if strings.TrimSpace(metadata.OperLogTag) != "" {
			operLogTag = metadata.OperLogTag
		}
		if strings.TrimSpace(metadata.Title) != "" {
			title = metadata.Title
		}
		if strings.TrimSpace(metadata.Summary) != "" {
			operSummary = metadata.Summary
		}
	}

	if !shouldRecordAuditRequest(request.Method, operLogTag) {
		return
	}

	var (
		operParam        = buildAuditRequestParam(request)
		jsonResult       = buildAuditResponseResult(request)
		status, errorMsg = resolveAuditStatus(request)
	)
	input := hostaudit.RecordInput{
		Title:         title,
		OperSummary:   operSummary,
		OperType:      inferOperType(request.Method, request.URL.Path, operLogTag),
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

	dispatchCtx := request.GetNeverDoneCtx()
	if dispatchCtx == nil {
		dispatchCtx = request.Context()
	}
	dispatchAuditEvent(dispatchCtx, input)
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
func dispatchAuditEvent(ctx context.Context, input hostaudit.RecordInput) {
	if err := grpool.AddWithRecover(ctx, func(taskCtx context.Context) {
		if recordErr := auditSvc.Record(taskCtx, input); recordErr != nil {
			logger.Warningf(taskCtx, "dispatch operation audit hook failed err=%v", recordErr)
		}
	}, func(taskCtx context.Context, panicErr error) {
		logger.Errorf(taskCtx, "monitor-operlog middleware panic: %v", panicErr)
	}); err != nil {
		logger.Warningf(ctx, "schedule operation audit task failed err=%v", err)
		if recordErr := auditSvc.Record(ctx, input); recordErr != nil {
			logger.Warningf(ctx, "fallback dispatch operation audit hook failed err=%v", recordErr)
		}
	}
}

// inferOperType determines the audit operation type from HTTP method, path, and operLog tag.
func inferOperType(method string, path string, operLogTag string) int {
	if strings.TrimSpace(operLogTag) != "" {
		if operType, ok := resolveOperLogTag(operLogTag); ok {
			return operType
		}
		return operLogTypeOther
	}

	switch method {
	case http.MethodPost:
		if strings.Contains(strings.ToLower(path), "import") {
			return operLogTypeImport
		}
		return operLogTypeCreate
	case http.MethodPut:
		return operLogTypeUpdate
	case http.MethodDelete:
		return operLogTypeDelete
	default:
		return operLogTypeOther
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
func buildAuditResponseResult(request *ghttp.Request) string {
	responseContentType := request.Response.Header().Get("Content-Type")
	responseBody := request.Response.BufferString()
	if metadata := auditSvc.DynamicRouteMetadata(request); metadata != nil {
		if responseContentType == "" {
			responseContentType = metadata.ResponseContentType
		}
		if responseBody == "" {
			responseBody = metadata.ResponseBody
		}
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
			return operLogStatusFail, err.Error()
		}
		return operLogStatusFail, ""
	}
	return operLogStatusSuccess, ""
}

// getRequestParam extracts request parameters as a JSON string.
func getRequestParam(request *ghttp.Request) string {
	body := request.GetBodyString()
	if body != "" {
		return body
	}
	params := request.GetQueryMap()
	if len(params) == 0 {
		return ""
	}
	buffer, err := json.Marshal(params)
	if err != nil {
		return ""
	}
	return string(buffer)
}

// sanitizeOperLogParam recursively masks sensitive request parameters before persistence.
func sanitizeOperLogParam(param string) string {
	if param == "" {
		return param
	}

	var data any
	if err := json.Unmarshal([]byte(param), &data); err != nil {
		return param
	}

	sanitized, changed := sanitizeOperLogValue(data)
	if !changed {
		return param
	}
	buffer, err := json.Marshal(sanitized)
	if err != nil {
		return param
	}
	return string(buffer)
}

// sanitizeOperLogValue traverses one decoded JSON value and masks password and environment payloads.
func sanitizeOperLogValue(value any) (any, bool) {
	switch typed := value.(type) {
	case map[string]any:
		sanitized := make(map[string]any, len(typed))
		changed := false
		for key, item := range typed {
			switch {
			case isOperLogPasswordField(key):
				sanitized[key] = operLogMaskedPassword
				changed = true
			case isOperLogEnvField(key):
				redacted, redactedChanged := redactOperLogEnvValue(item)
				sanitized[key] = redacted
				changed = changed || redactedChanged
			default:
				child, childChanged := sanitizeOperLogValue(item)
				sanitized[key] = child
				changed = changed || childChanged
			}
		}
		if !changed {
			return value, false
		}
		return sanitized, true

	case []any:
		sanitized := make([]any, len(typed))
		changed := false
		for index, item := range typed {
			child, childChanged := sanitizeOperLogValue(item)
			sanitized[index] = child
			changed = changed || childChanged
		}
		if !changed {
			return value, false
		}
		return sanitized, true
	}

	return value, false
}

// redactOperLogEnvValue masks one shell-environment payload while preserving visible keys when possible.
func redactOperLogEnvValue(value any) (any, bool) {
	switch typed := value.(type) {
	case map[string]any:
		if len(typed) == 0 {
			return value, false
		}
		sanitized := make(map[string]any, len(typed))
		for key := range typed {
			sanitized[key] = operLogRedactedValue
		}
		return sanitized, true

	case []any:
		if len(typed) == 0 {
			return value, false
		}
		sanitized := make([]any, len(typed))
		for index, item := range typed {
			sanitized[index] = redactOperLogEnvEntry(item)
		}
		return sanitized, true
	}

	return operLogRedactedValue, true
}

// redactOperLogEnvEntry masks one environment-variable entry inside an array payload.
func redactOperLogEnvEntry(value any) any {
	typed, ok := value.(map[string]any)
	if !ok {
		return operLogRedactedValue
	}

	sanitized := make(map[string]any, len(typed))
	for key, item := range typed {
		switch strings.ToLower(strings.TrimSpace(key)) {
		case "key", "name":
			sanitized[key] = item
		case "value":
			sanitized[key] = operLogRedactedValue
		default:
			sanitized[key] = operLogRedactedValue
		}
	}
	return sanitized
}

// isOperLogPasswordField reports whether the field name carries password semantics and must be masked.
func isOperLogPasswordField(field string) bool {
	switch strings.ToLower(strings.TrimSpace(field)) {
	case "password", "newpassword", "oldpassword":
		return true
	}
	return false
}

// isOperLogEnvField reports whether the field name carries shell environment variables and must be redacted.
func isOperLogEnvField(field string) bool {
	return strings.EqualFold(strings.TrimSpace(field), "env")
}

// isBinaryContentType reports whether the given content type represents binary data.
func isBinaryContentType(contentType string) bool {
	if contentType == "" {
		return false
	}
	lowerContentType := strings.ToLower(contentType)
	return strings.Contains(lowerContentType, "multipart/form-data") ||
		strings.Contains(lowerContentType, "application/octet-stream") ||
		strings.Contains(lowerContentType, "spreadsheetml") ||
		strings.Contains(lowerContentType, "image/") ||
		strings.Contains(lowerContentType, "audio/") ||
		strings.Contains(lowerContentType, "video/")
}

// truncate limits a string length and appends a suffix when truncation happens.
func truncate(value string, maxLen int) string {
	if len(value) <= maxLen {
		return value
	}
	return value[:maxLen] + "...(truncated)"
}
