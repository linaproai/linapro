// This file records operation logs for host and plugin-backed HTTP requests.

package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/grpool"

	"lina-core/internal/service/operlog"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/pkg/logger"
)

// maxParamLen bounds serialized request and response snippets captured by
// operation logs.
const maxParamLen = 2000

// Sensitive request-field masking tokens used by operation-log sanitization.
const (
	operLogMaskedPassword = "***"
	operLogRedactedValue  = "[REDACTED]"
)

// OperLog records operation logs for write operations and specially tagged GET operations.
func (s *serviceImpl) OperLog(r *ghttp.Request) {
	startTime := time.Now()
	r.Middleware.Next()

	// Collect all data synchronously (r.Response buffer is only available now)
	var (
		method      = r.Method
		handler     = r.GetServeHandler()
		operLogTag  = ""
		title       = ""
		operSummary = ""
		dynamicMeta = pluginsvc.GetDynamicRouteOperLogMetadata(r)
	)
	if handler != nil {
		operLogTag = handler.GetMetaTag("operLog")
		title = handler.GetMetaTag("tags")
		operSummary = handler.GetMetaTag("summary")
	}
	if dynamicMeta != nil {
		if strings.TrimSpace(dynamicMeta.OperLogTag) != "" {
			operLogTag = dynamicMeta.OperLogTag
		}
		if strings.TrimSpace(dynamicMeta.Title) != "" {
			title = dynamicMeta.Title
		}
		if strings.TrimSpace(dynamicMeta.Summary) != "" {
			operSummary = dynamicMeta.Summary
		}
	}

	// Only log write operations (POST/PUT/DELETE) or GET with operLog tag
	shouldLog := false
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodDelete:
		shouldLog = true
	case http.MethodGet:
		shouldLog = operLogTag != ""
	}
	if !shouldLog {
		return
	}

	operType := inferOperType(method, r.URL.Path, operLogTag)
	operName := ""
	if bizCtx := s.bizCtxSvc.Get(r.Context()); bizCtx != nil {
		operName = bizCtx.Username
	}

	// Get request parameters (skip binary content like file uploads)
	operParam := ""
	reqContentType := r.GetHeader("Content-Type")
	if isBinaryContentType(reqContentType) {
		operParam = "[二进制内容]"
	} else {
		operParam = truncate(sanitizeOperLogParam(getRequestParam(r)), maxParamLen)
	}

	// Get response result (skip binary content like xlsx exports)
	jsonResult := ""
	resContentType := r.Response.Header().Get("Content-Type")
	if resContentType == "" && dynamicMeta != nil {
		resContentType = dynamicMeta.ResponseContentType
	}
	if isBinaryContentType(resContentType) {
		jsonResult = "[二进制内容]"
	} else {
		jsonResult = truncate(r.Response.BufferString(), maxParamLen)
		if jsonResult == "" && dynamicMeta != nil {
			jsonResult = truncate(dynamicMeta.ResponseBody, maxParamLen)
		}
	}

	status := operlog.OperStatusSuccess
	errorMsg := ""
	if r.Response.Status >= 400 || r.GetError() != nil {
		status = operlog.OperStatusFail
		if r.GetError() != nil {
			errorMsg = r.GetError().Error()
		}
	}

	var (
		costTime  = int(time.Since(startTime).Milliseconds())
		urlPath   = r.URL.Path
		urlString = r.URL.String()
		clientIp  = r.GetClientIp()
		input     = operlog.CreateInput{
			Title:         title,
			OperSummary:   operSummary,
			OperType:      operType,
			Method:        urlPath,
			RequestMethod: method,
			OperName:      operName,
			OperUrl:       urlString,
			OperIp:        clientIp,
			OperParam:     operParam,
			JsonResult:    jsonResult,
			Status:        status,
			ErrorMsg:      errorMsg,
			CostTime:      costTime,
		}
	)

	// Async write using grpool (goroutine pool) with NeverDoneCtx
	ctx := r.GetNeverDoneCtx()
	if err := grpool.AddWithRecover(ctx, func(ctx context.Context) {
		if createErr := s.operLogSvc.Create(ctx, input); createErr != nil {
			logger.Warningf(ctx, "create operation log failed err=%v", createErr)
		}
	}, func(ctx context.Context, err error) {
		logger.Errorf(ctx, "operlog middleware panic: %v", err)
	}); err != nil {
		logger.Warningf(ctx, "schedule operation log task failed err=%v", err)
		if createErr := s.operLogSvc.Create(ctx, input); createErr != nil {
			logger.Warningf(ctx, "fallback create operation log failed err=%v", createErr)
		}
	}
}

// inferOperType determines operation type from HTTP method and path.
func inferOperType(method, path, operLogTag string) int {
	if operLogTag != "" {
		if operType, ok := operlog.ResolveOperTag(operLogTag); ok {
			return operType
		}
		return operlog.OperTypeOther
	}

	switch method {
	case http.MethodPost:
		if strings.Contains(strings.ToLower(path), "import") {
			return operlog.OperTypeImport
		}
		return operlog.OperTypeCreate
	case http.MethodPut:
		return operlog.OperTypeUpdate
	case http.MethodDelete:
		return operlog.OperTypeDelete
	default:
		return operlog.OperTypeOther
	}
}

// getRequestParam extracts request parameters as JSON string.
func getRequestParam(r *ghttp.Request) string {
	body := r.GetBodyString()
	if body != "" {
		return body
	}
	params := r.GetQueryMap()
	if len(params) > 0 {
		b, err := json.Marshal(params)
		if err != nil {
			return ""
		}
		return string(b)
	}
	return ""
}

// sanitizeOperLogParam recursively masks sensitive request parameters before
// operation-log persistence.
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
	b, err := json.Marshal(sanitized)
	if err != nil {
		return param
	}
	return string(b)
}

// sanitizeOperLogValue traverses one decoded JSON value and masks password and
// shell-environment payloads.
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

// redactOperLogEnvValue masks one shell-environment payload while preserving
// environment variable keys when the payload shape exposes them.
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

// redactOperLogEnvEntry masks one environment-variable entry inside an array
// payload while keeping the variable key or name visible when possible.
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

// isOperLogPasswordField reports whether the field name carries password
// semantics and must be masked.
func isOperLogPasswordField(field string) bool {
	switch strings.ToLower(strings.TrimSpace(field)) {
	case "password", "newpassword", "oldpassword":
		return true
	}
	return false
}

// isOperLogEnvField reports whether the field name carries shell environment
// variables and must be redacted.
func isOperLogEnvField(field string) bool {
	return strings.EqualFold(strings.TrimSpace(field), "env")
}

// isBinaryContentType checks if the content type represents binary data.
func isBinaryContentType(contentType string) bool {
	if contentType == "" {
		return false
	}
	ct := strings.ToLower(contentType)
	return strings.Contains(ct, "multipart/form-data") ||
		strings.Contains(ct, "application/octet-stream") ||
		strings.Contains(ct, "spreadsheetml") ||
		strings.Contains(ct, "image/") ||
		strings.Contains(ct, "audio/") ||
		strings.Contains(ct, "video/")
}

// truncate truncates a string to maxLen and appends suffix if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "...(truncated)"
}
