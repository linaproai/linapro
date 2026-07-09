// Package usercap defines the stable user-domain capability contract exposed to
// plugins without leaking sys_user storage or host DAO models.
package usercap

import (
	"context"
	"errors"

	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/statusflag"
)

// ErrProvisionEmailConflict is the stable sentinel returned by
// ProvisionExternal when an existing local account already uses the asserted
// email. The host minting primitive enforces this safety invariant itself
// (the unauthenticated login path has no actor context for data-scoped
// lookups); plugin callers detect it with errors.Is and map it into their own
// caller-visible conflict policy instead of minting or silently linking.
var ErrProvisionEmailConflict = errors.New("usercap: an account with the same email already exists")

// Service defines governed user capability methods for plugins. Read methods
// use tenant/data-permission filtering and bounded batch or page sizes; command
// methods must validate target visibility, tenant boundary, status values,
// audit source, and cache impact before mutating host user state.
type Service interface {
	// Current returns the current actor's visible user info.
	Current(ctx context.Context) (*UserInfo, error)
	// Get returns one visible user info record. Risk: read; resource: user ID;
	// context: actor and tenant; data permission: target visibility check;
	// performance: delegates to bounded BatchGet; audit/cache: read-only.
	Get(ctx context.Context, id UserID) (*UserInfo, error)
	// BatchGet returns visible user info records and opaque missing IDs.
	BatchGet(ctx context.Context, ids []UserID) (*capmodel.BatchResult[*UserInfo, UserID], error)
	// BatchResolve resolves visible users by IDs, usernames, email addresses, or phone numbers.
	BatchResolve(ctx context.Context, input BatchResolveInput) (*capmodel.BatchResult[*UserInfo, ResolveKey], error)
	// List returns visible user candidates with bounded paging. Risk: read;
	// resource: user query scope; context: actor and tenant; data permission:
	// database-side data-scope filtering; performance: bounded page and
	// info query; audit/cache: read-only.
	List(ctx context.Context, input ListInput) (*capmodel.PageResult[*UserInfo], error)
	// EnsureVisible rejects when any requested user is absent or invisible.
	EnsureVisible(ctx context.Context, ids []UserID) error
	// Create creates one governed user through the host user owner. Risk:
	// mutate; resource: tenant, role and organization references; context:
	// actor and tenant; data permission: create boundary validation;
	// performance: one transaction; audit/cache: authorization revision impact.
	Create(ctx context.Context, input CreateInput) (UserID, error)
	// ProvisionExternal mints one least-privilege platform account for a
	// verified external identity, distinct from operator-driven Create. Unlike
	// Create it takes no operator context, tenant, role or create-boundary
	// validation: the username is derived from the email local part, or from a
	// deterministic UsernameAnchor when Email is empty, the password is random
	// and unusable, and no roles or tenants are assigned. Risk: mutate; context:
	// no acting operator (system provisioning); data permission: none — callers
	// decide WHEN provisioning is allowed and own idempotent link de-duplication;
	// performance: one insert; audit/cache: none. This method is source-plugin
	// only; dynamic (WASM) plugins receive a fail-closed stub because an
	// operator-less account-minting primitive must not be exposed to sandboxed
	// guests. Returns the new user ID, an ErrProvisionEmailConflict sentinel
	// when an existing local account already uses the asserted email (the
	// minting primitive enforces this takeover-safety invariant itself because
	// the login path has no actor context), or an error when neither a valid
	// email nor an anchor is supplied.
	ProvisionExternal(ctx context.Context, input ProvisionExternalInput) (UserID, error)
	// Update mutates one visible user through the host user owner. Risk:
	// mutate; resource: target user and relation references; context: actor and
	// tenant; data permission: target visibility check; performance: one
	// transaction; audit/cache: authorization revision impact.
	Update(ctx context.Context, input UpdateInput) error
	// Delete deletes one visible user through the host user owner. Risk:
	// mutate; resource: target user; context: actor and tenant; data
	// permission: target visibility check; performance: one transaction;
	// audit/cache: authorization revision impact.
	Delete(ctx context.Context, id UserID) error
	// SetStatus changes one visible user's lifecycle status after validating
	// tenant scope, data permission, allowed status values, audit source, and
	// authorization-cache revision impact.
	SetStatus(ctx context.Context, id UserID, status statusflag.Enabled) error
	// ResetPassword resets one visible user's password through the host auth
	// owner. Risk: manage; resource: target user; context: actor and tenant;
	// data permission: target visibility check; performance: one owner write;
	// audit/cache: credential/session side effects remain host-owned.
	ResetPassword(ctx context.Context, id UserID, password string) error
	// Assignment returns user role assignment operations.
	Assignment() AssignmentService
}

// AssignmentService defines user-role assignment operations exposed under the
// user domain subresource.
type AssignmentService interface {
	// ReplaceRoles replaces one visible user's role assignments. Risk: manage;
	// resource: target user and role IDs; context: actor and tenant; data
	// permission: user and role visibility checks; performance: one owner
	// transaction; audit/cache: authorization revision impact.
	ReplaceRoles(ctx context.Context, id UserID, roleIDs []int) error
}

