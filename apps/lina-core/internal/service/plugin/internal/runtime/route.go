// This file defines the fixed-prefix dynamic route matcher and request dispatch
// runtime used by dynamic plugin REST execution.

package runtime

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/golang-jwt/jwt/v5"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/pkg/logger"
	"lina-core/pkg/pluginbridge"
	"lina-core/pkg/pluginhost"
)

// RoutePublicPrefix is the fixed host URL prefix for all dynamic plugin routes.
const RoutePublicPrefix = "/api/v1/extensions"

const (
	dynamicRouteCtxVarState    gctx.StrKey = "plugin_dynamic_route_state"
	dynamicRouteCtxVarIdentity gctx.StrKey = "plugin_dynamic_route_identity"
	dynamicRouteCtxVarOperLog  gctx.StrKey = "plugin_dynamic_route_operlog"

	// statusNormal represents the normal/enabled status for role and menu queries.
	statusNormal = 1
)

// DynamicRouteDispatchInput describes one host-side dynamic route dispatch call.
type DynamicRouteDispatchInput struct {
	// Request is the original GoFrame request that entered the fixed public prefix.
	Request *ghttp.Request
}

// DynamicRouteOperLogMetadata stores operation-log metadata synthesized from one
// matched dynamic route so host middleware can reuse the standard logging chain.
type DynamicRouteOperLogMetadata struct {
	// Title is the route tag projection used as the operation-log title.
	Title string
	// Summary is the route summary projected into the operation-log record.
	Summary string
	// OperLogTag is the normalized operation-log tag reused by host logging rules.
	OperLogTag string
	// ResponseBody stores the raw bridge response body for middleware-side logging.
	ResponseBody string
	// ResponseContentType stores the resolved content type of the bridge response.
	ResponseContentType string
}

type dynamicRouteMatch struct {
	PluginID     string
	PublicPath   string
	InternalPath string
	Route        *pluginbridge.RouteContract
	PathParams   map[string]string
}

type dynamicRouteRuntimeState struct {
	Manifest *catalog.Manifest
	Match    *dynamicRouteMatch
}

// DynamicRouteMatch is the exported form of dynamicRouteMatch for cross-package access.
type DynamicRouteMatch = dynamicRouteMatch

// DynamicRouteRuntimeState is the exported form of dynamicRouteRuntimeState for cross-package access.
type DynamicRouteRuntimeState = dynamicRouteRuntimeState

// MatchDynamicRoutePath is the exported form of matchDynamicRoutePath for cross-package access.
func MatchDynamicRoutePath(routePath string, actualPath string) (map[string]string, bool) {
	return matchDynamicRoutePath(routePath, actualPath)
}

// BuildDynamicRouteOperLogMetadata is the exported form of buildDynamicRouteOperLogMetadata for cross-package access.
func BuildDynamicRouteOperLogMetadata(runtimeState *DynamicRouteRuntimeState) *DynamicRouteOperLogMetadata {
	return buildDynamicRouteOperLogMetadata(runtimeState)
}

type dynamicRouteClaims struct {
	TokenId  string `json:"tokenId"`
	UserId   int    `json:"userId"`
	Username string `json:"username"`
	Status   int    `json:"status"`
	jwt.RegisteredClaims
}

type dynamicRouteAccessContext struct {
	Permissions  []string
	RoleNames    []string
	IsSuperAdmin bool
}

// RegisterDynamicRouteDispatcher binds the fixed-prefix dispatcher into one host
// router group so dynamic routes reuse the standard RouterGroup registration flow.
func (s *serviceImpl) RegisterDynamicRouteDispatcher(group *ghttp.RouterGroup) {
	if group == nil {
		return
	}
	group.ALL("/*dynamicPath", func(r *ghttp.Request) {
		s.handleDynamicRouteRequest(r)
	})
}

// PrepareDynamicRouteMiddleware resolves the active dynamic route contract and
// caches host-owned runtime state on the request before later middlewares run.
func (s *serviceImpl) PrepareDynamicRouteMiddleware(r *ghttp.Request) {
	if r == nil {
		return
	}
	runtimeState, failure, err := s.prepareDynamicRouteRuntime(r.Context(), r)
	if err != nil {
		s.writeDynamicRouteResponse(r, pluginbridge.NewInternalErrorResponse(err.Error()))
		r.ExitAll()
		return
	}
	if failure != nil {
		s.writeDynamicRouteResponse(r, failure)
		r.ExitAll()
		return
	}
	setDynamicRouteRuntimeState(r, runtimeState)
	setDynamicRouteOperLogMetadata(r, buildDynamicRouteOperLogMetadata(runtimeState))
	r.Middleware.Next()
}

