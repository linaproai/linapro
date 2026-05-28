// This file adapts the shared JWT client-type contract to auth service errors.

package auth

import (
	"lina-core/pkg/authtoken"
	"lina-core/pkg/bizerr"
)

// ClientType identifies the user-facing client that created an authenticated user session.
type ClientType = authtoken.ClientType

const (
	// ClientTypeWeb represents browser Web apps and the default LinaPro workbench.
	ClientTypeWeb = authtoken.ClientTypeWeb
	// ClientTypeMobile represents mobile app user sessions.
	ClientTypeMobile = authtoken.ClientTypeMobile
	// ClientTypeDesktop represents desktop client user sessions.
	ClientTypeDesktop = authtoken.ClientTypeDesktop
	// ClientTypeCLI represents user sessions created through command-line clients.
	ClientTypeCLI = authtoken.ClientTypeCLI
)

// ParseClientType validates one user-session client type value.
func ParseClientType(value string) (ClientType, error) {
	if clientType, ok := authtoken.ParseClientType(value); ok {
		return clientType, nil
	}
	return "", bizerr.NewCode(CodeAuthClientTypeInvalid)
}
