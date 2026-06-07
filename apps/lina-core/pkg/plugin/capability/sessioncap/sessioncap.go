// Package sessioncap defines online-session capability contracts for plugins
// without exposing sys_online_session storage.
package sessioncap

import (
	"context"
	"lina-core/pkg/plugin/capability/capmodel"
	"time"
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
	// SearchSessions returns one bounded visible session page.
	SearchSessions(ctx context.Context, capCtx capmodel.CapabilityContext, input SearchInput) (*capmodel.PageResult[*Projection], error)
	// BatchGetSessions returns visible sessions and opaque missing IDs.
	BatchGetSessions(ctx context.Context, capCtx capmodel.CapabilityContext, ids []SessionID) (*capmodel.BatchResult[*Projection, SessionID], error)
}

// AdminService defines session management commands.
type AdminService interface {
	// RevokeSession invalidates one visible online session.
	RevokeSession(ctx context.Context, capCtx capmodel.CapabilityContext, id SessionID) error
}

// ScopeService defines host-internal session visibility helpers.
type ScopeService interface {
	// EnsureSessionsVisible rejects when any session is outside caller scope.
	EnsureSessionsVisible(ctx context.Context, capCtx capmodel.CapabilityContext, ids []SessionID) error
}
