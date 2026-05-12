// Package sysinfo implements runtime information aggregation for the version
// page and related host diagnostics.
package sysinfo

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/gogf/gf/v2"
	"github.com/gogf/gf/v2/frame/g"

	"lina-core/internal/service/cachecoord"
	"lina-core/internal/service/cluster"
	"lina-core/internal/service/config"
	"lina-core/internal/service/coordination"
	"lina-core/pkg/dialect"
	"lina-core/pkg/logger"
)

// Component section keys used when projecting metadata.yaml into the system-info response.
const (
	componentSectionBackend  = "backend"
	componentSectionFrontend = "frontend"
)

// Service defines the sysinfo service contract.
type Service interface {
	// GetInfo returns system runtime information.
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

// New creates and returns a new Service instance.
func New() Service {
	configSvc := config.New()
	return &serviceImpl{
		startTime: time.Now(),
		configSvc: configSvc,
		cacheCoordSvc: cachecoord.Default(
			cachecoord.NewStaticTopology(configSvc.IsClusterEnabled(context.Background())),
		),
	}
}

// NewWithDiagnostics creates a sysinfo service using runtime-owned cluster and
// coordination services so diagnostics report the actual HTTP process wiring.
func NewWithDiagnostics(
	configSvc config.Service,
	clusterSvc cluster.Service,
	coordinationSvc coordination.Service,
	cacheCoordSvc cachecoord.Service,
) Service {
	if configSvc == nil {
		configSvc = config.New()
	}
	if cacheCoordSvc == nil {
		cacheCoordSvc = cachecoord.Default(
			cachecoord.NewStaticTopology(configSvc.IsClusterEnabled(context.Background())),
		)
	}
	return &serviceImpl{
		startTime:       time.Now(),
		configSvc:       configSvc,
		clusterSvc:      clusterSvc,
		coordinationSvc: coordinationSvc,
		cacheCoordSvc:   cacheCoordSvc,
	}
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

// GetInfo returns system runtime information.
func (s *serviceImpl) GetInfo(ctx context.Context) (*SystemInfo, error) {
	metadata := s.configSvc.GetMetadata(ctx)
	info := &SystemInfo{
		Framework: FrameworkInfo{
			Name:          metadata.Framework.Name,
			Version:       metadata.Framework.Version,
			Description:   metadata.Framework.Description,
			Homepage:      metadata.Framework.Homepage,
			RepositoryURL: metadata.Framework.RepositoryURL,
			License:       metadata.Framework.License,
		},
		GoVersion: runtime.Version(),
		GfVersion: gf.VERSION,
		Os:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		StartTime: s.startTime.Format("2006-01-02 15:04:05"),
	}

	durationSeconds := int64(time.Since(s.startTime).Seconds())
	info.RunDurationSeconds = durationSeconds
	info.RunDuration = formatRunDurationFallback(durationSeconds)

	dbVersion, err := s.getDbVersion(ctx)
	if err != nil {
		logger.Warningf(ctx, "Failed to get database version: %v", err)
		info.DbVersion = "unknown"
	} else {
		info.DbVersion = dbVersion
	}

	info.BackendComponents = s.loadComponents(metadata, componentSectionBackend, dbVersion)
	info.FrontendComponents = s.loadComponents(metadata, componentSectionFrontend, "")
	info.Coordination = s.loadCoordination(ctx)
	info.CacheCoordination = s.loadCacheCoordination(ctx)

	return info, nil
}

// loadCoordination returns process-level cluster coordination diagnostics
// without exposing passwords, Redis addresses, or token-bearing keys.
func (s *serviceImpl) loadCoordination(ctx context.Context) CoordinationInfo {
	info := CoordinationInfo{}
	if s.clusterSvc != nil {
		info.ClusterEnabled = s.clusterSvc.IsEnabled()
		info.NodeID = s.clusterSvc.NodeID()
		info.Primary = s.clusterSvc.IsPrimary()
	} else if s.configSvc != nil {
		info.ClusterEnabled = s.configSvc.IsClusterEnabled(ctx)
		info.Primary = !info.ClusterEnabled
	}
	if info.NodeID == "" {
		info.NodeID = "local-node"
	}
	if s.coordinationSvc == nil || s.coordinationSvc.Health() == nil {
		return info
	}

	snapshot := s.coordinationSvc.Health().Snapshot(ctx)
	info.Backend = snapshot.Backend
	info.RedisHealthy = snapshot.Backend == coordination.BackendRedis && snapshot.Healthy
	info.LastSuccessAt = snapshot.LastSuccessAt
	info.LastError = sanitizeCoordinationError(snapshot.LastError)
	return info
}

// loadCacheCoordination returns best-effort cache coordination diagnostics.
func (s *serviceImpl) loadCacheCoordination(ctx context.Context) []CacheCoordinationInfo {
	if s.cacheCoordSvc == nil {
		return nil
	}

	items, err := s.cacheCoordSvc.Snapshot(ctx)
	if err != nil {
		logger.Warningf(ctx, "Failed to get cache coordination snapshot: %v", err)
		return nil
	}
	diagnostics := make([]CacheCoordinationInfo, 0, len(items))
	for _, item := range items {
		diagnostics = append(diagnostics, CacheCoordinationInfo{
			Domain:           string(item.Domain),
			Scope:            string(item.Scope),
			AuthoritySource:  item.AuthoritySource,
			ConsistencyModel: string(item.ConsistencyModel),
			MaxStale:         item.MaxStale,
			FailureStrategy:  string(item.FailureStrategy),
			Backend:          item.Backend,
			Healthy:          item.CoordinationHealthy,
			LocalRevision:    item.LocalRevision,
			SharedRevision:   item.SharedRevision,
			LastSyncedAt:     item.LastSyncedAt,
			EventSubscriber:  item.EventSubscriberRunning,
			LastEventAt:      item.LastEventReceivedAt,
			RecentError:      sanitizeCoordinationError(item.RecentError),
			StaleSeconds:     item.StaleSeconds,
		})
	}
	return diagnostics
}

// sanitizeCoordinationError removes connection strings, key material, and
// credentials from health diagnostics while preserving an actionable category.
func sanitizeCoordinationError(message string) string {
	trimmed := strings.TrimSpace(message)
	if trimmed == "" {
		return ""
	}
	lower := strings.ToLower(trimmed)
	switch {
	case strings.Contains(lower, "redis://"),
		strings.Contains(lower, "password"),
		strings.Contains(lower, "token"),
		strings.Contains(lower, "linapro:"):
		return "coordination backend error"
	case strings.Contains(lower, "dial tcp"),
		strings.Contains(lower, "connection refused"),
		strings.Contains(lower, "connection reset"):
		return "redis coordination connection failed"
	default:
		return trimmed
	}
}

// formatRunDurationFallback formats uptime with an English developer fallback.
func formatRunDurationFallback(totalSeconds int64) string {
	if totalSeconds < 0 {
		totalSeconds = 0
	}

	hours := totalSeconds / int64(time.Hour/time.Second)
	minutes := totalSeconds / int64(time.Minute/time.Second) % 60
	seconds := totalSeconds % 60
	if hours > 0 {
		return fmt.Sprintf("%d hours %d minutes %d seconds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%d minutes %d seconds", minutes, seconds)
	}
	return fmt.Sprintf("%d seconds", seconds)
}

// loadComponents reads version-page component metadata from metadata.yaml.
func (s *serviceImpl) loadComponents(metadata *config.MetadataConfig, sectionKey string, dbVersion string) []ComponentInfo {
	var source []config.MetadataComponentInfo
	switch sectionKey {
	case componentSectionBackend:
		source = metadata.Backend
	case componentSectionFrontend:
		source = metadata.Frontend
	default:
		return nil
	}

	if len(source) == 0 {
		return nil
	}

	components := make([]ComponentInfo, 0, len(source))
	for _, item := range source {
		component := ComponentInfo{
			Name:        item.Name,
			Version:     item.Version,
			Url:         item.Url,
			Description: item.Description,
		}
		if component.Version == "auto" {
			switch component.Name {
			case "GoFrame":
				component.Version = gf.VERSION
			case "PostgreSQL":
				if dbVersion != "" {
					component.Version = dbVersion
				}
			}
		}
		components = append(components, component)
	}

	return components
}

// getDbVersion retrieves the database version.
func (s *serviceImpl) getDbVersion(ctx context.Context) (string, error) {
	return dialect.DatabaseVersion(ctx, g.DB())
}
