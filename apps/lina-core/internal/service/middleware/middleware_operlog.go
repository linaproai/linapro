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

const maxParamLen = 2000 // Max length for parameters and results

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
		operParam = truncate(maskPassword(getRequestParam(r)), maxParamLen)
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

// maskPassword replaces password field values with ***.
func maskPassword(param string) string {
	if param == "" {
		return param
	}
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(param), &data); err != nil {
		return param
	}
	masked := false
	for k := range data {
		lower := strings.ToLower(k)
		if lower == "password" || lower == "newpassword" || lower == "oldpassword" {
			data[k] = "***"
			masked = true
		}
	}
	if !masked {
		return param
	}
	b, err := json.Marshal(data)
	if err != nil {
		return param
	}
	return string(b)
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