const (
	// MaxBatchResolveIDs limits one user batch-resolve call by user ID count.
	MaxBatchResolveIDs = 100
	// MaxBatchResolveUsernames limits one user batch-resolve call by username count.
	MaxBatchResolveUsernames = 100
	// MaxBatchResolveContacts limits one user batch-resolve call by phone or email count.
	MaxBatchResolveContacts = 100
	// MaxBatchResolveKeys limits the normalized key count across all resolve dimensions.
	MaxBatchResolveKeys = 300
)

// UserID identifies a user at plugin-visible domain boundaries.
type UserID string

// ResolveKey identifies one requested user lookup key without exposing which
// lookup dimension matched a visible user.
type ResolveKey string

// UserInfo is the minimal user display information exposed to plugins.
type UserInfo struct {
	// ID is the user domain identifier.
	ID UserID
	// Username is the stable login name.
	Username string
	// Nickname is the display name.
	Nickname string
	// Avatar is an optional avatar URL or governed file reference.
	Avatar string
	// Status is the stable user lifecycle status.
	Status statusflag.Enabled
	// TenantID is the owning tenant domain identifier.
	TenantID capmodel.DomainID
	// LabelKey is the optional i18n key for synthetic labels.
	LabelKey string
	// Label is the optional locale-resolved display label.
	Label string
}

// ListInput constrains bounded user candidate listing.
type ListInput struct {
	// Keyword filters visible users by username, nickname, or phone/email owner fields.
	Keyword string
	// Status filters by user lifecycle state. Empty includes all visible states.
	Status *statusflag.Enabled
	// TenantID optionally narrows results to one visible tenant.
	TenantID capmodel.DomainID
	// EnabledOnly is a convenience filter for enabled user candidates.
	EnabledOnly bool
	// Page constrains candidate size and sorting.
	Page capmodel.PageRequest
}

// BatchResolveInput constrains user lookup by stable domain IDs and login or
// contact identifiers. Missing results must not distinguish absent users from
// tenant or data-permission filtering.
type BatchResolveInput struct {
	// IDs contains user domain identifiers.
	IDs []UserID
	// Usernames contains stable login names.
	Usernames []string
	// Contacts contains email addresses or phone numbers.
	Contacts []string
}

// CreateInput describes one governed user create request.
type CreateInput struct {
	// Username is the stable login name.
	Username string
	// Password is the initial password that the host auth owner hashes.
	Password string
	// Nickname is the display name. Empty lets the owner apply defaults.
	Nickname string
	// Email is the optional email address.
	Email string
	// Phone is the optional phone number.
	Phone string
	// Sex is the host user gender code.
	Sex int
	// Status is the initial lifecycle status.
	Status *statusflag.Enabled
	// Remark stores optional operator notes.
	Remark string
	// DeptID is the optional primary organization department identifier.
	DeptID *int
	// PostIDs are optional organization post identifiers.
	PostIDs []int
	// RoleIDs are role identifiers assigned to the user.
	RoleIDs []int
	// TenantIDs are active tenant memberships requested for the user.
	TenantIDs []int
}

// ProvisionExternalInput describes one least-privilege external-identity
// provisioning request. Email is the verified address asserted by the external
// identity provider and may be empty for email-less providers (for example
// WeChat); when Email is empty, UsernameAnchor is required so the host can
// derive a stable, collision-resistant username without an email local part.
type ProvisionExternalInput struct {
	// Email is the verified email address from the external provider. It may be
	// empty for email-less providers, in which case UsernameAnchor is required.
	Email string
	// DisplayName optionally seeds the nickname. Empty lets the owner apply
	// defaults derived from the username.
	DisplayName string
	// Remark records the provisioning source for audit, e.g. the provider ID.
	Remark string
	// UsernameAnchor is a deterministic anchor used to derive a username when
	// Email is empty. It MUST be collision-resistant per distinct external
	// identity so two identities cannot alias onto one account.
	UsernameAnchor string
}

// UpdateInput describes one governed user update request.
type UpdateInput struct {
	// ID identifies the target user.
	ID UserID
	// Username optionally updates the stable login name.
	Username *string
	// Password optionally updates the password through the host auth owner.
	Password *string
	// Nickname optionally updates the display name.
	Nickname *string
	// Email optionally updates the email address.
	Email *string
	// Phone optionally updates the phone number.
	Phone *string
	// Sex optionally updates the host gender code.
	Sex *int
	// Status optionally updates the lifecycle status.
	Status *statusflag.Enabled
	// Remark optionally updates operator notes.
	Remark *string
	// DeptID optionally replaces the primary organization department.
	DeptID *int
	// PostIDs optionally replaces organization post assignments when non-nil.
	PostIDs []int
	// RoleIDs optionally replaces role assignments when non-nil.
	RoleIDs []int
	// TenantIDs optionally replaces tenant memberships when non-nil.
	TenantIDs []int
}
