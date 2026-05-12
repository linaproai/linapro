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
	"lina-core/internal/service/datascope"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/pkg/logger"
	bridgecodec "lina-core/pkg/pluginbridge/codec"
	bridgecontract "lina-core/pkg/pluginbridge/contract"
)

// RoutePublicPrefix is the fixed host URL prefix for all dynamic plugin routes.
const RoutePublicPrefix = "/api/v1/extensions"

// Request-context keys and sentinel values used by the dynamic route pipeline.
const (
	dynamicRouteCtxVarState    gctx.StrKey = "plugin_dynamic_route_state"
	dynamicRouteCtxVarIdentity gctx.StrKey = "plugin_dynamic_route_identity"
	dynamicRouteCtxVarMetadata gctx.StrKey = "plugin_dynamic_route_metadata"

	// statusNormal represents the normal/enabled status for role and menu queries.
	statusNormal = 1
)

// DynamicRouteDispatchInput describes one host-side dynamic route dispatch call.
type DynamicRouteDispatchInput struct {
	// Request is the original GoFrame request that entered the fixed public prefix.
	Request *ghttp.Request
}

// DynamicRouteMetadata stores generic metadata synthesized from one matched
// dynamic route for downstream source-plugin middleware.
type DynamicRouteMetadata struct {
	// PluginID is the dynamic plugin that owns the matched route.
	PluginID string
	// Method is the declared dynamic route HTTP method.
	Method string
	// PublicPath is the public host path matched by the request.
	PublicPath string
	// Tags are the route tags declared by the dynamic plugin manifest.
	Tags []string
	// Summary is the route summary declared by the dynamic plugin manifest.
	Summary string
	// Meta contains additional route declaration metadata by source tag name.
	Meta map[string]string
	// ResponseBody stores the raw bridge response body for middleware-side logging.
	ResponseBody string
	// ResponseContentType stores the resolved content type of the bridge response.
	ResponseContentType string
}

// dynamicRouteMatch stores the resolved plugin route and path parameters for one request.
type dynamicRouteMatch struct {
	PluginID     string
	PublicPath   string
	InternalPath string
	Route        *bridgecontract.RouteContract
	PathParams   map[string]string
}

// dynamicRouteRuntimeState stores the active manifest and route match cached on the request.
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

// BuildDynamicRouteMetadata is the exported form of buildDynamicRouteMetadata for cross-package access.
func BuildDynamicRouteMetadata(runtimeState *DynamicRouteRuntimeState) *DynamicRouteMetadata {
	return buildDynamicRouteMetadata(runtimeState)
}

// dynamicRouteClaims mirrors the JWT claims needed by host-side dynamic route auth.
type dynamicRouteClaims struct {
	TokenId         string `json:"tokenId"`
	TokenType       string `json:"tokenType"`
	TenantId        int    `json:"tenantId"`
	UserId          int    `json:"userId"`
	Username        string `json:"username"`
	Status          int    `json:"status"`
	ActingUserId    int    `json:"actingUserId"`
	ActingAsTenant  bool   `json:"actingAsTenant"`
	IsImpersonation bool   `json:"isImpersonation"`
	jwt.RegisteredClaims
}

