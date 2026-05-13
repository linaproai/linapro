// This file defines the source-plugin visible online-session contract.

package contract

import (
	"context"

	"github.com/gogf/gf/v2/os/gtime"
)

// Session is the stable online-session projection published to source plugins.
type Session struct {
	// TokenId is the unique token identifier.
	TokenId string
	// TenantId is the owning tenant identifier, where 0 means platform.
	TenantId int
	// UserId is the authenticated user identifier.
	UserId int
	// Username is the authenticated username.
	Username string
	// DeptName is the projected department display name.
	DeptName string
	// Ip is the login IP address.
	Ip string
	// Browser is the login browser fingerprint.
	Browser string
	// Os is the login operating system fingerprint.
	Os string
	// LoginTime is the first login time of this session.
	LoginTime *gtime.Time
	// LastActiveTime is the most recent activity time tracked by the host.
	LastActiveTime *gtime.Time
}

// ListFilter is the stable session-list filter contract published to plugins.
type ListFilter struct {
	// Username filters sessions by username using fuzzy matching.
	Username string
	// Ip filters sessions by login IP using fuzzy matching.
	Ip string
}

// ListResult is the stable paged session-list result published to plugins.
type ListResult struct {
	// Items is the current page of online sessions.
	Items []*Session
	// Total is the total number of matching sessions.
	Total int
}

// SessionService defines online-session operations published to source plugins.
type SessionService interface {
	// ListPage returns one paginated online-session list for the optional filter.
	ListPage(ctx context.Context, filter *ListFilter, pageNum, pageSize int) (*ListResult, error)
	// Revoke invalidates one online session by token ID.
	Revoke(ctx context.Context, tokenID string) error
}
