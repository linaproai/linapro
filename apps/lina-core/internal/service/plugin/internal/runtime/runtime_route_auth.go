// This file owns host-side authentication and permission checks for dynamic
// plugin routes before requests enter guest code.

package runtime

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/golang-jwt/jwt/v5"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/datascope"
	"lina-core/internal/service/plugin/internal/catalog"
	tokencap "lina-core/pkg/plugin/capability/authcap/token"
	bridgecontract "lina-core/pkg/plugin/pluginbridge/contract"
	bridgecodec "lina-core/pkg/plugin/pluginbridge/protocol"
)

// dynamicRouteClaims mirrors the JWT claims needed by host-side dynamic route auth.
type dynamicRouteClaims struct {
	TokenId         string `json:"tokenId"`
	TokenType       string `json:"tokenType"`
	ClientType      string `json:"clientType"`
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
		s.userCtx.SetUser(ctx, claims.TokenId, claims.UserId, claims.Username, claims.Status, claims.ClientType)
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
	if claims.TokenType != tokencap.KindAccess {
		return nil, gerror.New("invalid token")
	}
	if _, ok := tokencap.ParseClientType(claims.ClientType); !ok {
		return nil, gerror.New("invalid token")
	}
	return claims, nil
}

// touchDynamicRouteSession refreshes the last-active timestamp for one
// tenant/token session and tolerates second-level TIMESTAMP precision when no
// row is reported as updated.
func (s *serviceImpl) touchDynamicRouteSession(ctx context.Context, tenantID int, tokenID string) (bool, error) {
	if s == nil || s.sessionStore == nil {
		return false, nil
	}
	timeout := 24 * time.Hour
	if s.jwtConfig != nil {
		configTimeout, err := s.jwtConfig.GetSessionTimeout(ctx)
		if err != nil {
			return false, err
		}
		if configTimeout > 0 {
			timeout = configTimeout
		}
	}
	return s.sessionStore.TouchOrValidate(ctx, tenantID, tokenID, timeout)
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
