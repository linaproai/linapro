// This file verifies localized system-info response projections.

package sysinfo

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/os/gctx"
	_ "lina-core/pkg/dbdriver"

	"lina-core/internal/model"
	i18nsvc "lina-core/internal/service/i18n"
	sysinfosvc "lina-core/internal/service/sysinfo"
)

// fakeSysInfoService provides deterministic system-info data for controller
// response mapping tests.
type fakeSysInfoService struct {
	info *sysinfosvc.SystemInfo
}

// TestSysinfoAPIDocI18nContainsCoordinationFields verifies newly exposed
// diagnostics have dedicated apidoc i18n entries.
func TestSysinfoAPIDocI18nContainsCoordinationFields(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatalf("resolve lina-core root: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(root, "manifest", "i18n", "zh-CN", "apidoc", "core-api-sysinfo.json"))
	if err != nil {
		t.Fatalf("read sysinfo apidoc i18n: %v", err)
	}
	var catalog map[string]any
	if err = json.Unmarshal(content, &catalog); err != nil {
		t.Fatalf("parse sysinfo apidoc i18n: %v", err)
	}

	requiredPaths := []string{
		"core.api.sysinfo.v1.CoordinationInfo.fields.clusterEnabled.dc",
		"core.api.sysinfo.v1.CoordinationInfo.fields.backend.dc",
		"core.api.sysinfo.v1.CoordinationInfo.fields.redisHealthy.dc",
		"core.api.sysinfo.v1.CoordinationInfo.fields.nodeId.dc",
		"core.api.sysinfo.v1.CoordinationInfo.fields.primary.dc",
		"core.api.sysinfo.v1.CoordinationInfo.fields.lastSuccessAt.dc",
		"core.api.sysinfo.v1.CoordinationInfo.fields.lastError.dc",
		"core.api.sysinfo.v1.CacheCoordinationInfo.fields.backend.dc",
		"core.api.sysinfo.v1.CacheCoordinationInfo.fields.healthy.dc",
		"core.api.sysinfo.v1.CacheCoordinationInfo.fields.eventSubscriber.dc",
		"core.api.sysinfo.v1.CacheCoordinationInfo.fields.lastEventAt.dc",
		"core.api.sysinfo.v1.GetInfoRes.fields.coordination.dc",
	}
	for _, path := range requiredPaths {
		if value := lookupNestedString(catalog, path); value == "" {
			t.Fatalf("expected apidoc i18n value at %s", path)
		}
	}
}

// lookupNestedString returns a dotted-path string from a JSON object.
func lookupNestedString(root map[string]any, path string) string {
	var current any = root
	for _, part := range strings.Split(path, ".") {
		asMap, ok := current.(map[string]any)
		if !ok {
			return ""
		}
		current = asMap[part]
	}
	value, ok := current.(string)
	if !ok {
		return ""
	}
	return value
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
		{name: "default locale minutes", locale: i18nsvc.DefaultLocale, seconds: 125, expected: "2分钟5秒"},
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
				Coordination: sysinfosvc.CoordinationInfo{
					ClusterEnabled: true,
					Backend:        "redis",
					RedisHealthy:   true,
					NodeID:         "node-a",
					Primary:        true,
					LastSuccessAt:  syncedAt,
					LastError:      "",
				},
				CacheCoordination: []sysinfosvc.CacheCoordinationInfo{
					{
						Domain:           "runtime-config",
						Scope:            "global",
						AuthoritySource:  "sys_config protected runtime parameters",
						ConsistencyModel: "shared-revision",
						MaxStale:         10 * time.Second,
						FailureStrategy:  "return-visible-error",
						Backend:          "redis",
						Healthy:          true,
						LocalRevision:    3,
						SharedRevision:   4,
						LastSyncedAt:     syncedAt,
						EventSubscriber:  true,
						LastEventAt:      syncedAt.Add(time.Second),
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
	if !res.Coordination.ClusterEnabled ||
		res.Coordination.Backend != "redis" ||
		!res.Coordination.RedisHealthy ||
		res.Coordination.NodeId != "node-a" ||
		!res.Coordination.Primary ||
		res.Coordination.LastSuccessAt != "2025-01-01 08:00:00" ||
		res.Coordination.LastError != "" {
		t.Fatalf("unexpected coordination diagnostics: %#v", res.Coordination)
	}
	item := res.CacheCoordination[0]
	if item.Domain != "runtime-config" ||
		item.Scope != "global" ||
		item.MaxStaleSeconds != 10 ||
		item.Backend != "redis" ||
		!item.Healthy ||
		item.LocalRevision != 3 ||
		item.SharedRevision != 4 ||
		item.LastSyncedAt != "2025-01-01 08:00:00" ||
		!item.EventSubscriber ||
		item.LastEventAt != "2025-01-01 08:00:01" ||
		item.RecentError != "previous read failed" ||
		item.StaleSeconds != 2 {
		t.Fatalf("unexpected cache coordination response row: %#v", item)
	}
}
