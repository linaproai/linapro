// Package externallogin defines the external-identity login capability
// published through authcap. A source plugin that has already verified a
// third-party identity (for example a Google or Discord OIDC subject) uses this
// contract to exchange the verified identity for a host session, without
// gaining access to host JWT signing, session storage, user provisioning, or
// tenant membership internals.
//
// Implementations are plugin-scoped by the host. The calling plugin identity
// and the ownership of the requested provider are enforced by the host adapter
// from the source-plugin-scoped capability directory, never trusted from the
// plugin-supplied input. Provisioning stays host-owned and closed by default:
// an external identity with no linked local account is rejected without
// creating a user.
package externallogin

import "context"

// Service defines the external-identity login operation published to source
// plugins.
type Service interface {
	// LoginByVerifiedIdentity exchanges a plugin-verified external identity for
	// a host session. On success it returns a token pair for users with zero or
	// one active tenant, or a pre-login token plus tenant candidates for
	// multi-tenant users. The caller must have already completed provider
	// verification; the host performs no OAuth or token exchange. Requests for a
	// provider the calling plugin does not own, or from a disabled plugin, are
	// rejected. An unlinked identity is rejected without provisioning.
	LoginByVerifiedIdentity(ctx context.Context, in LoginInput) (*LoginOutput, error)
}

// LoginInput defines a plugin-verified external identity to exchange for a host
// session.
type LoginInput struct {
	// Provider is the stable external provider ID that must be owned by the
	// calling plugin. Requests for an unowned provider are rejected.
	Provider string
	// Subject is the immutable provider-issued subject identifier used together
	// with Provider as the authoritative resolution key.
	Subject string
	// Email is captured for audit and hook context only; it is never used as a
	// resolution key.
	Email string
	// DisplayName is captured for audit and hook context only.
	DisplayName string
}

// LoginOutput contains the host login outcome for a verified external identity.
type LoginOutput struct {
	// AccessToken is set when the resolved user has zero or one active tenant.
	AccessToken string
	// RefreshToken is set when the resolved user has zero or one active tenant.
	RefreshToken string
	// PreToken is set when the resolved user has more than one active tenant;
	// the caller passes it to authcap.Token().SelectTenant for tenant selection.
	PreToken string
	// TenantCandidates lists selectable tenants when PreToken is set.
	TenantCandidates []TenantCandidate
}

// TenantCandidate describes one selectable tenant during two-stage login.
type TenantCandidate struct {
	// ID is the tenant ID.
	ID int
	// Code is the tenant code.
	Code string
	// Name is the tenant display name.
	Name string
	// Status is the tenant status.
	Status string
}