// AuthenticateDynamicRouteMiddleware applies host-owned login and permission
// governance for the matched dynamic route before bridge execution starts.
func (s *serviceImpl) AuthenticateDynamicRouteMiddleware(r *ghttp.Request) {
	if r == nil {
		return
	}
	runtimeState := getDynamicRouteRuntimeState(r)
	if runtimeState == nil {
		s.writeDynamicRouteResponse(
			r,
			pluginbridge.NewInternalErrorResponse("Dynamic route runtime state is missing"),
		)
		r.ExitAll()
		return
	}

	identity, failure, err := s.authorizeDynamicRouteRequest(r.Context(), runtimeState, r)
	if err != nil {
		s.writeDynamicRouteResponse(r, pluginbridge.NewInternalErrorResponse(err.Error()))
		r.ExitAll()
		return
	}
	if failure != nil {
		s.writeDynamicRouteResponse(r, failure)
		r.ExitAll()
		return
	}
	if identity != nil {
		setDynamicRouteIdentitySnapshot(r, identity)
	}
	r.Middleware.Next()
}

func (s *serviceImpl) handleDynamicRouteRequest(r *ghttp.Request) {
	if r == nil {
		return
	}
	runtimeState := getDynamicRouteRuntimeState(r)
	if runtimeState == nil || runtimeState.Match == nil || runtimeState.Manifest == nil {
		s.writeDynamicRouteResponse(r, pluginbridge.NewInternalErrorResponse("Dynamic route runtime state is missing"))
		r.ExitAll()
		return
	}

	response, err := s.executePreparedDynamicRoute(
		r.Context(),
		runtimeState,
		getDynamicRouteIdentitySnapshot(r),
		r,
	)
	if err != nil {
		response = pluginbridge.NewInternalErrorResponse(err.Error())
	}
	if response == nil {
		response = pluginbridge.NewInternalErrorResponse("Dynamic route dispatcher returned nil response")
	}
	s.writeDynamicRouteResponse(r, response)
	r.ExitAll()
}

// DispatchDynamicRoute dispatches one fixed-prefix request into the active release
// of one dynamic plugin. Matching always happens against the archived active manifest
// so staged uploads cannot affect live traffic before reconcile.
func (s *serviceImpl) DispatchDynamicRoute(
	ctx context.Context,
	in *DynamicRouteDispatchInput,
) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	if in == nil || in.Request == nil {
		return pluginbridge.NewBadRequestResponse("Dynamic route request is missing"), nil
	}

	runtimeState, failure, err := s.prepareDynamicRouteRuntime(ctx, in.Request)
	if err != nil {
		return nil, err
	}
	if failure != nil {
		return failure, nil
	}
	identity, failure, err := s.authorizeDynamicRouteRequest(ctx, runtimeState, in.Request)
	if err != nil {
		return nil, err
	}
	if failure != nil {
		return failure, nil
	}
	return s.executePreparedDynamicRoute(ctx, runtimeState, identity, in.Request)
}

