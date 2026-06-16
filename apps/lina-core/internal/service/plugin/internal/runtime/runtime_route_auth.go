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

	rolesvc "lina-core/internal/service/role"
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
	accessContext, err := s.buildDynamicRouteAccessProjection(ctx, claims)
	if err != nil {
		return nil, nil, err
	}
	if match.Route.Permission != "" && !hasDynamicRoutePermission(accessContext, match.Route.Permission) {
		return nil, bridgecodec.NewForbiddenResponse("Permission denied"), nil
	}
	if s.userCtx != nil {
		s.userCtx.SetUserAccess(
			ctx,
			int(accessContext.DataScope),
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

// buildDynamicRouteAccessProjection delegates permission and data-scope
// projection to the role module that owns token access snapshots.
func (s *serviceImpl) buildDynamicRouteAccessProjection(
	ctx context.Context,
	claims *dynamicRouteClaims,
) (*rolesvc.DynamicRouteAccessProjection, error) {
	if s == nil || s.roleAccess == nil {
		return nil, gerror.New("dynamic route role access projector is not configured")
	}
	if claims == nil {
		return nil, gerror.New("dynamic route token claims are missing")
	}
	return s.roleAccess.BuildDynamicRouteAccessProjection(
		ctx,
		claims.TokenId,
		claims.UserId,
		claims.TenantId,
	)
}

// hasDynamicRoutePermission reports whether the access context satisfies the
// route permission, with super-admin bypass support.
func hasDynamicRoutePermission(accessContext *rolesvc.DynamicRouteAccessProjection, permission string) bool {
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
