// This file maps sysinfo service output into the v1 system-info response.

package sysinfo

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode"

	"lina-core/api/sysinfo/v1"
)

const (
	runDurationHoursMinutesSecondsKey = "systemInfo.runDuration.format.hoursMinutesSeconds"
	runDurationMinutesSecondsKey      = "systemInfo.runDuration.format.minutesSeconds"
	runDurationSecondsKey             = "systemInfo.runDuration.format.seconds"
)

// GetInfo returns system information
func (c *ControllerV1) GetInfo(ctx context.Context, req *v1.GetInfoReq) (res *v1.GetInfoRes, err error) {
	info, err := c.sysInfoSvc.GetInfo(ctx)
	if err != nil {
		return nil, err
	}

	res = &v1.GetInfoRes{
		Framework: v1.FrameworkInfo{
			Name:          info.Framework.Name,
			Version:       info.Framework.Version,
			Description:   c.i18nSvc.Translate(ctx, "framework.description", info.Framework.Description),
			Homepage:      info.Framework.Homepage,
			RepositoryURL: info.Framework.RepositoryURL,
			License:       info.Framework.License,
		},
		GoVersion:          info.GoVersion,
		GfVersion:          info.GfVersion,
		Os:                 info.Os,
		Arch:               info.Arch,
		DbVersion:          info.DbVersion,
		StartTime:          info.StartTime,
		RunDuration:        c.formatRunDuration(ctx, info.RunDurationSeconds),
		RunDurationSeconds: info.RunDurationSeconds,
	}

	// Map backend components
	for _, component := range info.BackendComponents {
		res.BackendComponents = append(res.BackendComponents, v1.ComponentInfo{
			Name:    component.Name,
			Version: component.Version,
			Url:     component.Url,
			Description: c.i18nSvc.Translate(
				ctx,
				"systemInfo.components.backend."+normalizeComponentKey(component.Name)+".description",
				component.Description,
			),
		})
	}

	// Map frontend components
	for _, component := range info.FrontendComponents {
		res.FrontendComponents = append(res.FrontendComponents, v1.ComponentInfo{
			Name:    component.Name,
			Version: component.Version,
			Url:     component.Url,
			Description: c.i18nSvc.Translate(
				ctx,
				"systemInfo.components.frontend."+normalizeComponentKey(component.Name)+".description",
				component.Description,
			),
		})
	}

	for _, item := range info.CacheCoordination {
		res.CacheCoordination = append(res.CacheCoordination, v1.CacheCoordinationInfo{
			Domain:           item.Domain,
			Scope:            item.Scope,
			AuthoritySource:  item.AuthoritySource,
			ConsistencyModel: item.ConsistencyModel,
			MaxStaleSeconds:  int64(item.MaxStale / time.Second),
			FailureStrategy:  item.FailureStrategy,
			LocalRevision:    item.LocalRevision,
			SharedRevision:   item.SharedRevision,
			LastSyncedAt:     formatOptionalTime(item.LastSyncedAt),
			RecentError:      item.RecentError,
			StaleSeconds:     item.StaleSeconds,
		})
	}

	return res, nil
}

// formatRunDuration returns the localized human-readable uptime string.
func (c *ControllerV1) formatRunDuration(ctx context.Context, totalSeconds int64) string {
	if totalSeconds < 0 {
		totalSeconds = 0
	}

	hours := totalSeconds / int64(time.Hour/time.Second)
	minutes := totalSeconds / int64(time.Minute/time.Second) % 60
	seconds := totalSeconds % 60
	if hours > 0 {
		template := c.i18nSvc.Translate(ctx, runDurationHoursMinutesSecondsKey, "%d hours %d minutes %d seconds")
		return fmt.Sprintf(template, hours, minutes, seconds)
	}
	if minutes > 0 {
		template := c.i18nSvc.Translate(ctx, runDurationMinutesSecondsKey, "%d minutes %d seconds")
		return fmt.Sprintf(template, minutes, seconds)
	}
	template := c.i18nSvc.Translate(ctx, runDurationSecondsKey, "%d seconds")
	return fmt.Sprintf(template, seconds)
}

// normalizeComponentKey converts a component display name into a stable i18n key segment.
func normalizeComponentKey(name string) string {
	trimmed := strings.TrimSpace(strings.ToLower(name))
	if trimmed == "" {
		return "unknown"
	}

	var builder strings.Builder
	lastWasHyphen := false
	for _, r := range trimmed {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			builder.WriteRune(r)
			lastWasHyphen = false
		case !lastWasHyphen:
			builder.WriteByte('-')
			lastWasHyphen = true
		}
	}

	normalized := strings.Trim(builder.String(), "-")
	if normalized == "" {
		return "unknown"
	}
	return normalized
}

// formatOptionalTime returns an empty value for unset diagnostic timestamps.
func formatOptionalTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02 15:04:05")
}
