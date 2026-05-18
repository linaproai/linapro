// Package contract defines stable bridge ABI contracts and validation rules
// shared by Lina dynamic plugin hosts, builders, and guests.
package contract

// Bridge ABI constants define the stable runtime defaults shared between host,
// builder, and guest implementations.
const (
	// CodecProtobuf is the only supported executable bridge envelope codec.
	CodecProtobuf = "protobuf"

	// AccessPublic allows anonymous access.
	AccessPublic = "public"
	// AccessLogin requires authenticated access.
	AccessLogin = "login"

	// RuntimeKindWasm identifies a wasm runtime artifact.
	RuntimeKindWasm = "wasm"
	// ABIVersionV1 is the current bridge ABI version.
	ABIVersionV1 = "v1"
	// SupportedABIVersion is the current runtime artifact ABI version.
	SupportedABIVersion = ABIVersionV1

	// DefaultGuestAllocExport is the default guest allocator export.
	DefaultGuestAllocExport = "lina_dynamic_route_alloc"
	// DefaultGuestExecuteExport is the default guest executor export.
	DefaultGuestExecuteExport = "lina_dynamic_route_execute"
)

// Bridge failure codes normalize guest execution failures into stable
// machine-readable categories.
const (
	// BridgeFailureCodeUnauthorized identifies unauthenticated bridge failures.
	BridgeFailureCodeUnauthorized = "UNAUTHORIZED"
	// BridgeFailureCodeForbidden identifies permission-denied bridge failures.
	BridgeFailureCodeForbidden = "FORBIDDEN"
	// BridgeFailureCodeBadRequest identifies malformed bridge request failures.
	BridgeFailureCodeBadRequest = "BAD_REQUEST"
	// BridgeFailureCodeNotFound identifies missing dynamic route failures.
	BridgeFailureCodeNotFound = "NOT_FOUND"
	// BridgeFailureCodeInternal identifies internal bridge execution failures.
	BridgeFailureCodeInternal = "INTERNAL_ERROR"
)

// RouteContract describes one dynamic plugin route contract embedded into the artifact.
type RouteContract struct {
	// Path is the public plugin route path exposed by the host router.
	Path string `json:"path" yaml:"path"`
	// Method is the normalized HTTP method accepted by the route.
	Method string `json:"method" yaml:"method"`
	// Tags lists semantic route tags used for grouping and governance metadata.
	Tags []string `json:"tags,omitempty" yaml:"tags,omitempty"`
	// Summary is the short human-readable route summary shown in manifests or docs.
	Summary string `json:"summary,omitempty" yaml:"summary,omitempty"`
	// Description is the detailed business description for the route contract.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// Access declares whether the route is public or requires login context.
	Access string `json:"access,omitempty" yaml:"access,omitempty"`
	// Permission stores the host permission key enforced for authenticated access.
	Permission string `json:"permission,omitempty" yaml:"permission,omitempty"`
	// Meta stores plugin-defined route metadata that the host transports without interpretation.
	Meta map[string]string `json:"meta,omitempty" yaml:"meta,omitempty"`
	// RequestType is the guest controller request binding name resolved at runtime.
	RequestType string `json:"requestType,omitempty" yaml:"requestType,omitempty"`
}

// BridgeSpec defines the stable guest ABI contract embedded into the artifact.
type BridgeSpec struct {
	// ABIVersion identifies the stable bridge ABI version implemented by the guest.
	ABIVersion string `json:"abiVersion" yaml:"abiVersion"`
	// RuntimeKind identifies the runtime family expected by the bridge host.
	RuntimeKind string `json:"runtimeKind" yaml:"runtimeKind"`
	// RouteExecution reports whether the guest exposes executable route handlers.
	RouteExecution bool `json:"routeExecution" yaml:"routeExecution"`
	// RequestCodec is the envelope codec used for host-to-guest requests.
	RequestCodec string `json:"requestCodec,omitempty" yaml:"requestCodec,omitempty"`
	// ResponseCodec is the envelope codec used for guest-to-host responses.
	ResponseCodec string `json:"responseCodec,omitempty" yaml:"responseCodec,omitempty"`
	// AllocExport is the guest export name used to allocate request memory.
	AllocExport string `json:"allocExport,omitempty" yaml:"allocExport,omitempty"`
	// ExecuteExport is the guest export name used to execute one encoded request.
	ExecuteExport string `json:"executeExport,omitempty" yaml:"executeExport,omitempty"`
}

// BridgeRequestEnvelopeV1 is the host-to-guest request envelope.
type BridgeRequestEnvelopeV1 struct {
	// PluginID identifies the plugin that should execute the request.
	PluginID string `json:"pluginId"`
	// Route carries the matched route metadata resolved by the host router.
	Route *RouteMatchSnapshotV1 `json:"route,omitempty"`
	// Request carries the sanitized inbound HTTP request snapshot.
	Request *HTTPRequestSnapshotV1 `json:"request,omitempty"`
	// Identity carries the authenticated user context injected by the host.
	Identity *IdentitySnapshotV1 `json:"identity,omitempty"`
	// RequestID carries the host-generated trace identifier for this invocation.
	RequestID string `json:"requestId,omitempty"`
}

