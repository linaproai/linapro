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

	"lina-core/internal/service/config"
	"lina-core/pkg/logger"
)

const (
	componentSectionBackend  = "backend"
	componentSectionFrontend = "frontend"
)

// Service defines the sysinfo service contract.
type Service interface {
	// GetInfo returns system runtime information.
	GetInfo(ctx context.Context) (*SystemInfo, error)
}

var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	startTime time.Time      // startTime stores the host service boot time.
	configSvc config.Service // configSvc loads embedded metadata and runtime config values.
}

// New creates and returns a new Service instance.
func New() Service {
	return &serviceImpl{
		startTime: time.Now(),
		configSvc: config.New(),
	}
}

// SystemInfo holds the system runtime information.
type SystemInfo struct {
	GoVersion          string          // GoVersion is the active Go runtime version.
	GfVersion          string          // GfVersion is the active GoFrame runtime version.
	Os                 string          // Os is the operating system name.
	Arch               string          // Arch is the runtime architecture.
	DbVersion          string          // DbVersion is the database server version string.
	StartTime          string          // StartTime is the host start timestamp.
	RunDuration        string          // RunDuration is the formatted uptime string.
	BackendComponents  []ComponentInfo // BackendComponents lists backend technology cards.
	FrontendComponents []ComponentInfo // FrontendComponents lists frontend technology cards.
}

// ComponentInfo holds component display information.
type ComponentInfo struct {
	Name        string // Name is the component display name.
	Version     string // Version is the resolved component version label.
	Url         string // Url is the component homepage.
	Description string // Description is the short component summary.
}

// GetInfo returns system runtime information.
func (s *serviceImpl) GetInfo(ctx context.Context) (*SystemInfo, error) {
	info := &SystemInfo{
		GoVersion: runtime.Version(),
		GfVersion: gf.VERSION,
		Os:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		StartTime: s.startTime.Format("2006-01-02 15:04:05"),
	}

	duration := time.Since(s.startTime)
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60
	if hours > 0 {
		info.RunDuration = fmt.Sprintf("%d小时%d分钟%d秒", hours, minutes, seconds)
	} else if minutes > 0 {
		info.RunDuration = fmt.Sprintf("%d分钟%d秒", minutes, seconds)
	} else {
		info.RunDuration = fmt.Sprintf("%d秒", seconds)
	}

	dbVersion, err := s.getDbVersion(ctx)
	if err != nil {
		logger.Warningf(ctx, "Failed to get database version: %v", err)
		info.DbVersion = "unknown"
	} else {
		info.DbVersion = dbVersion
	}

	info.BackendComponents = s.loadComponents(ctx, componentSectionBackend, dbVersion)
	info.FrontendComponents = s.loadComponents(ctx, componentSectionFrontend, "")

	return info, nil
}

// loadComponents reads version-page component metadata from metadata.yaml.
func (s *serviceImpl) loadComponents(ctx context.Context, sectionKey string, dbVersion string) []ComponentInfo {
	metadata := s.configSvc.GetMetadata(ctx)

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
	result, err := g.DB().GetValue(ctx, "SELECT VERSION()")
	if err != nil {
		return "", err
	}
	return result.String(), nil
}
