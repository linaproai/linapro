// Package authtoken declares the JWT `tokenType` claim contract shared by the
// host auth service, host-side dynamic route parsing, and source plugins that
// need to sign or validate host-compatible JWTs (for example multi-tenant
// impersonation). Keeping these literals in one place prevents the host signer
// and the plugin/runtime validators from drifting apart.
package authtoken

// Kind names the intended use of one signed JWT carried in the `tokenType`
// claim. Use these constants instead of hard-coding the string literals.
const (
	// KindAccess marks JWTs accepted by protected API middleware and
	// host-side dynamic plugin route dispatch.
	KindAccess = "access"
	// KindRefresh marks JWTs accepted only by the refresh-token endpoint.
	KindRefresh = "refresh"
)
