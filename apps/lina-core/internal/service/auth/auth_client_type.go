// This file adapts the shared JWT client-type contract to auth service errors.

package auth

import (
	"lina-core/pkg/bizerr"
	tokencap "lina-core/pkg/plugin/capability/authcap/token"
)

// ClientType identifies the user-facing client that created an authenticated user session.
type ClientType = tokencap.ClientType

// ParseClientType validates one user-session client type value.
func ParseClientType(value string) (ClientType, error) {
	if clientType, ok := tokencap.ParseClientType(value); ok {
		return clientType, nil
	}
	return "", bizerr.NewCode(CodeAuthClientTypeInvalid)
}
