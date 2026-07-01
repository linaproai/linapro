// This file implements guest-side authorization capability hostcall clients.
// Permission and platform-admin request DTOs stay local to the authz domain.

package domainhostcall

import (
	"context"

	"lina-core/pkg/plugin/capability/authcap/authz"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// authzService adapts authorization reads to host services.
type authzService struct{ baseService }

// Authz creates the authorization-domain guest client.
func Authz(invoker Invoker) authz.Service {
	return authzService{baseService: newBaseService(invoker)}
}

// BatchGetPermissions returns visible permission projections and opaque missing keys.
func (s authzService) BatchGetPermissions(_ context.Context, keys []authz.PermissionKey) (*capmodel.BatchResult[*authz.PermissionInfo, authz.PermissionKey], error) {
	out := &capmodel.BatchResult[*authz.PermissionInfo, authz.PermissionKey]{Items: map[authz.PermissionKey]*authz.PermissionInfo{}}
	err := s.callJSONRequest(protocol.HostServiceAuthz, protocol.HostServiceMethodAuthzBatchGetPermissions, idsRequest{IDs: permissionKeysToStrings(keys)}, out)
	return out, err
}

// BatchHasPermissions reports whether the actor has each permission in the current scope.
func (s authzService) BatchHasPermissions(_ context.Context, keys []authz.PermissionKey) (map[authz.PermissionKey]bool, error) {
	out := map[authz.PermissionKey]bool{}
	err := s.callJSONRequest(protocol.HostServiceAuthz, protocol.HostServiceMethodAuthzBatchHasPermissions, idsRequest{IDs: permissionKeysToStrings(keys)}, &out)
	return out, err
}

// HasPermission reports whether the actor has one permission in the current scope.
func (s authzService) HasPermission(_ context.Context, key authz.PermissionKey) (bool, error) {
	var out bool
	err := s.callJSONRequest(protocol.HostServiceAuthz, protocol.HostServiceMethodAuthzHasPermission, keyRequest{Key: string(key)}, &out)
	return out, err
}

// IsPlatformAdmin reports whether the user has a platform all-data role.
func (s authzService) IsPlatformAdmin(_ context.Context, userID authz.UserID) (bool, error) {
	var out bool
	err := s.callJSONRequest(protocol.HostServiceAuthz, protocol.HostServiceMethodAuthzIsPlatformAdmin, userIDRequest{UserID: string(userID)}, &out)
	return out, err
}

// ReplaceRolePermissions is not published as a dynamic authz host-service method.
func (s authzService) ReplaceRolePermissions(context.Context, authz.RoleID, []authz.PermissionKey) error {
	return unsupportedDynamicMethodError("authz.role_permissions.replace")
}

// keyRequest carries one string key for JSON capability methods.
type keyRequest struct {
	Key string `json:"key"`
}

// userIDRequest carries one user identifier for JSON capability methods.
type userIDRequest struct {
	UserID string `json:"userId"`
}

// permissionKeysToStrings converts permission keys to transport strings.
func permissionKeysToStrings(ids []authz.PermissionKey) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, string(id))
	}
	return out
}

var _ authz.Service = (*authzService)(nil)
