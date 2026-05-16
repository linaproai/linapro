// Package sysinfo implements runtime information aggregation for the version
// page and related host diagnostics.
package sysinfo

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/service/cachecoord"
	"lina-core/internal/service/cluster"
	"lina-core/internal/service/config"
	"lina-core/internal/service/coordination"
)

// Component section keys used when projecting metadata.yaml into the system-info response.
const (
	componentSectionBackend  = "backend"
	componentSectionFrontend = "frontend"
)

// Service defines the sysinfo service contract.
type Service interface {
	// GetInfo returns system runtime information for diagnostics and the
	// version page. The response includes framework metadata, runtime versions,
	// uptime, database version, cluster coordination health, and cache
	// coordination status. Database version lookup failures are logged and
	// degraded to "unknown"; required config/cache dependencies are validated at
	// construction.
	GetInfo(ctx context.Context) (*SystemInfo, error)
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	startTime       time.Time            // startTime stores the host service boot time.
	configSvc       config.Service       // configSvc loads embedded metadata and runtime config values.
	clusterSvc      cluster.Service      // clusterSvc exposes cluster topology diagnostics.
	coordinationSvc coordination.Service // coordinationSvc exposes distributed coordination health.
	cacheCoordSvc   cachecoord.Service   // cacheCoordSvc exposes process-wide cache coordination diagnostics.
}

// New creates and returns a new sysinfo service from explicit runtime-owned dependencies.
func New(
	configSvc config.Service,
	clusterSvc cluster.Service,
	coordinationSvc coordination.Service,
	cacheCoordSvc cachecoord.Service,
) (Service, error) {
	if configSvc == nil {
		return nil, gerror.New("sysinfo service requires a non-nil config service")
	}
	if cacheCoordSvc == nil {
		return nil, gerror.New("sysinfo service requires a non-nil cache coordination service")
	}
	return &serviceImpl{
		startTime:       time.Now(),
		configSvc:       configSvc,
		clusterSvc:      clusterSvc,
		coordinationSvc: coordinationSvc,
		cacheCoordSvc:   cacheCoordSvc,
	}, nil
}

// SystemInfo holds the system runtime information.
type SystemInfo struct {
	Framework          FrameworkInfo           // Framework contains top-level framework metadata.
	GoVersion          string                  // GoVersion is the active Go runtime version.
	GfVersion          string                  // GfVersion is the active GoFrame runtime version.
	Os                 string                  // Os is the operating system name.
	Arch               string                  // Arch is the runtime architecture.
	DbVersion          string                  // DbVersion is the database server version string.
	StartTime          string                  // StartTime is the host start timestamp.
	RunDuration        string                  // RunDuration is the English fallback uptime string.
	RunDurationSeconds int64                   // RunDurationSeconds is the total uptime in seconds.
	BackendComponents  []ComponentInfo         // BackendComponents lists backend technology cards.
	FrontendComponents []ComponentInfo         // FrontendComponents lists frontend technology cards.
	Coordination       CoordinationInfo        // Coordination reports cluster coordination health.
	CacheCoordination  []CacheCoordinationInfo // CacheCoordination lists critical cache coordination diagnostics.
}

// FrameworkInfo holds framework-level project information.
type FrameworkInfo struct {
	Name          string // Name is the framework display name.
	Version       string // Version is the current framework version.
	Description   string // Description is the framework summary.
	Homepage      string // Homepage is the framework website.
	RepositoryURL string // RepositoryURL is the framework source repository URL.
	License       string // License is the framework license label.
}

// ComponentInfo holds component display information.
type ComponentInfo struct {
	Name        string // Name is the component display name.
	Version     string // Version is the resolved component version label.
	Url         string // Url is the component homepage.
	Description string // Description is the short component summary.
}

// CacheCoordinationInfo holds one cache coordination diagnostic row.
type CacheCoordinationInfo struct {
	Domain           string                   // Domain is the cache coordination domain identifier.
	Scope            string                   // Scope is the explicit invalidation scope inside the domain.
	AuthoritySource  string                   // AuthoritySource is the canonical data source for rebuilds.
	ConsistencyModel string                   // ConsistencyModel is the declared freshness model.
	MaxStale         time.Duration            // MaxStale is the configured stale window.
	FailureStrategy  string                   // FailureStrategy is the caller-visible degradation behavior.
	Backend          coordination.BackendName // Backend is the active coordination backend.
	Healthy          bool                     // Healthy reports coordination backend health.
	LocalRevision    int64                    // LocalRevision is the latest revision consumed by this process.
	SharedRevision   int64                    // SharedRevision is the latest shared coordination revision observed.
	LastSyncedAt     time.Time                // LastSyncedAt is the latest successful local synchronization time.
	EventSubscriber  bool                     // EventSubscriber reports whether event consumption is active.
	LastEventAt      time.Time                // LastEventAt records the latest consumed coordination event.
	RecentError      string                   // RecentError is the most recent coordination failure.
	StaleSeconds     int64                    // StaleSeconds is the elapsed time since LastSyncedAt.
}

// CoordinationInfo holds cluster coordination health diagnostics.
type CoordinationInfo struct {
	ClusterEnabled bool                     // ClusterEnabled reports whether clustered deployment mode is active.
	Backend        coordination.BackendName // Backend is the active distributed coordination backend.
	RedisHealthy   bool                     // RedisHealthy reports Redis health when Redis is the backend.
	NodeID         string                   // NodeID is the current host node identifier.
	Primary        bool                     // Primary reports whether this node owns primary work.
	LastSuccessAt  time.Time                // LastSuccessAt records the latest successful health check.
	LastError      string                   // LastError stores a sanitized recent coordination error.
}
