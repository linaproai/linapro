// Package sessioncap defines online-session capability contracts for plugins
// without exposing sys_online_session storage.
package sessioncap

import (
	"context"
	"time"

	"lina-core/pkg/plugin/capability/capmodel"
)

// Service defines governed online-session capability methods. Reads apply
// tenant and data-scope filtering with bounded batch or page sizes; revocation
// validates target visibility and delegates token invalidation to the host auth
// owner.
type Service interface {
	// Current returns the visible session info for the current token.
	Current(ctx context.Context) (*SessionInfo, error)
	// Get returns one visible session info.
	Get(ctx context.Context, id SessionID) (*SessionInfo, error)
	// List returns one bounded visible session page.
	List(ctx context.Context, input ListInput) (*capmodel.PageResult[*SessionInfo], error)
	// BatchGet returns visible session info records and opaque missing IDs.
	BatchGet(ctx context.Context, ids []SessionID) (*capmodel.BatchResult[*SessionInfo, SessionID], error)
	// BatchGetUserOnlineStatus returns visible users' online status in one bounded call.
	BatchGetUserOnlineStatus(ctx context.Context, userIDs []string) (*capmodel.BatchResult[*UserOnlineStatus, string], error)
	// EnsureVisible rejects when any requested online session is absent or invisible.
	EnsureVisible(ctx context.Context, ids []SessionID) error
	// Revoke invalidates one visible online session after tenant, data-scope,
	// target visibility, actor, and audit-boundary checks.
	Revoke(ctx context.Context, id SessionID) error
	// RevokeMany invalidates visible online sessions after bounded target
	// visibility checks. Any invisible target rejects the whole operation.
	RevokeMany(ctx context.Context, ids []SessionID) error
}

const (
	// MaxBatchGetUserOnlineStatus limits one user online-status batch call.
	MaxBatchGetUserOnlineStatus = 200
	// MaxEnsureVisible limits one online-session visibility check.
	MaxEnsureVisible = 200
)

// SessionID identifies one online session token at domain boundaries.
type SessionID string

// SessionInfo describes one online session visible to a plugin.
type SessionInfo struct {
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

// UserOnlineStatus reports whether one visible user has an online session.
type UserOnlineStatus struct {
	// UserID is the requested user identifier.
	UserID string
	// Online reports whether the user currently has at least one visible session.
	Online bool
	// SessionCount is the number of visible online sessions for the user.
	SessionCount int
}

// ListInput constrains online-session queries.
type ListInput struct {
	// Username filters by username.
	Username string
	// IP filters by login IP.
	IP string
	// Page constrains result size.
	Page capmodel.PageRequest
}
