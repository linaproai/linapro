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
	"lina-core/pkg/pluginhost"
	hostapidoc "lina-core/pkg/pluginservice/contract"
	"lina-plugin-monitor-operlog/backend/internal/model/operlogtype"
	operlogsvc "lina-plugin-monitor-operlog/backend/internal/service/operlog"
)

// maxParamLen bounds serialized request and response snippets captured by operation logs.
const maxParamLen = 2000

// GoFrame route metadata tag names consumed by the audit middleware.
const (
	routeMetaOperLog = "operLog"
	routeMetaPath    = "path"
	routeMetaMethod  = "method"
	routeMetaTags    = "tags"
	routeMetaSummary = "summary"
)

// Sensitive request-field masking tokens used by operation-log sanitization.
const (
	operLogMaskedPassword = "***"
	operLogRedactedValue  = "[REDACTED]"
	operLogBinaryContent  = "[BINARY CONTENT]"
)

// auditRouteMetadata stores the route-level audit metadata collected from the
// static handler declaration and the dynamic-route runtime projection.
type auditRouteMetadata struct {
	operLogTag          string
	title               string
	operSummary         string
	routeOwner          string
	routeMethod         string
	routePath           string
	routeDocKey         string
	responseBody        string
	responseContentType string
}

// Audit captures one completed request and writes the normalized operation log.
func (s *serviceImpl) Audit(request *ghttp.Request) {
	if request == nil {
		return
	}

	// Capture the start time before handing control to the remaining middleware
	// chain so costTime covers auth, controller execution, dynamic plugin dispatch,
	// and response serialization.
	startTime := time.Now()
	request.Middleware.Next()

	// The authenticated operator comes from the host biz context populated by
	// upstream auth middleware. Public or unauthenticated requests are skipped
	// because an operation log entry must be attributable to an operator.
	operName := s.currentUsername(request.Context())
	if strings.TrimSpace(operName) == "" {
		return
	}

	// Route metadata is resolved after the handler has run because dynamic route
	// dispatch attaches its matched plugin route metadata to the request context
	// during execution.
	metadata := s.resolveAuditRouteMetadata(request)
	if !shouldRecordAuditRequest(request.Method, metadata.operLogTag) {
		return
	}

	input := buildOperLogCreateInput(request, metadata, operName, startTime)
	dispatchCtx := request.GetNeverDoneCtx()
	if dispatchCtx == nil {
		dispatchCtx = request.Context()
	}
	// Persist with the request's never-done context when available so the
	// asynchronous audit task can finish even after GoFrame completes the HTTP
	// response lifecycle.
	s.persistOperLogRecord(dispatchCtx, input)
}

// currentUsername reads the authenticated operator username from the request context.
func (s *serviceImpl) currentUsername(ctx context.Context) string {
	if s == nil || s.bizCtxSvc == nil {
		return ""
	}
	return s.bizCtxSvc.Current(ctx).Username
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
		// Static core/source-plugin routes expose their g.Meta values through the
		// parsed GoFrame handler. This is the first metadata source for the audit
		// title, operation summary, explicit operLog tag, and stable route anchor.
		metadata.operLogTag = handler.GetMetaTag(routeMetaOperLog)
		metadata.title = handler.GetMetaTag(routeMetaTags)
		metadata.operSummary = handler.GetMetaTag(routeMetaSummary)
		metadata.routeOwner, metadata.routeMethod, metadata.routePath, metadata.routeDocKey = buildStaticRouteAnchor(request, handler)
	}

	if s == nil || s.routeMetaSvc == nil {
		return metadata
	}
	dynamicMetadata := s.routeMetaSvc.DynamicRouteMetadata(request)
	if dynamicMetadata == nil {
		return metadata
	}
	// Dynamic plugin routes enter through a fixed host dispatcher, so the GoFrame
	// handler metadata describes the dispatcher instead of the matched plugin
	// route. The host runtime stores the matched manifest route on the request
	// context, and the published route service exposes that projection to this source plugin.
	// Non-empty dynamic fields therefore override the static dispatcher metadata.
	if strings.TrimSpace(dynamicMetadata.Meta[routeMetaOperLog]) != "" {
		metadata.operLogTag = dynamicMetadata.Meta[routeMetaOperLog]
	}
	if len(dynamicMetadata.Tags) > 0 {
		metadata.title = strings.Join(dynamicMetadata.Tags, ",")
	}
	if strings.TrimSpace(dynamicMetadata.Summary) != "" {
		metadata.operSummary = dynamicMetadata.Summary
	}
	if strings.TrimSpace(dynamicMetadata.PluginID) != "" {
		metadata.routeOwner = dynamicMetadata.PluginID
	}
	if strings.TrimSpace(dynamicMetadata.Method) != "" {
		metadata.routeMethod = dynamicMetadata.Method
	}
	if strings.TrimSpace(dynamicMetadata.PublicPath) != "" {
		metadata.routePath = dynamicMetadata.PublicPath
	}
	if strings.TrimSpace(metadata.routePath) != "" {
		// The route doc key is rebuilt from the final public path/method so later
		// display logic can reuse apidoc i18n resources regardless of whether the
		// request came from a static route or a dynamic plugin route.
		metadata.routeDocKey = hostapidoc.BuildOperationKeyFromPath(metadata.routePath, metadata.routeMethod)
	}
	if dynamicMetadata.ResponseBody != "" {
		// Dynamic route responses are written by the bridge dispatcher as raw
		// payloads; keep them as a fallback for cases where GoFrame's response
		// buffer no longer exposes the body to downstream middleware.
		metadata.responseBody = dynamicMetadata.ResponseBody
	}
	if dynamicMetadata.ResponseContentType != "" {
		metadata.responseContentType = dynamicMetadata.ResponseContentType
	}
	return metadata
}