// matchDynamicRoute resolves the fixed `/api/v1/extensions/{pluginId}/...`
// public path to the plugin-declared internal route contract.
func (s *serviceImpl) matchDynamicRoute(ctx context.Context, request *ghttp.Request) (*dynamicRouteMatch, error) {
	publicPath := strings.TrimSpace(request.URL.Path)
	if !strings.HasPrefix(publicPath, RoutePublicPrefix+"/") {
		return nil, nil
	}
	pathSuffix := strings.TrimPrefix(publicPath, RoutePublicPrefix+"/")
	segments := strings.Split(pathSuffix, "/")
	if len(segments) == 0 || strings.TrimSpace(segments[0]) == "" {
		return nil, gerror.New("动态插件路径缺少 pluginId")
	}
	pluginID := strings.TrimSpace(segments[0])
	internalPath := "/"
	if len(segments) > 1 {
		internalPath = "/" + strings.Join(segments[1:], "/")
	}
	if internalPath == "/" && strings.HasSuffix(publicPath, "/") {
		internalPath = "/"
	}

	manifest, err := s.catalogSvc.GetActiveManifest(ctx, pluginID)
	if err != nil {
		return nil, nil
	}
	if manifest == nil || len(manifest.Routes) == 0 {
		return nil, nil
	}

	method := strings.ToUpper(strings.TrimSpace(request.Method))
	for _, route := range manifest.Routes {
		params, ok := matchDynamicRoutePath(route.Path, internalPath)
		if !ok {
			continue
		}
		if strings.ToUpper(strings.TrimSpace(route.Method)) != method {
			continue
		}
		return &dynamicRouteMatch{
			PluginID:     pluginID,
			PublicPath:   publicPath,
			InternalPath: internalPath,
			Route:        route,
			PathParams:   params,
		}, nil
	}
	return nil, nil
}

func matchDynamicRoutePath(routePath string, actualPath string) (map[string]string, bool) {
	normalizedRoute := normalizeDynamicRoutePath(routePath)
	normalizedActual := normalizeDynamicRoutePath(actualPath)
	routeSegments := strings.Split(strings.TrimPrefix(normalizedRoute, "/"), "/")
	actualSegments := strings.Split(strings.TrimPrefix(normalizedActual, "/"), "/")
	if normalizedRoute == "/" {
		routeSegments = []string{}
	}
	if normalizedActual == "/" {
		actualSegments = []string{}
	}
	if len(routeSegments) != len(actualSegments) {
		return nil, false
	}

	params := make(map[string]string)
	for index, routeSegment := range routeSegments {
		actualSegment := actualSegments[index]
		if strings.HasPrefix(routeSegment, "{") && strings.HasSuffix(routeSegment, "}") {
			paramName := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(routeSegment, "{"), "}"))
			if paramName == "" {
				return nil, false
			}
			params[paramName] = actualSegment
			continue
		}
		if routeSegment != actualSegment {
			return nil, false
		}
	}
	return params, true
}

func normalizeDynamicRoutePath(path string) string {
	normalized := strings.TrimSpace(path)
	if normalized == "" {
		return "/"
	}
	if !strings.HasPrefix(normalized, "/") {
		normalized = "/" + normalized
	}
	if len(normalized) > 1 {
		normalized = strings.TrimSuffix(normalized, "/")
	}
	return normalized
}

func (s *serviceImpl) prepareDynamicRouteRuntime(
	ctx context.Context,
	request *ghttp.Request,
) (*dynamicRouteRuntimeState, *pluginbridge.BridgeResponseEnvelopeV1, error) {
	if request == nil {
		return nil, pluginbridge.NewBadRequestResponse("Dynamic route request is missing"), nil
	}

	match, err := s.matchDynamicRoute(ctx, request)
	if err != nil {
		return nil, pluginbridge.NewBadRequestResponse(err.Error()), nil
	}
	if match == nil || match.Route == nil {
		return nil, pluginbridge.NewNotFoundResponse("Dynamic route not found"), nil
	}

	manifest, err := s.catalogSvc.GetActiveManifest(ctx, match.PluginID)
	if err != nil {
		return nil, pluginbridge.NewNotFoundResponse(err.Error()), nil
	}
	registry, err := s.catalogSvc.GetRegistry(ctx, match.PluginID)
	if err != nil {
		return nil, nil, err
	}
	if registry == nil || registry.Installed != catalog.InstalledYes || registry.Status != catalog.StatusEnabled {
		return nil, pluginbridge.NewNotFoundResponse("Dynamic plugin is not enabled"), nil
	}
	return &dynamicRouteRuntimeState{
		Manifest: manifest,
		Match:    match,
	}, nil, nil
}

func (s *serviceImpl) authorizeDynamicRouteRequest(
	ctx context.Context,
	runtimeState *dynamicRouteRuntimeState,
	request *ghttp.Request,
) (*pluginbridge.IdentitySnapshotV1, *pluginbridge.BridgeResponseEnvelopeV1, error) {
	if runtimeState == nil || runtimeState.Match == nil || runtimeState.Match.Route == nil {
		return nil, pluginbridge.NewInternalErrorResponse("Dynamic route runtime state is incomplete"), nil
	}
	if runtimeState.Match.Route.Access != pluginbridge.AccessLogin {
		return nil, nil, nil
	}
	return s.buildDynamicRouteIdentitySnapshot(ctx, runtimeState.Match, request)
}

