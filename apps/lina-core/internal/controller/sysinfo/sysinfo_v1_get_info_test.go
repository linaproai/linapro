// This file verifies localized system-info response projections.

package sysinfo

import (
	"context"
	"testing"
	"time"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/os/gctx"

	"lina-core/internal/model"
	i18nsvc "lina-core/internal/service/i18n"
	sysinfosvc "lina-core/internal/service/sysinfo"
)

// fakeSysInfoService provides deterministic system-info data for controller
// response mapping tests.
type fakeSysInfoService struct {
	info *sysinfosvc.SystemInfo
}

// GetInfo returns the configured system-info payload.
func (f *fakeSysInfoService) GetInfo(_ context.Context) (*sysinfosvc.SystemInfo, error) {
	return f.info, nil
}

// TestFormatRunDurationUsesRuntimeLocale verifies uptime strings use runtime i18n resources.
func TestFormatRunDurationUsesRuntimeLocale(t *testing.T) {
	t.Parallel()

	controller := &ControllerV1{i18nSvc: i18nsvc.New()}

	testCases := []struct {
		name     string
		locale   string
		seconds  int64
		expected string
	}{
		{name: "default locale hours", locale: i18nsvc.DefaultLocale, seconds: 3661, expected: "1小时1分钟1秒"},
		{name: "traditional locale minutes", locale: "zh-TW", seconds: 125, expected: "2分鐘5秒"},
		{name: "english locale seconds", locale: i18nsvc.EnglishLocale, seconds: 42, expected: "42 seconds"},
		{name: "english locale hours", locale: i18nsvc.EnglishLocale, seconds: 7322, expected: "2 hours 2 minutes 2 seconds"},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.WithValue(
				context.Background(),
				gctx.StrKey("BizCtx"),
				&model.Context{Locale: testCase.locale},
			)
			if actual := controller.formatRunDuration(ctx, testCase.seconds); actual != testCase.expected {
				t.Fatalf("expected %q, got %q", testCase.expected, actual)
			}
		})
	}
}

// TestGetInfoMapsCacheCoordinationDiagnostics verifies cache coordination
// snapshots are included in the HTTP response payload.
func TestGetInfoMapsCacheCoordinationDiagnostics(t *testing.T) {
	ctx := context.Background()
	syncedAt := time.Date(2025, 1, 1, 8, 0, 0, 0, time.UTC)
	controller := &ControllerV1{
		i18nSvc: i18nsvc.New(),
		sysInfoSvc: &fakeSysInfoService{
			info: &sysinfosvc.SystemInfo{
				Framework: sysinfosvc.FrameworkInfo{Name: "LinaPro"},
				CacheCoordination: []sysinfosvc.CacheCoordinationInfo{
					{
						Domain:           "runtime-config",
						Scope:            "global",
						AuthoritySource:  "sys_config protected runtime parameters",
						ConsistencyModel: "shared-revision",
						MaxStale:         10 * time.Second,
						FailureStrategy:  "return-visible-error",
						LocalRevision:    3,
						SharedRevision:   4,
						LastSyncedAt:     syncedAt,
						RecentError:      "previous read failed",
						StaleSeconds:     2,
					},
				},
			},
		},
	}

	res, err := controller.GetInfo(ctx, nil)
	if err != nil {
		t.Fatalf("get system info failed: %v", err)
	}
	if len(res.CacheCoordination) != 1 {
		t.Fatalf("expected one cache coordination row, got %d", len(res.CacheCoordination))
	}
	item := res.CacheCoordination[0]
	if item.Domain != "runtime-config" ||
		item.Scope != "global" ||
		item.MaxStaleSeconds != 10 ||
		item.LocalRevision != 3 ||
		item.SharedRevision != 4 ||
		item.LastSyncedAt != "2025-01-01 08:00:00" ||
		item.RecentError != "previous read failed" ||
		item.StaleSeconds != 2 {
		t.Fatalf("unexpected cache coordination response row: %#v", item)
	}
}