// buildOperLogCreateInput normalizes one completed request into the plugin-owned operation-log create input.
func buildOperLogCreateInput(
	request *ghttp.Request,
	metadata auditRouteMetadata,
	operName string,
	startTime time.Time,
) operlogsvc.CreateInput {
	var (
		// Request and response snippets are sanitized/truncated before persistence
		// so operation logs remain useful without storing credentials or large
		// payloads.
		operParam        = buildAuditRequestParam(request)
		jsonResult       = buildAuditResponseResult(request, metadata)
		status, errorMsg = resolveAuditStatus(request)
	)
	return operlogsvc.CreateInput{
		Title:         metadata.title,
		OperSummary:   metadata.operSummary,
		RouteOwner:    metadata.routeOwner,
		RouteMethod:   metadata.routeMethod,
		RoutePath:     metadata.routePath,
		RouteDocKey:   metadata.routeDocKey,
		OperType:      inferOperType(request.Method, request.URL.Path, metadata.operLogTag),
		Method:        request.URL.Path,
		RequestMethod: request.Method,
		OperName:      operName,
		OperUrl:       request.URL.String(),
		OperIp:        request.GetClientIp(),
		OperParam:     operParam,
		JsonResult:    jsonResult,
		Status:        status,
		ErrorMsg:      errorMsg,
		CostTime:      int(time.Since(startTime).Milliseconds()),
	}
}

// buildStaticRouteAnchor builds the stable route anchor persisted with one
// audit record so display-time localization can reuse apidoc i18n resources.
func buildStaticRouteAnchor(request *ghttp.Request, handler *ghttp.HandlerItemParsed) (string, string, string, string) {
	routeOwner := "core"
	if pluginID := pluginhost.SourcePluginIDFromRequest(request); pluginID != "" {
		// Source plugins register normal GoFrame routes through the host registrar;
		// the registrar marks the request context with the source plugin id so the
		// audit record can distinguish plugin-owned routes from core routes.
		routeOwner = pluginID
	}

	routeMethod := ""
	routePath := ""
	if handler != nil {
		// g.Meta path/method tags are preferred as stable documentation anchors.
		// Runtime router values fill gaps for handlers without explicit doc tags.
		routeMethod = handler.GetMetaTag(routeMetaMethod)
		routePath = handler.GetMetaTag(routeMetaPath)
		if handler.Handler != nil && handler.Handler.Router != nil {
			if strings.TrimSpace(handler.Handler.Router.Method) != "" {
				routeMethod = handler.Handler.Router.Method
			}
			if strings.TrimSpace(handler.Handler.Router.Uri) != "" {
				routePath = handler.Handler.Router.Uri
			}
		}
	}
	if strings.TrimSpace(routeMethod) == "" && request != nil {
		routeMethod = request.Method
	}
	if strings.TrimSpace(routePath) == "" && request != nil && request.URL != nil {
		routePath = request.URL.Path
	}

	routeMethod = strings.ToUpper(strings.TrimSpace(routeMethod))
	routePath = normalizeRoutePath(routePath)
	routeDocKey := hostapidoc.BuildOperationKeyFromHandler(handler)
	if routeDocKey == "" {
		// Some dynamic or projected routes do not have a handler-derived apidoc
		// key, so fall back to the same path/method key shape used by apidoc.
		routeDocKey = hostapidoc.BuildOperationKeyFromPath(routePath, routeMethod)
	}
	return routeOwner, routeMethod, routePath, routeDocKey
}