// dynamicRouteAccessContext stores role-derived access data used by permission checks.
type dynamicRouteAccessContext struct {
	Permissions          []string
	RoleNames            []string
	DataScope            int
	DataScopeUnsupported bool
	UnsupportedDataScope int
	IsSuperAdmin         bool
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
		s.writeDynamicRouteResponse(r, bridgecodec.NewInternalErrorResponse(err.Error()))
		r.ExitAll()
		return
	}
	if failure != nil {
		s.writeDynamicRouteResponse(r, failure)
		r.ExitAll()
		return
	}
	setDynamicRouteRuntimeState(r, runtimeState)
	setDynamicRouteMetadata(r, buildDynamicRouteMetadata(runtimeState))
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
			bridgecodec.NewInternalErrorResponse("Dynamic route runtime state is missing"),
		)
		r.ExitAll()
		return
	}

	identity, failure, err := s.authorizeDynamicRouteRequest(r.Context(), runtimeState, r)
	if err != nil {
		s.writeDynamicRouteResponse(r, bridgecodec.NewInternalErrorResponse(err.Error()))
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

// handleDynamicRouteRequest executes the prepared dynamic route after earlier
// middleware stages cached route and identity state on the request.
func (s *serviceImpl) handleDynamicRouteRequest(r *ghttp.Request) {
	if r == nil {
		return
	}
	runtimeState := getDynamicRouteRuntimeState(r)
	if runtimeState == nil || runtimeState.Match == nil || runtimeState.Manifest == nil {
		s.writeDynamicRouteResponse(r, bridgecodec.NewInternalErrorResponse("Dynamic route runtime state is missing"))
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
		response = bridgecodec.NewInternalErrorResponse(err.Error())
	}
	if response == nil {
		response = bridgecodec.NewInternalErrorResponse("Dynamic route dispatcher returned nil response")
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
) (*bridgecontract.BridgeResponseEnvelopeV1, error) {
	if in == nil || in.Request == nil {
		return bridgecodec.NewBadRequestResponse("Dynamic route request is missing"), nil
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
		return nil, gerror.New("dynamic plugin path is missing pluginId")
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

// matchDynamicRoutePath compares one declared route template against the
// actual internal path and returns extracted path params when it matches.
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

// normalizeDynamicRoutePath canonicalizes route paths for matching.
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

// prepareDynamicRouteRuntime resolves the active manifest and matched route for
// one incoming fixed-prefix request.
func (s *serviceImpl) prepareDynamicRouteRuntime(
	ctx context.Context,
	request *ghttp.Request,
) (*dynamicRouteRuntimeState, *bridgecontract.BridgeResponseEnvelopeV1, error) {
	if request == nil {
		return nil, bridgecodec.NewBadRequestResponse("Dynamic route request is missing"), nil
	}

	match, err := s.matchDynamicRoute(ctx, request)
	if err != nil {
		return nil, bridgecodec.NewBadRequestResponse(err.Error()), nil
	}
	if match == nil || match.Route == nil {
		return nil, bridgecodec.NewNotFoundResponse("Dynamic route not found"), nil
	}

	manifest, err := s.catalogSvc.GetActiveManifest(ctx, match.PluginID)
	if err != nil {
		return nil, bridgecodec.NewNotFoundResponse(err.Error()), nil
	}
	registry, err := s.catalogSvc.GetRegistry(ctx, match.PluginID)
	if err != nil {
		return nil, nil, err
	}
	if registry == nil || registry.Installed != catalog.InstalledYes || registry.Status != catalog.StatusEnabled {
		return nil, bridgecodec.NewNotFoundResponse("Dynamic plugin is not enabled"), nil
	}
	if s.menuFilter != nil && !s.menuFilter.IsEnabled(ctx, match.PluginID) {
		return nil, bridgecodec.NewNotFoundResponse("Dynamic plugin is not enabled"), nil
	}
	return &dynamicRouteRuntimeState{
		Manifest: manifest,
		Match:    match,
	}, nil, nil
}

// authorizeDynamicRouteRequest applies host-side login and permission checks
// for the matched dynamic route.
func (s *serviceImpl) authorizeDynamicRouteRequest(
	ctx context.Context,
	runtimeState *dynamicRouteRuntimeState,
	request *ghttp.Request,
) (*bridgecontract.IdentitySnapshotV1, *bridgecontract.BridgeResponseEnvelopeV1, error) {
	if runtimeState == nil || runtimeState.Match == nil || runtimeState.Match.Route == nil {
		return nil, bridgecodec.NewInternalErrorResponse("Dynamic route runtime state is incomplete"), nil
	}
	if runtimeState.Match.Route.Access != bridgecontract.AccessLogin {
		return nil, nil, nil
	}
	return s.buildDynamicRouteIdentitySnapshot(ctx, runtimeState.Match, request)
}

// executePreparedDynamicRoute builds the bridge request envelope and invokes
// the runtime executor for the matched active route.
func (s *serviceImpl) executePreparedDynamicRoute(
	ctx context.Context,
	runtimeState *dynamicRouteRuntimeState,
	identity *bridgecontract.IdentitySnapshotV1,
	request *ghttp.Request,
) (*bridgecontract.BridgeResponseEnvelopeV1, error) {
	if runtimeState == nil || runtimeState.Match == nil || runtimeState.Manifest == nil {
		return bridgecodec.NewInternalErrorResponse("Dynamic route runtime state is incomplete"), nil
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
		return bridgecodec.NewFailureResponse(
			http.StatusNotImplemented,
			"BRIDGE_NOT_IMPLEMENTED",
			"Dynamic route bridge is not executable for the active plugin release",
		), nil
	}
	return s.executeDynamicRoute(ctx, runtimeState.Manifest, requestEnvelope)
}

// buildDynamicRouteRequestEnvelopeWithIdentity snapshots the matched request
// into the bridge payload forwarded to guest code.
func (s *serviceImpl) buildDynamicRouteRequestEnvelopeWithIdentity(
	match *dynamicRouteMatch,
	request *ghttp.Request,
	identity *bridgecontract.IdentitySnapshotV1,
) (*bridgecontract.BridgeRequestEnvelopeV1, error) {
	body := request.GetBody()
	queryValues := request.URL.Query()
	return &bridgecontract.BridgeRequestEnvelopeV1{
		PluginID: match.PluginID,
		Route: &bridgecontract.RouteMatchSnapshotV1{
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
		Request: &bridgecontract.HTTPRequestSnapshotV1{
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
) (*bridgecontract.IdentitySnapshotV1, *bridgecontract.BridgeResponseEnvelopeV1, error) {
	tokenHeader := strings.TrimSpace(request.GetHeader("Authorization"))
	if tokenHeader == "" {
		return nil, bridgecodec.NewUnauthorizedResponse("Missing Authorization header"), nil
	}
	tokenString := strings.TrimSpace(strings.TrimPrefix(tokenHeader, "Bearer "))
	if tokenString == "" || tokenString == tokenHeader {
		return nil, bridgecodec.NewUnauthorizedResponse("Invalid bearer token"), nil
	}
	claims, err := s.parseDynamicRouteToken(ctx, tokenString)
	if err != nil {
		return nil, bridgecodec.NewUnauthorizedResponse(err.Error()), nil
	}
	exists, err := s.touchDynamicRouteSession(ctx, claims.TenantId, claims.TokenId)
	if err != nil {
		return nil, nil, err
	}
	if !exists {
		return nil, bridgecodec.NewUnauthorizedResponse("Session has expired"), nil
	}

	if s.userCtx != nil {
		s.userCtx.SetUser(ctx, claims.TokenId, claims.UserId, claims.Username, claims.Status)
		s.userCtx.SetTenant(ctx, claims.TenantId)
		if claims.ActingAsTenant || claims.IsImpersonation {
			if impersonationSetter, ok := s.userCtx.(userImpersonationSetter); ok {
				impersonationSetter.SetImpersonation(
					ctx,
					claims.ActingUserId,
					claims.TenantId,
					claims.ActingAsTenant,
					claims.IsImpersonation,
				)
			}
		}
	}
	accessContext, err := s.getDynamicRouteAccessContext(ctx, claims.UserId, claims.TenantId)
	if err != nil {
		return nil, nil, err
	}
	if match.Route.Permission != "" && !hasDynamicRoutePermission(accessContext, match.Route.Permission) {
		return nil, bridgecodec.NewForbiddenResponse("Permission denied"), nil
	}
	if s.userCtx != nil {
		s.userCtx.SetUserAccess(
			ctx,
			accessContext.DataScope,
			accessContext.DataScopeUnsupported,
			accessContext.UnsupportedDataScope,
		)
	}

	return &bridgecontract.IdentitySnapshotV1{
		TokenID:              claims.TokenId,
		TenantId:             int32(claims.TenantId),
		UserID:               int32(claims.UserId),
		Username:             claims.Username,
		Status:               int32(claims.Status),
		ActingUserId:         int32(claims.ActingUserId),
		ActingAsTenant:       claims.ActingAsTenant,
		IsImpersonation:      claims.IsImpersonation,
		Permissions:          append([]string(nil), accessContext.Permissions...),
		RoleNames:            append([]string(nil), accessContext.RoleNames...),
		DataScope:            int32(accessContext.DataScope),
		DataScopeUnsupported: accessContext.DataScopeUnsupported,
		UnsupportedDataScope: int32(accessContext.UnsupportedDataScope),
		IsSuperAdmin:         accessContext.IsSuperAdmin,
	}, nil, nil
}

// parseDynamicRouteToken validates the bearer token and extracts route claims.
func (s *serviceImpl) parseDynamicRouteToken(ctx context.Context, tokenString string) (*dynamicRouteClaims, error) {
	secret := ""
	if s.jwtConfig != nil {
		secret = s.jwtConfig.GetJwtSecret(ctx)
	}
	token, err := jwt.ParseWithClaims(tokenString, &dynamicRouteClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, gerror.New("invalid token")
	}
	claims, ok := token.Claims.(*dynamicRouteClaims)
	if !ok || !token.Valid {
		return nil, gerror.New("invalid token")
	}
	if claims.TokenType != "access" {
		return nil, gerror.New("invalid token")
	}
	return claims, nil
}

// touchDynamicRouteSession refreshes the last-active timestamp for one
// tenant/token session and tolerates second-level TIMESTAMP precision when no
// row is reported as updated.
func (s *serviceImpl) touchDynamicRouteSession(ctx context.Context, tenantID int, tokenID string) (bool, error) {
	result, err := dao.SysOnlineSession.Ctx(ctx).
		Where(do.SysOnlineSession{TenantId: tenantID, TokenId: tokenID}).
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

	// TIMESTAMP precision can still produce a zero affected-row update when a
	// dynamic route is hit without changing last_active_time, even though the
	// session still exists and remains valid.
	count, err := dao.SysOnlineSession.Ctx(ctx).
		Where(do.SysOnlineSession{TenantId: tenantID, TokenId: tokenID}).
		Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// getDynamicRouteAccessContext loads permissions and role names for one user ID
// within the tenant carried by the current dynamic-route token.
func (s *serviceImpl) getDynamicRouteAccessContext(
	ctx context.Context,
	userID int,
	tenantID int,
) (*dynamicRouteAccessContext, error) {
	roleIDs, err := s.getDynamicRouteUserRoleIDs(ctx, userID, tenantID)
	if err != nil {
		return nil, err
	}
	roles, err := s.getDynamicRouteRoles(ctx, roleIDs, tenantID)
	if err != nil {
		return nil, err
	}
	roleNames := dynamicRouteRoleNames(roles)
	dataScope, unsupported, unsupportedValue := dynamicRouteDataScope(roles)
	permissions, err := s.getDynamicRoutePermissionsByRoleIDs(ctx, roleIDs, tenantID)
	if err != nil {
		return nil, err
	}
	return &dynamicRouteAccessContext{
		Permissions:          permissions,
		RoleNames:            roleNames,
		DataScope:            dataScope,
		DataScopeUnsupported: unsupported,
		UnsupportedDataScope: unsupportedValue,
		IsSuperAdmin:         containsInt(roleIDs, 1),
	}, nil
}

// getDynamicRouteUserRoleIDs returns the deduplicated tenant-local role IDs
// assigned to the user.
func (s *serviceImpl) getDynamicRouteUserRoleIDs(ctx context.Context, userID int, tenantID int) ([]int, error) {
	items := make([]*entity.SysUserRole, 0)
	if err := dao.SysUserRole.Ctx(ctx).
		Where(do.SysUserRole{UserId: userID, TenantId: tenantID}).
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

// getDynamicRouteRoles loads active tenant-local roles for the given role IDs.
func (s *serviceImpl) getDynamicRouteRoles(ctx context.Context, roleIDs []int, tenantID int) ([]*entity.SysRole, error) {
	if len(roleIDs) == 0 {
		return []*entity.SysRole{}, nil
	}
	items := make([]*entity.SysRole, 0)
	if err := dao.SysRole.Ctx(ctx).
		WhereIn(dao.SysRole.Columns().Id, intsToInterfaces(roleIDs)).
		Where(do.SysRole{Status: statusNormal, TenantId: tenantID}).
		Scan(&items); err != nil {
		return nil, err
	}
	return items, nil
}

// dynamicRouteRoleNames projects active role rows into the identity snapshot.
func dynamicRouteRoleNames(roles []*entity.SysRole) []string {
	roleNames := make([]string, 0, len(roles))
	for _, item := range roles {
		if item == nil {
			continue
		}
		roleNames = append(roleNames, item.Name)
	}
	return roleNames
}

// dynamicRouteDataScope resolves one user's effective role data-scope from the
// role rows already used to build the dynamic-route identity snapshot.
func dynamicRouteDataScope(roles []*entity.SysRole) (int, bool, int) {
	scope := datascope.ScopeNone
	for _, item := range roles {
		if item == nil {
			continue
		}
		switch datascope.Scope(item.DataScope) {
		case datascope.ScopeAll:
			return int(datascope.ScopeAll), false, 0
		case datascope.ScopeTenant:
			if scope != datascope.ScopeAll {
				scope = datascope.ScopeTenant
			}
		case datascope.ScopeDept:
			if scope == datascope.ScopeNone || scope == datascope.ScopeSelf {
				scope = datascope.ScopeDept
			}
		case datascope.ScopeSelf:
			if scope == datascope.ScopeNone {
				scope = datascope.ScopeSelf
			}
		default:
			return int(datascope.ScopeNone), true, item.DataScope
		}
	}
	return int(scope), false, 0
}

// getDynamicRoutePermissionsByRoleIDs merges the role-menu and menu-permission
// lookups into a single pass: it fetches menu IDs bound to the given roles, then
// loads only button-type permission menus in one query (3 DB queries total for
// the full access context instead of 5).
func (s *serviceImpl) getDynamicRoutePermissionsByRoleIDs(
	ctx context.Context,
	roleIDs []int,
	tenantID int,
) ([]string, error) {
	if len(roleIDs) == 0 {
		return []string{}, nil
	}
	roleMenuItems := make([]*entity.SysRoleMenu, 0)
	if err := dao.SysRoleMenu.Ctx(ctx).
		WhereIn(dao.SysRoleMenu.Columns().RoleId, intsToInterfaces(roleIDs)).
		Where(do.SysRoleMenu{TenantId: tenantID}).
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

// hasDynamicRoutePermission reports whether the access context satisfies the
// route permission, with super-admin bypass support.
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

// containsInt reports whether target appears in the slice.
func containsInt(values []int, target int) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

// intsToInterfaces converts role or menu IDs into interface values for WhereIn.
func intsToInterfaces(values []int) []interface{} {
	items := make([]interface{}, 0, len(values))
	for _, value := range values {
		items = append(items, value)
	}
	return items
}

// sanitizeDynamicRouteHeaders clones request headers while stripping bearer tokens.
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

// collectRequestCookies snapshots request cookies into a simple name-value map.
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

// cloneURLValues deep-copies URL query values for bridge payload serialization.
func cloneURLValues(values url.Values) map[string][]string {
	result := make(map[string][]string, len(values))
	for key, items := range values {
		result[key] = append([]string(nil), items...)
	}
	return result
}

// cloneStringMap deep-copies string maps used in request and route snapshots.
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

// buildDynamicRequestID derives a stable host-side request ID for bridge logging.
func buildDynamicRequestID(match *dynamicRouteMatch, request *ghttp.Request) string {
	if request == nil {
		return match.PluginID + ":" + base64.StdEncoding.EncodeToString([]byte(match.InternalPath))
	}
	return match.PluginID + ":" + request.Method + ":" + match.InternalPath
}

// setDynamicRouteRuntimeState stores the resolved runtime state on the request context.
func setDynamicRouteRuntimeState(request *ghttp.Request, runtimeState *dynamicRouteRuntimeState) {
	if request == nil {
		return
	}
	request.SetCtxVar(dynamicRouteCtxVarState, runtimeState)
}

// getDynamicRouteRuntimeState reads the cached runtime state from the request context.
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

// setDynamicRouteIdentitySnapshot stores the resolved identity snapshot on the request.
func setDynamicRouteIdentitySnapshot(request *ghttp.Request, identity *bridgecontract.IdentitySnapshotV1) {
	if request == nil {
		return
	}
	request.SetCtxVar(dynamicRouteCtxVarIdentity, identity)
}

// getDynamicRouteIdentitySnapshot reads the cached identity snapshot from the request.
func getDynamicRouteIdentitySnapshot(request *ghttp.Request) *bridgecontract.IdentitySnapshotV1 {
	if request == nil {
		return nil
	}
	value := request.GetCtxVar(dynamicRouteCtxVarIdentity).Val()
	if value == nil {
		return nil
	}
	identity, _ := value.(*bridgecontract.IdentitySnapshotV1)
	return identity
}

// setDynamicRouteMetadata stores generic dynamic-route metadata on the request context.
func setDynamicRouteMetadata(request *ghttp.Request, metadata *DynamicRouteMetadata) {
	if request == nil || metadata == nil {
		return
	}
	request.SetCtxVar(dynamicRouteCtxVarMetadata, metadata)
}

// buildDynamicRouteMetadata maps matched route declarations into generic
// request metadata for source-plugin middleware.
func buildDynamicRouteMetadata(runtimeState *dynamicRouteRuntimeState) *DynamicRouteMetadata {
	if runtimeState == nil || runtimeState.Match == nil || runtimeState.Match.Route == nil {
		return nil
	}
	metadata := &DynamicRouteMetadata{
		PluginID:   strings.TrimSpace(runtimeState.Match.PluginID),
		Method:     strings.TrimSpace(runtimeState.Match.Route.Method),
		PublicPath: strings.TrimSpace(runtimeState.Match.PublicPath),
		Tags:       append([]string(nil), runtimeState.Match.Route.Tags...),
		Summary:    strings.TrimSpace(runtimeState.Match.Route.Summary),
		Meta:       cloneStringMap(runtimeState.Match.Route.Meta),
	}
	return metadata
}

// GetDynamicRouteMetadata returns generic dynamic-route metadata attached
// during the host middleware chain.
func GetDynamicRouteMetadata(request *ghttp.Request) *DynamicRouteMetadata {
	if request == nil {
		return nil
	}
	value := request.GetCtxVar(dynamicRouteCtxVarMetadata).Val()
	if value == nil {
		return nil
	}
	metadata, _ := value.(*DynamicRouteMetadata)
	return metadata
}

// writeDynamicRouteResponse writes the guest response back without going through
// GoFrame's default success wrapper, otherwise raw plugin payloads would be
// polluted by host-managed response formatting.
func (s *serviceImpl) writeDynamicRouteResponse(request *ghttp.Request, response *bridgecontract.BridgeResponseEnvelopeV1) {
	if request == nil || response == nil {
		return
	}
	metadata := GetDynamicRouteMetadata(request)
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