func (s *serviceImpl) executePreparedDynamicRoute(
	ctx context.Context,
	runtimeState *dynamicRouteRuntimeState,
	identity *pluginbridge.IdentitySnapshotV1,
	request *ghttp.Request,
) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	if runtimeState == nil || runtimeState.Match == nil || runtimeState.Manifest == nil {
		return pluginbridge.NewInternalErrorResponse("Dynamic route runtime state is incomplete"), nil
	}

	requestEnvelope, err := s.buildDynamicRouteRequestEnvelopeWithIdentity(
		runtimeState.Match,
		request,
		identity,
	)
	if err != nil {
		return nil, err
	}
	if runtimeState.Manifest.BridgeSpec == nil || !runtimeState.Manifest.BridgeSpec.RouteExecution {
		return pluginbridge.NewFailureResponse(
			http.StatusNotImplemented,
			"BRIDGE_NOT_IMPLEMENTED",
			"Dynamic route bridge is not executable for the active plugin release",
		), nil
	}
	return s.executeDynamicRoute(ctx, runtimeState.Manifest, requestEnvelope)
}

func (s *serviceImpl) buildDynamicRouteRequestEnvelopeWithIdentity(
	match *dynamicRouteMatch,
	request *ghttp.Request,
	identity *pluginbridge.IdentitySnapshotV1,
) (*pluginbridge.BridgeRequestEnvelopeV1, error) {
	body := request.GetBody()
	queryValues := request.URL.Query()
	return &pluginbridge.BridgeRequestEnvelopeV1{
		PluginID: match.PluginID,
		Route: &pluginbridge.RouteMatchSnapshotV1{
			Method:       strings.ToUpper(strings.TrimSpace(request.Method)),
			PublicPath:   match.PublicPath,
			InternalPath: match.InternalPath,
			RoutePath:    match.Route.Path,
			Access:       match.Route.Access,
			Permission:   match.Route.Permission,
			RequestType:  match.Route.RequestType,
			PathParams:   cloneStringMap(match.PathParams),
			QueryValues:  cloneURLValues(queryValues),
		},
		Request: &pluginbridge.HTTPRequestSnapshotV1{
			Method:       strings.ToUpper(strings.TrimSpace(request.Method)),
			PublicPath:   match.PublicPath,
			InternalPath: match.InternalPath,
			RawPath:      request.URL.Path,
			RawQuery:     request.URL.RawQuery,
			Host:         request.Host,
			Scheme:       request.URL.Scheme,
			RemoteAddr:   request.Request.RemoteAddr,
			ClientIP:     request.GetClientIp(),
			ContentType:  request.Header.Get("Content-Type"),
			Headers:      sanitizeDynamicRouteHeaders(request.Header),
			Cookies:      collectRequestCookies(request),
			Body:         append([]byte(nil), body...),
		},
		Identity:  identity,
		RequestID: buildDynamicRequestID(match, request),
	}, nil
}

