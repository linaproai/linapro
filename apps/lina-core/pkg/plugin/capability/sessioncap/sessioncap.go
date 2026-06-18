// Package sessioncap defines online-session capability contracts for plugins
// without exposing sys_online_session storage.
package sessioncap

import (
	"context"
	"lina-core/pkg/plugin/capability/capmodel"
	"time"
)

const (
	// MaxBatchGetUserOnlineStatus limits one user online-status batch call.
	MaxBatchGetUserOnlineStatus = 200
	// MaxEnsureVisible limits one online-session visibility check.
	MaxEnsureVisible = 200
)

// SessionID identifies one online session token at domain boundaries.
type SessionID string

// Projection describes one online session visible to a plugin.
type Projection struct {
	// ID is the session domain identifier.
	ID SessionID
	// TenantID is the current tenant identifier.
	TenantID capmodel.DomainID
	// UserID is the session user identifier.
	UserID string
	// Username is the authenticated username.
	Username string
	// ClientType is the client type.
	ClientType string
	// DeptName is the department display snapshot captured at login time.
	DeptName string
	// Ip is the login IP address snapshot.
	Ip string
	// Browser is the browser information snapshot.
	Browser string
	// Os is the operating system snapshot.
	Os string
	// LoginAt is the login timestamp.
	LoginAt *time.Time
	// LastActiveAt is the last activity timestamp.
	LastActiveAt *time.Time
}

// UserOnlineStatusProjection reports whether one visible user has an online session.
type UserOnlineStatusProjection struct {
	// UserID is the requested user identifier.
	UserID string
	// Online reports whether the user currently has at least one visible session.
	Online bool
	// SessionCount is the number of visible online sessions for the user.
	SessionCount int
}

// SearchInput constrains online-session queries.
type SearchInput struct {
	// Username filters by username.
	Username string
	// IP filters by login IP.
	IP string
	// Page constrains result size.
	Page capmodel.PageRequest
}

// Service defines read-oriented online-session capability methods.
type Service interface {
	// Current returns the visible session projection for the current token.
	Current(ctx context.Context, capCtx capmodel.CapabilityContext) (*Projection, error)
	// Search returns one bounded visible session page.
	Search(ctx context.Context, capCtx capmodel.CapabilityContext, input SearchInput) (*capmodel.PageResult[*Projection], error)
	// BatchGet returns visible sessions and opaque missing IDs.
	BatchGet(ctx context.Context, capCtx capmodel.CapabilityContext, ids []SessionID) (*capmodel.BatchResult[*Projection, SessionID], error)
	// BatchGetUserOnlineStatus returns visible users' online status in one bounded call.
	BatchGetUserOnlineStatus(ctx context.Context, capCtx capmodel.CapabilityContext, userIDs []string) (*capmodel.BatchResult[*UserOnlineStatusProjection, string], error)
	// EnsureVisible rejects when any requested online session is absent or invisible.
	EnsureVisible(ctx context.Context, capCtx capmodel.CapabilityContext, ids []SessionID) error
}

// AdminService defines session management commands.
type AdminService interface {
	// Revoke invalidates one visible online session.
	Revoke(ctx context.Context, capCtx capmodel.CapabilityContext, id SessionID) error
}

// ScopeService defines host-internal session visibility helpers.
type ScopeService interface {
	// EnsureVisible rejects when any session is outside caller scope.
	EnsureVisible(ctx context.Context, capCtx capmodel.CapabilityContext, ids []SessionID) error
}
