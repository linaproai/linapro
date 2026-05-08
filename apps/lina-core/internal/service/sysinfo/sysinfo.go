// Package sysinfo implements runtime information aggregation for the version
// page and related host diagnostics.
package sysinfo

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/gogf/gf/v2"
	"github.com/gogf/gf/v2/frame/g"

	"lina-core/internal/service/cachecoord"
	"lina-core/internal/service/config"
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
	startTime     time.Time          // startTime stores the host service boot time.
	configSvc     config.Service     // configSvc loads embedded metadata and runtime config values.
	cacheCoordSvc cachecoord.Service // cacheCoordSvc exposes process-wide cache coordination diagnostics.
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
	Domain           string        // Domain is the cache coordination domain identifier.
	Scope            string        // Scope is the explicit invalidation scope inside the domain.
	AuthoritySource  string        // AuthoritySource is the canonical data source for rebuilds.
	ConsistencyModel string        // ConsistencyModel is the declared freshness model.
	MaxStale         time.Duration // MaxStale is the configured stale window.
	FailureStrategy  string        // FailureStrategy is the caller-visible degradation behavior.
	LocalRevision    int64         // LocalRevision is the latest revision consumed by this process.
	SharedRevision   int64         // SharedRevision is the latest shared coordination revision observed.
	LastSyncedAt     time.Time     // LastSyncedAt is the latest successful local synchronization time.
	RecentError      string        // RecentError is the most recent coordination failure.
	StaleSeconds     int64         // StaleSeconds is the elapsed time since LastSyncedAt.
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
	info.CacheCoordination = s.loadCacheCoordination(ctx)

	return info, nil
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
			LocalRevision:    item.LocalRevision,
			SharedRevision:   item.SharedRevision,
			LastSyncedAt:     item.LastSyncedAt,
			RecentError:      item.RecentError,
			StaleSeconds:     item.StaleSeconds,
		})
	}
	return diagnostics
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
			case "MySQL":
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
