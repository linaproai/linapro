// This file defines client direct object-access types for storage providers.
// Direct access lets browsers and other clients transfer object bytes without
// proxying content through the host process, while keys and security scopes
// remain host-owned.

package storagecap

import (
	"context"
	"strings"
	"time"
)

// Direct access mode identifiers returned to generic callers.
const (
	// DirectAccessModePresignedURL is a single HTTP request to a signed URL.
	DirectAccessModePresignedURL = "presigned_url"
	// DirectAccessModeFormPost is a browser form POST with signed form fields.
	DirectAccessModeFormPost = "form_post"
	// DirectAccessModeTemporaryCredentials is short-lived vendor credentials.
	DirectAccessModeTemporaryCredentials = "temporary_credentials"
	// DirectAccessModeProxy indicates the client must use host-mediated transfer.
	DirectAccessModeProxy = "proxy"
)

// DirectAccessOperation identifies the client transfer intent.
type DirectAccessOperation string

const (
	// DirectAccessOpPut creates or replaces one object via the client.
	DirectAccessOpPut DirectAccessOperation = "put"
	// DirectAccessOpGet reads one object via the client.
	DirectAccessOpGet DirectAccessOperation = "get"
)

// DirectAccessProvider is optionally implemented by object backends that can
// issue short-lived client transfer credentials or URLs for scoped keys.
type DirectAccessProvider interface {
	// SupportsDirectAccess reports whether the backend can issue client access
	// for the given operation under the current configuration.
	SupportsDirectAccess(ctx context.Context, op DirectAccessOperation) bool
	// CreateDirectAccess issues one short-lived client access description for a
	// host-assigned scoped object key. Callers must never forward permanent
	// credentials to untrusted clients.
	CreateDirectAccess(ctx context.Context, in ProviderDirectAccessInput) (*DirectAccess, error)
}

// ProviderDirectAccessInput defines one provider-level direct access request.
// Key is the already-scoped provider object key.
type ProviderDirectAccessInput struct {
	// Key is the scoped provider object key.
	Key string
	// Operation is put or get.
	Operation DirectAccessOperation
	// Size is the expected object size for put when known. Negative means unknown.
	Size int64
	// ContentType is the optional MIME type constraint for put.
	ContentType string
	// TTL bounds how long the issued access remains valid. Zero uses provider default.
	TTL time.Duration
	// Overwrite controls whether put may replace an existing object when the
	// provider can encode that constraint.
	Overwrite bool
}

// DirectAccess is a vendor-neutral client transfer description. Callers branch
// on Mode, never on cloud provider IDs.
type DirectAccess struct {
	// Mode is one of the DirectAccessMode* constants.
	Mode string
	// Operation is put or get.
	Operation DirectAccessOperation
	// Method is the HTTP method for presigned_url mode (for example PUT or GET).
	Method string
	// URL is the target endpoint for presigned_url or form_post modes.
	URL string
	// Headers are required request headers for presigned_url mode.
	Headers map[string]string
	// FormFields are required form fields for form_post mode.
	FormFields map[string]string
	// AccessKeyID is temporary credential material when Mode is temporary_credentials.
	AccessKeyID string
	// SecretAccessKey is temporary credential material when Mode is temporary_credentials.
	SecretAccessKey string
	// SessionToken is temporary credential material when Mode is temporary_credentials.
	SessionToken string
	// ExpiresAt is the absolute expiry of the issued access.
	ExpiresAt time.Time
	// ProviderID identifies the backend that issued the access for diagnostics.
	// Callers must not branch business logic on this field.
	ProviderID string
}

// AsDirectAccessProvider returns the optional direct-access capability.
func AsDirectAccessProvider(provider Provider) (DirectAccessProvider, bool) {
	if provider == nil {
		return nil, false
	}
	direct, ok := provider.(DirectAccessProvider)
	return direct, ok
}

// SupportsDirectAccess reports whether provider can issue client access for op.
// Providers that do not implement DirectAccessProvider return false.
func SupportsDirectAccess(ctx context.Context, provider Provider, op DirectAccessOperation) bool {
	direct, ok := AsDirectAccessProvider(provider)
	if !ok || direct == nil {
		return false
	}
	return direct.SupportsDirectAccess(ctx, op)
}

// CreateDirectAccess issues client access when supported; otherwise returns a
// proxy-mode DirectAccess without error so callers can fall back uniformly.
// Resolve/provider unavailability errors from the caller still apply before
// invoking this helper.
func CreateDirectAccess(
	ctx context.Context,
	providerID string,
	provider Provider,
	in ProviderDirectAccessInput,
) (*DirectAccess, error) {
	op := NormalizeDirectAccessOperation(in.Operation)
	if op == "" {
		return nil, NewInvalidDirectAccessOperationError()
	}
	in.Operation = op
	if !SupportsDirectAccess(ctx, provider, op) {
		return &DirectAccess{
			Mode:       DirectAccessModeProxy,
			Operation:  op,
			ProviderID: strings.TrimSpace(providerID),
		}, nil
	}
	direct, ok := AsDirectAccessProvider(provider)
	if !ok || direct == nil {
		return &DirectAccess{
			Mode:       DirectAccessModeProxy,
			Operation:  op,
			ProviderID: strings.TrimSpace(providerID),
		}, nil
	}
	access, err := direct.CreateDirectAccess(ctx, in)
	if err != nil {
		return nil, err
	}
	if access == nil {
		return nil, NewDirectAccessIssueFailedError()
	}
	if strings.TrimSpace(access.Mode) == "" {
		access.Mode = DirectAccessModePresignedURL
	}
	if access.Operation == "" {
		access.Operation = op
	}
	if strings.TrimSpace(access.ProviderID) == "" {
		access.ProviderID = strings.TrimSpace(providerID)
	}
	if strings.EqualFold(strings.TrimSpace(access.Mode), DirectAccessModeProxy) {
		return access, nil
	}
	return access, nil
}

// NormalizeDirectAccessOperation canonicalizes operation strings.
func NormalizeDirectAccessOperation(op DirectAccessOperation) DirectAccessOperation {
	switch DirectAccessOperation(strings.ToLower(strings.TrimSpace(string(op)))) {
	case DirectAccessOpPut:
		return DirectAccessOpPut
	case DirectAccessOpGet:
		return DirectAccessOpGet
	default:
		return ""
	}
}

// IsProxyDirectAccess reports whether access requires host-mediated transfer.
func IsProxyDirectAccess(access *DirectAccess) bool {
	if access == nil {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(access.Mode), DirectAccessModeProxy)
}