// buildDynamicRouteIdentitySnapshot validates session state and permission grants
// on the host side before forwarding the request into guest code.
func (s *serviceImpl) buildDynamicRouteIdentitySnapshot(
	ctx context.Context,
	match *dynamicRouteMatch,
	request *ghttp.Request,
) (*pluginbridge.IdentitySnapshotV1, *pluginbridge.BridgeResponseEnvelopeV1, error) {
	tokenHeader := strings.TrimSpace(request.GetHeader("Authorization"))
	if tokenHeader == "" {
		return nil, pluginbridge.NewUnauthorizedResponse("Missing Authorization header"), nil
	}
	tokenString := strings.TrimSpace(strings.TrimPrefix(tokenHeader, "Bearer "))
	if tokenString == "" || tokenString == tokenHeader {
		return nil, pluginbridge.NewUnauthorizedResponse("Invalid bearer token"), nil
	}
	claims, err := s.parseDynamicRouteToken(ctx, tokenString)
	if err != nil {
		return nil, pluginbridge.NewUnauthorizedResponse(err.Error()), nil
	}
	exists, err := s.touchDynamicRouteSession(ctx, claims.TokenId)
	if err != nil {
		return nil, nil, err
	}
	if !exists {
		return nil, pluginbridge.NewUnauthorizedResponse("Session has expired"), nil
	}

	if s.userCtx != nil {
		s.userCtx.SetUser(ctx, claims.TokenId, claims.UserId, claims.Username, claims.Status)
	}
	if s.afterAuth != nil {
		s.afterAuth.DispatchAfterAuth(
			ctx,
			pluginhost.NewAfterAuthInput(
				request,
				claims.TokenId,
				claims.UserId,
				claims.Username,
				claims.Status,
			),
		)
	}
	accessContext, err := s.getDynamicRouteAccessContext(ctx, claims.UserId)
	if err != nil {
		return nil, nil, err
	}
	if match.Route.Permission != "" && !hasDynamicRoutePermission(accessContext, match.Route.Permission) {
		return nil, pluginbridge.NewForbiddenResponse("Permission denied"), nil
	}

	return &pluginbridge.IdentitySnapshotV1{
		TokenID:      claims.TokenId,
		UserID:       int32(claims.UserId),
		Username:     claims.Username,
		Status:       int32(claims.Status),
		Permissions:  append([]string(nil), accessContext.Permissions...),
		RoleNames:    append([]string(nil), accessContext.RoleNames...),
		IsSuperAdmin: accessContext.IsSuperAdmin,
	}, nil, nil
}

func (s *serviceImpl) parseDynamicRouteToken(ctx context.Context, tokenString string) (*dynamicRouteClaims, error) {
	secret := ""
	if s.jwtConfig != nil {
		secret = s.jwtConfig.GetJwtSecret(ctx)
	}
	token, err := jwt.ParseWithClaims(tokenString, &dynamicRouteClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, gerror.New("无效的Token")
	}
	claims, ok := token.Claims.(*dynamicRouteClaims)
	if !ok || !token.Valid {
		return nil, gerror.New("无效的Token")
	}
	return claims, nil
}

func (s *serviceImpl) touchDynamicRouteSession(ctx context.Context, tokenID string) (bool, error) {
	result, err := dao.SysOnlineSession.Ctx(ctx).
		Where(do.SysOnlineSession{TokenId: tokenID}).
		Data(do.SysOnlineSession{LastActiveTime: gtime.Now()}).Update()
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	if affected > 0 {
		return true, nil
	}

	// DATETIME precision in sys_online_session is second-level. When the dynamic
	// route is hit within the same second, MySQL reports zero affected rows even
	// though the session still exists and remains valid.
	count, err := dao.SysOnlineSession.Ctx(ctx).
		Where(do.SysOnlineSession{TokenId: tokenID}).
		Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *serviceImpl) getDynamicRouteAccessContext(ctx context.Context, userID int) (*dynamicRouteAccessContext, error) {
	roleIDs, err := s.getDynamicRouteUserRoleIDs(ctx, userID)
	if err != nil {
		return nil, err
	}
	roleNames, err := s.getDynamicRouteRoleNames(ctx, roleIDs)
	if err != nil {
		return nil, err
	}
	permissions, err := s.getDynamicRoutePermissionsByRoleIDs(ctx, roleIDs)
	if err != nil {
		return nil, err
	}
	return &dynamicRouteAccessContext{
		Permissions:  permissions,
		RoleNames:    roleNames,
		IsSuperAdmin: containsInt(roleIDs, 1),
	}, nil
}

func (s *serviceImpl) getDynamicRouteUserRoleIDs(ctx context.Context, userID int) ([]int, error) {
	items := make([]*entity.SysUserRole, 0)
	if err := dao.SysUserRole.Ctx(ctx).
		Where(do.SysUserRole{UserId: userID}).
		Scan(&items); err != nil {
		return nil, err
	}
	roleIDs := make([]int, 0, len(items))
	seen := make(map[int]struct{}, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		if _, ok := seen[item.RoleId]; ok {
			continue
		}
		seen[item.RoleId] = struct{}{}
		roleIDs = append(roleIDs, item.RoleId)
	}
	return roleIDs, nil
}