// RouteMatchSnapshotV1 describes the matched route and host path mapping.
type RouteMatchSnapshotV1 struct {
	// Method is the normalized HTTP method accepted by the matched route.
	Method string `json:"method,omitempty"`
	// PublicPath is the externally visible route path served by the host.
	PublicPath string `json:"publicPath,omitempty"`
	// InternalPath is the guest-side internal route path used for handler lookup.
	InternalPath string `json:"internalPath,omitempty"`
	// RoutePath is the original contract path declared by the plugin route.
	RoutePath string `json:"routePath,omitempty"`
	// Access records the resolved route access mode.
	Access string `json:"access,omitempty"`
	// Permission records the permission key enforced by the host for this route.
	Permission string `json:"permission,omitempty"`
	// RequestType is the guest controller binding name used for reflected dispatch.
	RequestType string `json:"requestType,omitempty"`
	// PathParams carries host-extracted path parameters from the matched request.
	PathParams map[string]string `json:"pathParams,omitempty"`
	// QueryValues carries the decoded query string values grouped by key.
	QueryValues map[string][]string `json:"queryValues,omitempty"`
}

// HTTPRequestSnapshotV1 describes the sanitized inbound HTTP request.
type HTTPRequestSnapshotV1 struct {
	// Method is the inbound HTTP method observed by the host.
	Method string `json:"method,omitempty"`
	// PublicPath is the routed public path seen by external callers.
	PublicPath string `json:"publicPath,omitempty"`
	// InternalPath is the host-mapped guest path used for route execution.
	InternalPath string `json:"internalPath,omitempty"`
	// RawPath is the raw request URL path before query string processing.
	RawPath string `json:"rawPath,omitempty"`
	// RawQuery stores the original encoded query string without the leading question mark.
	RawQuery string `json:"rawQuery,omitempty"`
	// Host stores the inbound HTTP host header value.
	Host string `json:"host,omitempty"`
	// Scheme stores the normalized request scheme such as http or https.
	Scheme string `json:"scheme,omitempty"`
	// RemoteAddr stores the direct peer network address observed by the host.
	RemoteAddr string `json:"remoteAddr,omitempty"`
	// ClientIP stores the resolved client IP after trusted proxy processing.
	ClientIP string `json:"clientIp,omitempty"`
	// ContentType stores the normalized request content type header.
	ContentType string `json:"contentType,omitempty"`
	// Headers carries the sanitized inbound request headers grouped by name.
	Headers map[string][]string `json:"headers,omitempty"`
	// Cookies carries the inbound request cookies grouped by cookie name.
	Cookies map[string]string `json:"cookies,omitempty"`
	// Body stores the raw request body bytes forwarded to the guest.
	Body []byte `json:"body,omitempty"`
}

// IdentitySnapshotV1 describes authenticated user context injected by the host.
type IdentitySnapshotV1 struct {
	// TokenID identifies the authenticated session or token presented to the host.
	TokenID string `json:"tokenId,omitempty"`
	// TenantId is the current tenant boundary associated with the token.
	TenantId int32 `json:"tenantId,omitempty"`
	// UserID is the host user identifier associated with the authenticated caller.
	UserID int32 `json:"userId,omitempty"`
	// Username is the normalized login name of the authenticated caller.
	Username string `json:"username,omitempty"`
	// Status is the host-defined user status code forwarded to the guest.
	Status int32 `json:"status,omitempty"`
	// ActingUserId is the real platform user ID when impersonation is active.
	ActingUserId int32 `json:"actingUserId,omitempty"`
	// ActingAsTenant reports whether the caller is operating through a tenant view.
	ActingAsTenant bool `json:"actingAsTenant,omitempty"`
	// IsImpersonation reports whether the token represents an impersonated context.
	IsImpersonation bool `json:"isImpersonation,omitempty"`
	// Permissions lists the permission keys granted to the authenticated caller.
	Permissions []string `json:"permissions,omitempty"`
	// RoleNames lists the resolved role names bound to the authenticated caller.
	RoleNames []string `json:"roleNames,omitempty"`
	// DataScope stores the effective role data-scope snapshot for host data operations.
	DataScope int32 `json:"dataScope,omitempty"`
	// DataScopeUnsupported reports whether the user's roles contain an unsupported data-scope value.
	DataScopeUnsupported bool `json:"dataScopeUnsupported,omitempty"`
	// UnsupportedDataScope stores the first unsupported data-scope value when DataScopeUnsupported is true.
	UnsupportedDataScope int32 `json:"unsupportedDataScope,omitempty"`
	// IsSuperAdmin reports whether the caller bypasses normal permission checks.
	IsSuperAdmin bool `json:"isSuperAdmin,omitempty"`
}

// BridgeResponseEnvelopeV1 is the guest-to-host response envelope.
type BridgeResponseEnvelopeV1 struct {
	// StatusCode is the HTTP status code returned by the guest handler.
	StatusCode int32 `json:"statusCode,omitempty"`
	// ContentType is the response content type returned to the client.
	ContentType string `json:"contentType,omitempty"`
	// Headers carries additional response headers grouped by header name.
	Headers map[string][]string `json:"headers,omitempty"`
	// Body stores the raw response payload emitted by the guest handler.
	Body []byte `json:"body,omitempty"`
	// Failure carries normalized failure metadata for non-successful execution.
	Failure *BridgeFailureV1 `json:"failure,omitempty"`
}

// BridgeFailureV1 contains normalized execution failure metadata.
type BridgeFailureV1 struct {
	// Code is the stable machine-readable failure code.
	Code string `json:"code,omitempty"`
	// Message is the human-readable failure message returned to callers.
	Message string `json:"message,omitempty"`
}