// normalizeRoutePath canonicalizes a route path for persistent route anchors.
func normalizeRoutePath(routePath string) string {
	trimmedPath := strings.TrimSpace(routePath)
	if trimmedPath == "" {
		return ""
	}
	if !strings.HasPrefix(trimmedPath, "/") {
		trimmedPath = "/" + trimmedPath
	}
	if trimmedPath != "/" {
		trimmedPath = strings.TrimRight(trimmedPath, "/")
	}
	return trimmedPath
}

// shouldRecordAuditRequest reports whether the current request matches audit logging rules.
func shouldRecordAuditRequest(method string, operLogTag string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodDelete:
		// Mutating REST methods are always operation-log candidates.
		return true
	case http.MethodGet:
		// GET requests are recorded only when the route explicitly opts in through
		// operLog metadata, such as export or other auditable read actions.
		return strings.TrimSpace(operLogTag) != ""
	default:
		return false
	}
}

// persistOperLogRecord schedules plugin-owned operation-log persistence without blocking the request path.
func (s *serviceImpl) persistOperLogRecord(ctx context.Context, input operlogsvc.CreateInput) {
	if err := grpool.AddWithRecover(ctx, func(taskCtx context.Context) {
		if recordErr := s.createOperLogRecord(taskCtx, input); recordErr != nil {
			logger.Warningf(taskCtx, "persist operation log failed err=%v", recordErr)
		}
	}, func(taskCtx context.Context, panicErr error) {
		logger.Errorf(taskCtx, "monitor-operlog middleware panic: %v", panicErr)
	}); err != nil {
		// If the worker pool cannot accept the task, write synchronously as a
		// best-effort fallback so the audit record is not silently lost.
		logger.Warningf(ctx, "schedule operation log task failed err=%v", err)
		if recordErr := s.createOperLogRecord(ctx, input); recordErr != nil {
			logger.Warningf(ctx, "fallback persist operation log failed err=%v", recordErr)
		}
	}
}

// createOperLogRecord writes one normalized operation-log record through the plugin-owned service.
func (s *serviceImpl) createOperLogRecord(ctx context.Context, input operlogsvc.CreateInput) error {
	if s == nil || s.operLogSvc == nil {
		return nil
	}
	return s.operLogSvc.Create(ctx, input)
}

// inferOperType determines the audit operation type from HTTP method, path, and operLog tag.
func inferOperType(method string, path string, operLogTag string) operlogtype.OperType {
	if strings.TrimSpace(operLogTag) != "" {
		// Explicit operLog metadata has priority over method/path heuristics because
		// routes such as export or import may not map cleanly to a REST method.
		if operType, ok := resolveOperLogTag(operLogTag); ok {
			return operType
		}
		return operlogtype.OperTypeOther
	}

	switch method {
	case http.MethodPost:
		if strings.Contains(strings.ToLower(path), "import") {
			return operlogtype.OperTypeImport
		}
		return operlogtype.OperTypeCreate
	case http.MethodPut:
		return operlogtype.OperTypeUpdate
	case http.MethodDelete:
		return operlogtype.OperTypeDelete
	default:
		return operlogtype.OperTypeOther
	}
}

// resolveOperLogTag converts a semantic operLog tag to the published audit operation type code.
func resolveOperLogTag(tag string) (operlogtype.OperType, bool) {
	operType := operlogtype.Normalize(tag)
	return operType, operlogtype.IsSupported(operType)
}

// buildAuditRequestParam extracts the request payload snippet suitable for operation logging.
func buildAuditRequestParam(request *ghttp.Request) string {
	if isBinaryContentType(request.GetHeader("Content-Type")) {
		return operLogBinaryContent
	}
	// Request bodies are preferred by getRequestParam; query parameters are used
	// as the fallback for GET-style audited actions.
	return truncate(sanitizeOperLogParam(getRequestParam(request)), maxParamLen)
}

// buildAuditResponseResult extracts the response snippet suitable for operation logging.
func buildAuditResponseResult(request *ghttp.Request, metadata auditRouteMetadata) string {
	responseContentType := request.Response.Header().Get("Content-Type")
	responseBody := request.Response.BufferString()
	if responseContentType == "" {
		// Dynamic route dispatch can store the bridge response content type in
		// metadata when GoFrame's response header is not available here.
		responseContentType = metadata.responseContentType
	}
	if responseBody == "" {
		// For dynamic plugin routes, the runtime dispatcher captures the raw bridge
		// body on metadata before writing it to the HTTP response.
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