func (s *serviceImpl) getDynamicRouteRoleNames(ctx context.Context, roleIDs []int) ([]string, error) {
	if len(roleIDs) == 0 {
		return []string{}, nil
	}
	items := make([]*entity.SysRole, 0)
	if err := dao.SysRole.Ctx(ctx).
		WhereIn(dao.SysRole.Columns().Id, intsToInterfaces(roleIDs)).
		Where(dao.SysRole.Columns().Status, statusNormal).
		Scan(&items); err != nil {
		return nil, err
	}
	roleNames := make([]string, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		roleNames = append(roleNames, item.Name)
	}
	return roleNames, nil
}

// getDynamicRoutePermissionsByRoleIDs merges the role-menu and menu-permission
// lookups into a single pass: it fetches menu IDs bound to the given roles, then
// loads only button-type permission menus in one query (3 DB queries total for
// the full access context instead of 5).
func (s *serviceImpl) getDynamicRoutePermissionsByRoleIDs(ctx context.Context, roleIDs []int) ([]string, error) {
	if len(roleIDs) == 0 {
		return []string{}, nil
	}
	roleMenuItems := make([]*entity.SysRoleMenu, 0)
	if err := dao.SysRoleMenu.Ctx(ctx).
		WhereIn(dao.SysRoleMenu.Columns().RoleId, intsToInterfaces(roleIDs)).
		Scan(&roleMenuItems); err != nil {
		return nil, err
	}
	menuIDs := make([]int, 0, len(roleMenuItems))
	seen := make(map[int]struct{}, len(roleMenuItems))
	for _, item := range roleMenuItems {
		if item == nil {
			continue
		}
		if _, ok := seen[item.MenuId]; ok {
			continue
		}
		seen[item.MenuId] = struct{}{}
		menuIDs = append(menuIDs, item.MenuId)
	}
	if len(menuIDs) == 0 {
		return []string{}, nil
	}
	menuItems := make([]*entity.SysMenu, 0)
	if err := dao.SysMenu.Ctx(ctx).
		WhereIn(dao.SysMenu.Columns().Id, intsToInterfaces(menuIDs)).
		Where(dao.SysMenu.Columns().Type, catalog.MenuTypeButton.String()).
		Where(dao.SysMenu.Columns().Status, statusNormal).
		Scan(&menuItems); err != nil {
		return nil, err
	}
	if s.menuFilter != nil {
		menuItems = s.menuFilter.FilterPermissionMenus(ctx, menuItems)
	}
	permissions := make([]string, 0, len(menuItems))
	for _, item := range menuItems {
		if item == nil || strings.TrimSpace(item.Perms) == "" {
			continue
		}
		permissions = append(permissions, item.Perms)
	}
	return permissions, nil
}

func hasDynamicRoutePermission(accessContext *dynamicRouteAccessContext, permission string) bool {
	if accessContext == nil {
		return false
	}
	if accessContext.IsSuperAdmin {
		return true
	}
	for _, item := range accessContext.Permissions {
		if strings.TrimSpace(item) == strings.TrimSpace(permission) {
			return true
		}
	}
	return false
}

func containsInt(values []int, target int) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func intsToInterfaces(values []int) []interface{} {
	items := make([]interface{}, 0, len(values))
	for _, value := range values {
		items = append(items, value)
	}
	return items
}

func sanitizeDynamicRouteHeaders(headers http.Header) map[string][]string {
	result := make(map[string][]string)
	if len(headers) == 0 {
		return result
	}
	keys := make([]string, 0, len(headers))
	for key := range headers {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if strings.EqualFold(key, "Authorization") {
			continue
		}
		values := headers.Values(key)
		if len(values) == 0 {
			continue
		}
		result[key] = append([]string(nil), values...)
	}
	return result
}

func collectRequestCookies(request *ghttp.Request) map[string]string {
	result := make(map[string]string)
	if request == nil || request.Request == nil {
		return result
	}
	for _, cookie := range request.Request.Cookies() {
		if cookie == nil {
			continue
		}
		result[cookie.Name] = cookie.Value
	}
	return result
}

func cloneURLValues(values url.Values) map[string][]string {
	result := make(map[string][]string, len(values))
	for key, items := range values {
		result[key] = append([]string(nil), items...)
	}
	return result
}

func cloneStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return map[string]string{}
	}
	result := make(map[string]string, len(values))
	for key, value := range values {
		result[key] = value
	}
	return result
}

