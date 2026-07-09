// Package externalidentityspi defines the source-plugin provider SPI that backs
// the external-identity login capability. The parent externallogin package
// remains the ordinary consumer contract; this package owns the seam a source
// plugin implements to resolve, provision, and link external identities to
// local accounts.
//
// Boundary: the provider is responsible only for mapping a verified external
// identity (provider + immutable subject) to a local user ID and for
// host-owned account provisioning and bind orchestration. Token minting,
// session persistence, tenant resolution, pre-login-token handoff, login-IP
// policy, disabled-account checks, and auth hooks stay in the host auth
// service and MUST NOT appear on this contract. When no provider is
// registered, the host external-login path is fail-closed (see the host auth
// service): no account is created and no session is issued.
package externalidentityspi

import (
	"context"
	"errors"
)

// ErrProviderUnavailable is the fail-closed sentinel returned by the host
// manager-backed Provider when no external-identity provider plugin is
// installed and enabled. The host auth service maps it to a neutral
// not-provisioned outcome, minting no account and no session. It carries no
// account-existence signal so external login never leaks whether a captured
// email already exists.
var ErrProviderUnavailable = errors.New("externalidentityspi: no external-identity provider is available")

// Provider is the external-identity capability implemented by a source plugin
// (for example linapro-oidc-core). The host holds at most one active provider
// and injects it into the auth service at startup. A nil provider keeps
// external login fail-closed.
type Provider interface {
	// Resolve maps a verified (provider, subject) pair to a linked local user
	// ID. found is false when no linkage exists; err is reserved for
	// infrastructure failures, not for the not-found case.
	Resolve(ctx context.Context, in ResolveInput) (userID int, found bool, err error)
	// Provision runs the host-owned auto-provisioning policy for one unlinked
	// verified identity and returns the resulting linked user ID. The provider
	// owns the policy (email-conflict rejection, email-less anchor derivation,
	// idempotent reuse keyed on the (provider, subject) link), while the actual
	// least-privilege account creation is delegated back to the host user owner.
	Provision(ctx context.Context, in ProvisionInput) (userID int, err error)
	// Bind links a verified external identity to an already-authenticated
	// session user. It MUST only act on the current session user's own links
	// and reject a (provider, subject) already owned by another account.
	Bind(ctx context.Context, in BindInput) error
	// Unbind removes one of the current session user's external-identity links.
	Unbind(ctx context.Context, in UnbindInput) error
	// List returns the current session user's bound external identities.
	List(ctx context.Context, userID int) ([]BoundIdentity, error)
}

// ResolveInput carries the authoritative resolution key for an external
// identity. Provider and Subject together form the unique key; a matching
// Subject under a different Provider is a distinct identity.
type ResolveInput struct {
	// Provider is the stable external provider ID owned by the calling plugin.
	Provider string
	// Subject is the immutable provider-issued subject identifier.
	Subject string
}

// ProvisionInput describes one unlinked verified identity to provision and
// link. Email is audit/context only and MUST NOT be used as a resolution key;
// UsernameAnchor lets email-less providers (for example WeChat) drive a
// deterministic derived username.
type ProvisionInput struct {
	// Provider is the stable external provider ID owned by the calling plugin.
	Provider string
	// Subject is the immutable provider-issued subject identifier.
	Subject string
	// Email is the verified email when the provider supplies one; empty for
	// email-less providers. It is captured for audit and same-email conflict
	// policy only, never as a resolution key.
	Email string
	// DisplayName optionally seeds the account nickname.
	DisplayName string
	// UsernameAnchor is an optional deterministic anchor used to derive a
	// username when Email is empty. It MUST be collision-resistant per distinct
	// (Provider, Subject) so distinct identities cannot alias onto one account.
	UsernameAnchor string
	// PluginID is the host-stamped calling plugin identity, recorded on the
	// link row for audit and ownership.
	PluginID string
	// AllowAutoProvision declares that the calling plugin permits host-owned
	// auto-provisioning for this login. The host policy still applies.
	AllowAutoProvision bool
}

// BindInput links a verified external identity to the current session user.
type BindInput struct {
	// UserID is the current session user; binds MUST NOT target another user.
	UserID int
	// Provider is the stable external provider ID owned by the calling plugin.
	Provider string
	// Subject is the immutable provider-issued subject identifier.
	Subject string
	// Email is captured for audit context only.
	Email string
	// PluginID is the host-stamped calling plugin identity.
	PluginID string
}

// UnbindInput removes one of the current session user's external-identity
// links.
type UnbindInput struct {
	// UserID is the current session user; unbinds MUST NOT target another user.
	UserID int
	// Provider is the stable external provider ID owned by the calling plugin.
	Provider string
	// Subject is the immutable provider-issued subject identifier.
	Subject string
}

// BoundIdentity is one external-identity link owned by a user.
type BoundIdentity struct {
	// Provider is the stable external provider ID.
	Provider string
	// Subject is the immutable provider-issued subject identifier.
	Subject string
	// Email is the captured email snapshot, if any.
	Email string
}
