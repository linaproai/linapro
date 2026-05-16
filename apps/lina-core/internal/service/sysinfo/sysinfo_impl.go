// This file contains sysinfo aggregation logic for runtime metadata,
// coordination diagnostics, uptime formatting, and database version lookup.

package sysinfo

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/gogf/gf/v2"
	"github.com/gogf/gf/v2/frame/g"

	"lina-core/internal/service/config"
	"lina-core/internal/service/coordination"
	"lina-core/pkg/dialect"
	"lina-core/pkg/logger"
)

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
