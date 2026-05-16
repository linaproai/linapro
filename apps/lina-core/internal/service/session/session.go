// Package session implements online-session storage and activity validation.
package session

import (
	"context"
	"sync"
	"time"

	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/service/coordination"
	"lina-core/internal/service/datascope"
	tenantcapsvc "lina-core/internal/service/tenantcap"
)

// sessionLastActiveUpdateWindow is the minimum interval between two
// last_active_time writes for one valid session.
const sessionLastActiveUpdateWindow time.Duration = time.Minute

const (
	sessionHotStateComponent = "session-hot-state"
	sessionHotStateSchema    = 1
	sessionUserIndexSchema   = 1
)

// Session represents an online user session.
type Session struct {
	TokenId        string      // Unique token identifier
	TenantId       int         // Tenant ID, where 0 means platform
	UserId         int         // User ID
	Username       string      // Username
	DeptName       string      // Department name
	Ip             string      // Login IP address
	Browser        string      // Browser information
	Os             string      // Operating system
	LoginTime      *gtime.Time // Login time
	LastActiveTime *gtime.Time // Last active time
}

// ListFilter defines filter options for listing sessions.
type ListFilter struct {
	Username string // Username, supports fuzzy search
	Ip       string // Login IP, supports fuzzy search
}

// ListResult defines the result for paginated session list.
type ListResult struct {
	Items []*Session // Session items
	Total int        // Total count
}

// Store defines the session storage interface for persistent online-session
// records.
type Store interface {
	// Set persists one online session record.
	Set(ctx context.Context, session *Session) error
	// Get returns one online session by its globally unique token ID.
	Get(ctx context.Context, tokenId string) (*Session, error)
	// Delete removes one online session by its globally unique token ID.
	Delete(ctx context.Context, tokenId string) error
	// DeleteByUserId removes all online sessions that belong to one user in one tenant.
	DeleteByUserId(ctx context.Context, tenantId int, userId int) error
	// List returns all online sessions that match the optional filter.
	List(ctx context.Context, filter *ListFilter) ([]*Session, error)
	// ListPage returns one paginated online-session list for the optional filter.
	ListPage(ctx context.Context, filter *ListFilter, pageNum, pageSize int) (*ListResult, error)
	// ListPageScoped returns one paginated online-session list constrained by
	// tenant ownership and the supplied data-scope service.
	ListPageScoped(
		ctx context.Context,
		filter *ListFilter,
		pageNum, pageSize int,
		scopeSvc datascope.Service,
		tenantSvc tenantcapsvc.Service,
	) (*ListResult, error)
	// Count returns the total number of active online sessions.
	Count(ctx context.Context) (int, error)
	// TouchOrValidate validates tenant ownership and session timeout, then
	// refreshes last_active_time outside the short write-throttle window for the
	// given tokenId. It returns true when the session remains valid.
	TouchOrValidate(ctx context.Context, tenantId int, tokenId string, timeout time.Duration) (bool, error)
	// CleanupInactive deletes sessions whose last_active_time exceeds the given timeout duration.
	CleanupInactive(ctx context.Context, timeout time.Duration) (int64, error)
}

// DBStore implements Store using the persistent online-session table.
type DBStore struct{}

// SessionConfigurableStore extends Store with runtime session-timeout
// propagation for hot-state implementations.
type SessionConfigurableStore interface {
	Store
	// SetDefaultTTL updates the hot-state TTL used for login-time writes.
	SetDefaultTTL(ttl time.Duration)
}

// processCoordinationSessionStore stores the deployment-selected coordination
// backend used by session stores created after HTTP startup configuration.
var processCoordinationSessionStore = struct {
	sync.RWMutex
	service coordination.Service
}{}

// NewDBStore creates a new DBStore instance.
func NewDBStore() Store {
	dbStore := &DBStore{}
	if coordinationSvc := currentCoordinationService(); coordinationSvc != nil {
		return NewCoordinationStore(coordinationSvc, dbStore)
	}
	return dbStore
}

// ConfigureCoordination switches new session stores to a coordination-backed
// hot-state implementation. Passing nil restores the DB-only implementation
// used by single-node deployments and tests.
func ConfigureCoordination(coordinationSvc coordination.Service) {
	processCoordinationSessionStore.Lock()
	processCoordinationSessionStore.service = coordinationSvc
	processCoordinationSessionStore.Unlock()
}

// currentCoordinationService returns the process-selected coordination service.
func currentCoordinationService() coordination.Service {
	processCoordinationSessionStore.RLock()
	coordinationSvc := processCoordinationSessionStore.service
	processCoordinationSessionStore.RUnlock()
	return coordinationSvc
}