func buildDynamicRequestID(match *dynamicRouteMatch, request *ghttp.Request) string {
	if request == nil {
		return match.PluginID + ":" + base64.StdEncoding.EncodeToString([]byte(match.InternalPath))
	}
	return match.PluginID + ":" + request.Method + ":" + match.InternalPath
}

func setDynamicRouteRuntimeState(request *ghttp.Request, runtimeState *dynamicRouteRuntimeState) {
	if request == nil {
		return
	}
	request.SetCtxVar(dynamicRouteCtxVarState, runtimeState)
}

func getDynamicRouteRuntimeState(request *ghttp.Request) *dynamicRouteRuntimeState {
	if request == nil {
		return nil
	}
	value := request.GetCtxVar(dynamicRouteCtxVarState).Val()
	if value == nil {
		return nil
	}
	runtimeState, _ := value.(*dynamicRouteRuntimeState)
	return runtimeState
}

func setDynamicRouteIdentitySnapshot(request *ghttp.Request, identity *pluginbridge.IdentitySnapshotV1) {
	if request == nil {
		return
	}
	request.SetCtxVar(dynamicRouteCtxVarIdentity, identity)
}

func getDynamicRouteIdentitySnapshot(request *ghttp.Request) *pluginbridge.IdentitySnapshotV1 {
	if request == nil {
		return nil
	}
	value := request.GetCtxVar(dynamicRouteCtxVarIdentity).Val()
	if value == nil {
		return nil
	}
	identity, _ := value.(*pluginbridge.IdentitySnapshotV1)
	return identity
}

func setDynamicRouteOperLogMetadata(request *ghttp.Request, metadata *DynamicRouteOperLogMetadata) {
	if request == nil || metadata == nil {
		return
	}
	request.SetCtxVar(dynamicRouteCtxVarOperLog, metadata)
}

func buildDynamicRouteOperLogMetadata(runtimeState *dynamicRouteRuntimeState) *DynamicRouteOperLogMetadata {
	if runtimeState == nil || runtimeState.Match == nil || runtimeState.Match.Route == nil {
		return nil
	}
	metadata := &DynamicRouteOperLogMetadata{
		Title:   strings.Join(runtimeState.Match.Route.Tags, ","),
		Summary: strings.TrimSpace(runtimeState.Match.Route.Summary),
	}
	if runtimeState.Match.Route.OperLog != "" {
		metadata.OperLogTag = runtimeState.Match.Route.OperLog
	}
	return metadata
}

// GetDynamicRouteOperLogMetadata returns dynamic-route operation-log metadata
// attached during the host middleware chain.
func GetDynamicRouteOperLogMetadata(request *ghttp.Request) *DynamicRouteOperLogMetadata {
	if request == nil {
		return nil
	}
	value := request.GetCtxVar(dynamicRouteCtxVarOperLog).Val()
	if value == nil {
		return nil
	}
	metadata, _ := value.(*DynamicRouteOperLogMetadata)
	return metadata
}

// writeDynamicRouteResponse writes the guest response back without going through
// GoFrame's default success wrapper, otherwise raw plugin payloads would be
// polluted by host-managed response formatting.
func (s *serviceImpl) writeDynamicRouteResponse(request *ghttp.Request, response *pluginbridge.BridgeResponseEnvelopeV1) {
	if request == nil || response == nil {
		return
	}
	metadata := GetDynamicRouteOperLogMetadata(request)
	if metadata != nil {
		metadata.ResponseBody = string(response.Body)
		metadata.ResponseContentType = strings.TrimSpace(response.ContentType)
	}
	for key, values := range response.Headers {
		for _, value := range values {
			request.Response.Header().Add(key, value)
		}
	}
	if strings.TrimSpace(response.ContentType) != "" {
		request.Response.Header().Set("Content-Type", response.ContentType)
	}
	statusCode := int(response.StatusCode)
	if statusCode <= 0 {
		statusCode = http.StatusOK
	}
	// RawWriter preserves the exact status/body emitted by the bridge envelope.
	request.Response.RawWriter().WriteHeader(statusCode)
	if len(response.Body) > 0 {
		if _, err := request.Response.RawWriter().Write(response.Body); err != nil {
			logger.Warningf(request.Context(), "write dynamic route response body failed err=%v", err)
		}
	}
}
