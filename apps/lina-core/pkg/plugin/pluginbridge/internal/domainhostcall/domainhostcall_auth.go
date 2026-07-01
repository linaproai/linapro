// This file implements guest-side authentication token hostcall clients and
// the authentication namespace that links token and authz sub-capabilities.

package domainhostcall

import (
	"context"

	"lina-core/pkg/plugin/capability/authcap"
	"lina-core/pkg/plugin/capability/authcap/authz"
	"lina-core/pkg/plugin/capability/authcap/token"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// authService exposes authentication sub-capabilities through host services.
type authService struct{ baseService }

// authTokenService adapts token handoff calls to host services.
type authTokenService struct{ baseService }

// authzService adapts authorization reads to auth-domain host services.
type authzService struct{ baseService }

// Auth creates the authentication and authorization guest namespace.
func Auth(invoker Invoker) authcap.Service {
	return authService{baseService: newBaseService(invoker)}
}

// Token returns tenant-token handoff operations.
func (s authService) Token() token.Service {
	return authTokenService{baseService: s.baseService}
}

// Authz returns authorization-domain ordinary operations.
func (s authService) Authz() authz.Service {
	return authzService{baseService: s.baseService}
}

// SelectTenant consumes a pre-login token and issues a tenant-bound token.
func (s authTokenService) SelectTenant(_ context.Context, in token.SelectTenantInput) (*token.TenantTokenOutput, error) {
	out := &token.TenantTokenOutput{}
	err := s.callJSONRequest(protocol.HostServiceAuth, protocol.HostServiceMethodAuthSelectTenant, in, out)
	return out, err
}

// SwitchTenant validates membership, revokes the current token, and issues a new token.
func (s authTokenService) SwitchTenant(_ context.Context, in token.SwitchTenantInput) (*token.TenantTokenOutput, error) {
	out := &token.TenantTokenOutput{}
	err := s.callJSONRequest(protocol.HostServiceAuth, protocol.HostServiceMethodAuthSwitchTenant, in, out)
	return out, err
}

// IssueImpersonationToken asks the host to issue one impersonation token.
func (s authTokenService) IssueImpersonationToken(_ context.Context, in token.ImpersonationTokenIssueInput) (*token.ImpersonationTokenOutput, error) {
	out := &token.ImpersonationTokenOutput{}
	err := s.callJSONRequest(protocol.HostServiceAuth, protocol.HostServiceMethodAuthIssueImpersonationToken, in, out)
	return out, err
}

// RevokeImpersonationToken asks the host to revoke one impersonation token.
func (s authTokenService) RevokeImpersonationToken(_ context.Context, in token.ImpersonationTokenRevokeInput) error {
	return s.callJSONRequest(protocol.HostServiceAuth, protocol.HostServiceMethodAuthRevokeImpersonationToken, in, nil)
}

// BatchGetPermissions returns visible permission projections and opaque missing keys.
func (s authzService) BatchGetPermissions(_ context.Context, keys []authz.PermissionKey) (*capmodel.BatchResult[*authz.PermissionInfo, authz.PermissionKey], error) {
	out := &capmodel.BatchResult[*authz.PermissionInfo, authz.PermissionKey]{Items: map[authz.PermissionKey]*authz.PermissionInfo{}}
	err := s.callJSONRequest(protocol.HostServiceAuth, protocol.HostServiceMethodAuthzBatchGetPermissions, idsRequest{IDs: permissionKeysToStrings(keys)}, out)
	return out, err
}

// BatchHasPermissions reports whether the actor has each permission in the current scope.
func (s authzService) BatchHasPermissions(_ context.Context, keys []authz.PermissionKey) (map[authz.PermissionKey]bool, error) {
	out := map[authz.PermissionKey]bool{}
	err := s.callJSONRequest(protocol.HostServiceAuth, protocol.HostServiceMethodAuthzBatchHasPermissions, idsRequest{IDs: permissionKeysToStrings(keys)}, &out)
	return out, err
}

// HasPermission reports whether the actor has one permission in the current scope.
func (s authzService) HasPermission(_ context.Context, key authz.PermissionKey) (bool, error) {
	var out bool
	err := s.callJSONRequest(protocol.HostServiceAuth, protocol.HostServiceMethodAuthzHasPermission, keyRequest{Key: string(key)}, &out)
	return out, err
}

// IsPlatformAdmin reports whether the user has a platform all-data role.
func (s authzService) IsPlatformAdmin(_ context.Context, userID authz.UserID) (bool, error) {
	var out bool
	err := s.callJSONRequest(protocol.HostServiceAuth, protocol.HostServiceMethodAuthzIsPlatformAdmin, userIDRequest{UserID: string(userID)}, &out)
	return out, err
}

// ReplaceRolePermissions replaces one role's visible permission assignments.
func (s authzService) ReplaceRolePermissions(_ context.Context, roleID authz.RoleID, keys []authz.PermissionKey) error {
	return s.callJSONRequest(protocol.HostServiceAuth, protocol.HostServiceMethodAuthzReplaceRolePermissions, authzReplaceRolePermissionsRequest{
		RoleID: string(roleID),
		Keys:   permissionKeysToStrings(keys),
	}, nil)
}

// keyRequest carries one string key for JSON capability methods.
type keyRequest struct {
	Key string `json:"key"`
}

// userIDRequest carries one user identifier for JSON capability methods.
type userIDRequest struct {
	UserID string `json:"userId"`
}

type authzReplaceRolePermissionsRequest struct {
	RoleID string   `json:"roleId"`
	Keys   []string `json:"keys"`
}

// permissionKeysToStrings converts permission keys to transport strings.
func permissionKeysToStrings(ids []authz.PermissionKey) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, string(id))
	}
	return out
}

var (
	_ authcap.Service = (*authService)(nil)
	_ token.Service   = (*authTokenService)(nil)
	_ authz.Service   = (*authzService)(nil)
)
