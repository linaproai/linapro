// Package authtoken declares JWT claim contracts shared by the host auth
// service, host-side dynamic route parsing, and source plugins that need to sign
// or validate host-compatible JWTs. Keeping these literals in one place
// prevents the host signer and plugin/runtime validators from drifting apart.
package authtoken

import "strings"

// Kind names the intended use of one signed JWT carried in the `tokenType`
// claim. Use these constants instead of hard-coding the string literals.
const (
	// KindAccess marks JWTs accepted by protected API middleware and
	// host-side dynamic plugin route dispatch.
	KindAccess = "access"
	// KindRefresh marks JWTs accepted only by the refresh-token endpoint.
	KindRefresh = "refresh"
)

// ClientType identifies the user-facing client that created an authenticated
// user session. It deliberately excludes service, plugin, and machine actors.
type ClientType string

const (
	// ClientTypeWeb represents browser Web apps and the default LinaPro workbench.
	ClientTypeWeb ClientType = "web"
	// ClientTypeMobile represents mobile app user sessions.
	ClientTypeMobile ClientType = "mobile"
	// ClientTypeDesktop represents desktop client user sessions.
	ClientTypeDesktop ClientType = "desktop"
	// ClientTypeCLI represents user sessions created through command-line clients.
	ClientTypeCLI ClientType = "cli"
)

// ParseClientType validates one user-session client type value.
func ParseClientType(value string) (ClientType, bool) {
	switch clientType := ClientType(strings.TrimSpace(value)); clientType {
	case ClientTypeWeb, ClientTypeMobile, ClientTypeDesktop, ClientTypeCLI:
		return clientType, true
	default:
		return "", false
	}
}

// String returns the wire value for one client type.
func (t ClientType) String() string {
	return string(t)
}
