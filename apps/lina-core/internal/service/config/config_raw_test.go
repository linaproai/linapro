// This file verifies raw host config reads across sys_config snapshots and
// static GoFrame config fallback paths.

package config

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	_ "lina-core/pkg/dbdriver"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/cachecoord"
	"lina-core/internal/service/coordination"
	"lina-core/internal/service/datascope"
	"lina-core/pkg/bizerr"
)

// TestGetRawReadsCustomSysConfigKey verifies source-plugin HostConfig reads are
// backed by the data-driven sys_config snapshot instead of a hard-coded key list.
func TestGetRawReadsCustomSysConfigKey(t *testing.T) {
	ctx := context.Background()
	key := uniqueRawConfigKey(t, "custom.feature.limit")
	insertRawSysConfig(t, ctx, datascope.PlatformTenantID, key, "100")
	markRuntimeParamChanged(t, ctx)

	value, err := New().(*serviceImpl).GetRaw(ctx, key)
	if err != nil {
		t.Fatalf("get custom sys_config raw value: %v", err)
	}
	if value == nil || value.String() != "100" {
		t.Fatalf("expected custom sys_config value 100, got %#v", value)
	}
}

// TestGetRawPrefersTenantSysConfigOverride verifies tenant context reads prefer
// tenant-owned sys_config rows over platform fallback rows.
func TestGetRawPrefersTenantSysConfigOverride(t *testing.T) {
	ctx := context.Background()
	key := uniqueRawConfigKey(t, "custom.feature.tenant")
	insertRawSysConfig(t, ctx, datascope.PlatformTenantID, key, "platform")
	insertRawSysConfig(t, ctx, 87, key, "tenant")
	markRuntimeParamChanged(t, ctx)

	value, err := New().(*serviceImpl).GetRaw(datascope.WithTenantScope(ctx, 87), key)
	if err != nil {
		t.Fatalf("get tenant custom sys_config raw value: %v", err)
	}
	if value == nil || value.String() != "tenant" {
		t.Fatalf("expected tenant sys_config override, got %#v", value)
	}
}

// TestGetRawReloadsCustomSysConfigAfterRevisionChange verifies revision bumps
// refresh custom sys_config keys, including deletion fallback behavior.
func TestGetRawReloadsCustomSysConfigAfterRevisionChange(t *testing.T) {
	var (
		ctx        = context.Background()
		key        = uniqueRawConfigKey(t, "custom.feature.reload")
		insertedID = insertRawSysConfig(t, ctx, datascope.PlatformTenantID, key, "100")
	)
	markRuntimeParamChanged(t, ctx)

	svc := New().(*serviceImpl)
	value, err := svc.GetRaw(ctx, key)
	if err != nil {
		t.Fatalf("get initial custom sys_config raw value: %v", err)
	}
	if value == nil || value.String() != "100" {
		t.Fatalf("expected initial value 100, got %#v", value)
	}

	if _, err = dao.SysConfig.Ctx(ctx).
		Where(do.SysConfig{Id: insertedID}).
		Data(do.SysConfig{Value: "200"}).
		Update(); err != nil {
		t.Fatalf("update custom sys_config value: %v", err)
	}
	markRuntimeParamChanged(t, ctx)
	value, err = svc.GetRaw(ctx, key)
	if err != nil {
		t.Fatalf("get updated custom sys_config raw value: %v", err)
	}
	if value == nil || value.String() != "200" {
		t.Fatalf("expected updated value 200, got %#v", value)
	}

	if _, err = dao.SysConfig.Ctx(ctx).Where(do.SysConfig{Id: insertedID}).Delete(); err != nil {
		t.Fatalf("delete custom sys_config value: %v", err)
	}
	markRuntimeParamChanged(t, ctx)
	value, err = svc.GetRaw(ctx, key)
	if err != nil {
		t.Fatalf("get deleted custom sys_config raw value: %v", err)
	}
	if value != nil && !value.IsNil() {
		t.Fatalf("expected deleted sys_config key to be absent, got %#v", value)
	}
}

// TestGetRawFallsBackToStaticConfigWhenSysConfigMissing verifies static host
// configuration remains available when no sys_config row exists.
func TestGetRawFallsBackToStaticConfigWhenSysConfigMissing(t *testing.T) {
	adapter, err := gcfg.NewAdapterContent(`
workspace:
  basePath: "/ops"
`)
	if err != nil {
		t.Fatalf("create static config adapter: %v", err)
	}
	originalAdapter := g.Cfg().GetAdapter()
	g.Cfg().SetAdapter(adapter)
	t.Cleanup(func() {
		g.Cfg().SetAdapter(originalAdapter)
	})

	value, err := New().(*serviceImpl).GetRaw(context.Background(), "workspace.basePath")
	if err != nil {
		t.Fatalf("get static config fallback: %v", err)
	}
	if value == nil || value.String() != "/ops" {
		t.Fatalf("expected static workspace.basePath /ops, got %#v", value)
	}
}

// uniqueRawConfigKey returns a key that avoids collisions across repeated test runs.
func uniqueRawConfigKey(t *testing.T, prefix string) string {
	t.Helper()
	return fmt.Sprintf("%s.%d", prefix, time.Now().UnixNano())
}

// insertRawSysConfig inserts one sys_config row and registers cleanup.
func insertRawSysConfig(t *testing.T, ctx context.Context, tenantID int, key string, value string) int64 {
	t.Helper()
	id, err := dao.SysConfig.Ctx(ctx).Data(do.SysConfig{
		TenantId: tenantID,
		Name:     key,
		Key:      key,
		Value:    value,
		Remark:   "raw host config test",
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert raw sys_config %s: %v", key, err)
	}
	t.Cleanup(func() {
		if _, cleanupErr := dao.SysConfig.Ctx(ctx).Unscoped().Where(do.SysConfig{Id: id}).Delete(); cleanupErr != nil {
			t.Fatalf("cleanup raw sys_config %s: %v", key, cleanupErr)
		}
		markRuntimeParamChanged(t, ctx)
	})
	return id
}

// TestNewCacheCoordRuntimeParamRevisionControllerSelectsByClusterMode verifies the
// constructor selects the local or clustered revision strategy correctly.
func TestNewCacheCoordRuntimeParamRevisionControllerSelectsByClusterMode(t *testing.T) {
	if _, ok := newCacheCoordRuntimeParamRevisionController(false).(*localRuntimeParamRevisionController); !ok {
		t.Fatal("expected single-node mode to use local runtime-param revision controller")
	}

	controller, ok := newCacheCoordRuntimeParamRevisionController(true).(*clusterRuntimeParamRevisionController)
	if !ok {
		t.Fatal("expected cluster mode to use shared runtime-param revision controller")
	}
	if controller.cacheCoordSvc == nil {
		t.Fatal("expected clustered runtime-param revision controller to use cachecoord")
	}
}

// TestRuntimeParamValueValidation verifies built-in runtime parameter validators
// accept valid values and reject malformed ones.
func TestRuntimeParamValueValidation(t *testing.T) {
	testCases := []struct {
		key       string
		value     string
		shouldErr bool
	}{
		{key: RuntimeParamKeyJWTExpire, value: "24h"},
		{key: RuntimeParamKeyJWTExpire, value: "bad", shouldErr: true},
		{key: RuntimeParamKeySessionTimeout, value: "30m"},
		{key: RuntimeParamKeySessionTimeout, value: "0s", shouldErr: true},
		{key: RuntimeParamKeyUploadMaxSize, value: "10"},
		{key: RuntimeParamKeyUploadMaxSize, value: "0", shouldErr: true},
		{key: RuntimeParamKeyLoginBlackIPList, value: "127.0.0.1;10.0.0.0/8"},
		{key: RuntimeParamKeyLoginBlackIPList, value: "invalid-ip", shouldErr: true},
		{key: RuntimeParamKeyLogRetentionDays, value: "90"},
		{key: RuntimeParamKeyLogRetentionDays, value: "0", shouldErr: true},
		{key: RuntimeParamKeyLogRetentionDays, value: "-1", shouldErr: true},
		{key: RuntimeParamKeyLogRetentionDays, value: "bad", shouldErr: true},
		{key: RuntimeParamKeyCronShellEnabled, value: "true"},
		{key: RuntimeParamKeyCronShellEnabled, value: "yes", shouldErr: true},
		{key: RuntimeParamKeyCronLogRetention, value: `{"mode":"days","value":30}`},
		{key: RuntimeParamKeyCronLogRetention, value: `{"mode":"count","value":200}`},
		{key: RuntimeParamKeyCronLogRetention, value: `{"mode":"none","value":0}`},
		{key: RuntimeParamKeyCronLogRetention, value: `{"mode":"none","value":-1}`, shouldErr: true},
		{key: RuntimeParamKeyCronLogRetention, value: `{"mode":"days","value":0}`, shouldErr: true},
		{key: RuntimeParamKeyCronLogRetention, value: `{"mode":"unknown","value":1}`, shouldErr: true},
	}

	for _, testCase := range testCases {
		err := validateRuntimeParamValue(testCase.key, testCase.value)
		if testCase.shouldErr && err == nil {
			t.Fatalf("expected validation error for %s=%q", testCase.key, testCase.value)
		}
		if !testCase.shouldErr && err != nil {
			t.Fatalf("expected validation success for %s=%q, got %v", testCase.key, testCase.value, err)
		}
	}
}

// TestRuntimeParamSpecCopiesAreDetached verifies callers cannot mutate the shared
// built-in runtime-parameter specification slice.
func TestRuntimeParamSpecCopiesAreDetached(t *testing.T) {
	specs := runtimeParamSpecsCopy()
	if len(specs) == 0 {
		t.Fatal("expected runtime param specs to be present")
	}

	uploadSpec, ok := lookupRuntimeParamSpec(RuntimeParamKeyUploadMaxSize)
	if !ok {
		t.Fatal("expected upload-size runtime param spec to be present")
	}
	if uploadSpec.DefaultValue != "100" {
		t.Fatalf("expected upload-size runtime default to be 100, got %q", uploadSpec.DefaultValue)
	}
	original := runtimeParamSpecs[0].DefaultValue
	specs[0].DefaultValue = "mutated"
	if runtimeParamSpecs[0].DefaultValue != original {
		t.Fatal("expected runtimeParamSpecsCopy to return a detached copy")
	}
}

// TestRuntimeParamSpecKeysUseSystemNamespace verifies all built-in runtime
// parameters managed through sys_config stay under the host system namespace.
func TestRuntimeParamSpecKeysUseSystemNamespace(t *testing.T) {
	for _, spec := range runtimeParamSpecsCopy() {
		if !strings.HasPrefix(spec.Key, "sys.") {
			t.Fatalf("expected built-in runtime parameter %q to use sys. prefix", spec.Key)
		}
	}
}

// TestRuntimeParamSpecKeysExcludeLoggerTraceIDSwitch verifies the TraceID switch
// is no longer exposed as one managed sys_config runtime parameter.
func TestRuntimeParamSpecKeysExcludeLoggerTraceIDSwitch(t *testing.T) {
	if _, ok := lookupRuntimeParamSpec("sys.logger.traceID.enabled"); ok {
		t.Fatal("expected logger TraceID switch to be removed from managed runtime params")
	}
	if isManagedRuntimeParamKey("sys.logger.traceID.enabled") {
		t.Fatal("expected logger TraceID switch not to be treated as managed")
	}
}

// TestRuntimeParamSnapshotReloadsAfterRevisionChange verifies direct reads
// rebuild the cached snapshot after the protected-config revision changes.
func TestRuntimeParamSnapshotReloadsAfterRevisionChange(t *testing.T) {
	ctx := context.Background()
	withRuntimeParamValue(t, RuntimeParamKeyJWTExpire, "12h")
	clearRuntimeParamSnapshotCache(t, ctx)

	svc := New()
	cfg, err := svc.GetJwt(ctx)
	if err != nil {
		t.Fatalf("get initial jwt config: %v", err)
	}
	if cfg.Expire != 12*time.Hour {
		t.Fatalf("expected initial cached jwt expire to be 12h, got %s", cfg.Expire)
	}

	original, err := queryRuntimeParam(ctx, RuntimeParamKeyJWTExpire)
	if err != nil {
		t.Fatalf("query jwt runtime param: %v", err)
	}
	if original == nil {
		t.Fatal("expected jwt runtime param to exist")
	}

	_, err = dao.SysConfig.Ctx(ctx).
		Unscoped().
		Where(do.SysConfig{Id: original.Id}).
		Data(do.SysConfig{Value: "6h"}).
		Update()
	if err != nil {
		t.Fatalf("update jwt runtime param without revision bump: %v", err)
	}
	t.Cleanup(func() {
		_, cleanupErr := dao.SysConfig.Ctx(ctx).
			Unscoped().
			Where(do.SysConfig{Id: original.Id}).
			Data(do.SysConfig{Value: original.Value}).
			Update()
		if cleanupErr != nil {
			t.Fatalf("restore jwt runtime param after snapshot reload test: %v", cleanupErr)
		}
		markRuntimeParamChanged(t, ctx)
	})

	cfg, err = svc.GetJwt(ctx)
	if err != nil {
		t.Fatalf("get cached jwt config: %v", err)
	}
	if cfg.Expire != 12*time.Hour {
		t.Fatalf("expected cached jwt expire to remain 12h before revision bump, got %s", cfg.Expire)
	}

	markRuntimeParamChanged(t, ctx)

	cfg, err = svc.GetJwt(ctx)
	if err != nil {
		t.Fatalf("get reloaded jwt config: %v", err)
	}
	if cfg.Expire != 6*time.Hour {
		t.Fatalf("expected jwt expire to reload to 6h after revision bump, got %s", cfg.Expire)
	}
}

// TestSyncRuntimeParamSnapshotKeepsCachedValueWhenRevisionUnchanged verifies
// watcher sync preserves the local snapshot when the revision does not change.
func TestSyncRuntimeParamSnapshotKeepsCachedValueWhenRevisionUnchanged(t *testing.T) {
	ctx := context.Background()
	withRuntimeParamValue(t, RuntimeParamKeyJWTExpire, "12h")
	clearRuntimeParamSnapshotCache(t, ctx)

	svc := New()
	if err := svc.SyncRuntimeParamSnapshot(ctx); err != nil {
		t.Fatalf("initial runtime param sync failed: %v", err)
	}
	cfg, err := svc.GetJwt(ctx)
	if err != nil {
		t.Fatalf("get synced jwt config: %v", err)
	}
	if cfg.Expire != 12*time.Hour {
		t.Fatalf("expected synced jwt expire to be 12h, got %s", cfg.Expire)
	}

	original, err := queryRuntimeParam(ctx, RuntimeParamKeyJWTExpire)
	if err != nil {
		t.Fatalf("query jwt runtime param: %v", err)
	}
	if original == nil {
		t.Fatal("expected jwt runtime param to exist")
	}

	_, err = dao.SysConfig.Ctx(ctx).
		Unscoped().
		Where(do.SysConfig{Id: original.Id}).
		Data(do.SysConfig{Value: "6h"}).
		Update()
	if err != nil {
		t.Fatalf("update jwt runtime param without revision bump: %v", err)
	}
	t.Cleanup(func() {
		_, cleanupErr := dao.SysConfig.Ctx(ctx).
			Unscoped().
			Where(do.SysConfig{Id: original.Id}).
			Data(do.SysConfig{Value: original.Value}).
			Update()
		if cleanupErr != nil {
			t.Fatalf("restore jwt runtime param after unchanged revision test: %v", cleanupErr)
		}
		markRuntimeParamChanged(t, ctx)
	})

	if err = svc.SyncRuntimeParamSnapshot(ctx); err != nil {
		t.Fatalf("runtime param sync with unchanged revision failed: %v", err)
	}
	cfg, err = svc.GetJwt(ctx)
	if err != nil {
		t.Fatalf("get cached jwt config after unchanged sync: %v", err)
	}
	if cfg.Expire != 12*time.Hour {
		t.Fatalf("expected cached jwt expire to remain 12h when revision is unchanged, got %s", cfg.Expire)
	}
}

// TestSyncRuntimeParamSnapshotReloadsAfterRevisionChange verifies watcher sync
// reloads the local snapshot after the shared revision advances.
func TestSyncRuntimeParamSnapshotReloadsAfterRevisionChange(t *testing.T) {
	ctx := context.Background()
	withRuntimeParamValue(t, RuntimeParamKeyJWTExpire, "12h")
	clearRuntimeParamSnapshotCache(t, ctx)

	svc := New()
	if err := svc.SyncRuntimeParamSnapshot(ctx); err != nil {
		t.Fatalf("initial runtime param sync failed: %v", err)
	}

	original, err := queryRuntimeParam(ctx, RuntimeParamKeyJWTExpire)
	if err != nil {
		t.Fatalf("query jwt runtime param: %v", err)
	}
	if original == nil {
		t.Fatal("expected jwt runtime param to exist")
	}

	_, err = dao.SysConfig.Ctx(ctx).
		Unscoped().
		Where(do.SysConfig{Id: original.Id}).
		Data(do.SysConfig{Value: "6h"}).
		Update()
	if err != nil {
		t.Fatalf("update jwt runtime param before revision sync: %v", err)
	}
	t.Cleanup(func() {
		_, cleanupErr := dao.SysConfig.Ctx(ctx).
			Unscoped().
			Where(do.SysConfig{Id: original.Id}).
			Data(do.SysConfig{Value: original.Value}).
			Update()
		if cleanupErr != nil {
			t.Fatalf("restore jwt runtime param after revision sync test: %v", cleanupErr)
		}
		markRuntimeParamChanged(t, ctx)
	})

	markRuntimeParamChanged(t, ctx)

	if err = svc.SyncRuntimeParamSnapshot(ctx); err != nil {
		t.Fatalf("runtime param sync after revision bump failed: %v", err)
	}
	cfg, err := svc.GetJwt(ctx)
	if err != nil {
		t.Fatalf("get reloaded jwt config after sync: %v", err)
	}
	if cfg.Expire != 6*time.Hour {
		t.Fatalf("expected jwt expire to reload to 6h after watcher sync, got %s", cfg.Expire)
	}
}

// TestSingleNodeRuntimeParamSnapshotStaysLocal verifies single-node mode avoids
// cachecoord traffic while still invalidating local snapshots.
func TestSingleNodeRuntimeParamSnapshotStaysLocal(t *testing.T) {
	ctx := context.Background()
	withRuntimeParamValue(t, RuntimeParamKeyJWTExpire, "12h")

	svc := New().(*serviceImpl)
	resetRuntimeParamCacheTestState(t)
	svc.runtimeParamRevisionCtrl = &localRuntimeParamRevisionController{}

	if err := svc.SyncRuntimeParamSnapshot(ctx); err != nil {
		t.Fatalf("single-node runtime param sync failed: %v", err)
	}
	cfg, err := svc.GetJwt(ctx)
	if err != nil {
		t.Fatalf("get single-node jwt config: %v", err)
	}
	if cfg.Expire != 12*time.Hour {
		t.Fatalf("expected initial jwt expire 12h, got %s", cfg.Expire)
	}

	original, err := queryRuntimeParam(ctx, RuntimeParamKeyJWTExpire)
	if err != nil {
		t.Fatalf("query jwt runtime param: %v", err)
	}
	if original == nil {
		t.Fatal("expected jwt runtime param to exist")
	}

	_, err = dao.SysConfig.Ctx(ctx).
		Unscoped().
		Where(do.SysConfig{Id: original.Id}).
		Data(do.SysConfig{Value: "6h"}).
		Update()
	if err != nil {
		t.Fatalf("update jwt runtime param before local invalidation: %v", err)
	}
	t.Cleanup(func() {
		_, cleanupErr := dao.SysConfig.Ctx(ctx).
			Unscoped().
			Where(do.SysConfig{Id: original.Id}).
			Data(do.SysConfig{Value: original.Value}).
			Update()
		if cleanupErr != nil {
			t.Fatalf("restore jwt runtime param after single-node test: %v", cleanupErr)
		}
		resetRuntimeParamCacheTestState(t)
		markRuntimeParamChanged(t, ctx)
	})

	cfg, err = svc.GetJwt(ctx)
	if err != nil {
		t.Fatalf("get cached single-node jwt config: %v", err)
	}
	if cfg.Expire != 12*time.Hour {
		t.Fatalf("expected cached jwt expire to stay 12h before local invalidation, got %s", cfg.Expire)
	}

	if err = svc.MarkRuntimeParamsChanged(ctx); err != nil {
		t.Fatalf("mark runtime params changed in single-node mode: %v", err)
	}
	cfg, err = svc.GetJwt(ctx)
	if err != nil {
		t.Fatalf("get reloaded single-node jwt config: %v", err)
	}
	if cfg.Expire != 6*time.Hour {
		t.Fatalf("expected jwt expire to reload to 6h after local invalidation, got %s", cfg.Expire)
	}
}

// withRuntimeParamValue writes one runtime parameter override for a test case
// and restores the previous database state afterward.
func withRuntimeParamValue(t *testing.T, key string, value string) {
	t.Helper()

	ctx := context.Background()
	original, err := queryRuntimeParam(ctx, key)
	if err != nil {
		t.Fatalf("query runtime param %s: %v", key, err)
	}

	if original == nil {
		_, err = dao.SysConfig.Ctx(ctx).Data(do.SysConfig{
			Name:   key,
			Key:    key,
			Value:  value,
			Remark: "test override",
		}).Insert()
		if err != nil {
			t.Fatalf("insert runtime param %s: %v", key, err)
		}
		markRuntimeParamChanged(t, ctx)
		t.Cleanup(func() {
			if _, cleanupErr := dao.SysConfig.Ctx(ctx).Unscoped().Where(do.SysConfig{Key: key}).Delete(); cleanupErr != nil {
				t.Fatalf("cleanup runtime param %s: %v", key, cleanupErr)
			}
			markRuntimeParamChanged(t, ctx)
		})
		return
	}

	_, err = dao.SysConfig.Ctx(ctx).
		Unscoped().
		Where(do.SysConfig{Id: original.Id}).
		Data(do.SysConfig{Value: value}).
		Update()
	if err != nil {
		t.Fatalf("update runtime param %s: %v", key, err)
	}
	markRuntimeParamChanged(t, ctx)
	t.Cleanup(func() {
		_, cleanupErr := dao.SysConfig.Ctx(ctx).
			Unscoped().
			Where(do.SysConfig{Id: original.Id}).
			Data(do.SysConfig{
				Name:      original.Name,
				Key:       original.Key,
				Value:     original.Value,
				IsBuiltin: original.IsBuiltin,
				Remark:    original.Remark,
			}).
			Update()
		if cleanupErr != nil {
			t.Fatalf("restore runtime param %s: %v", key, cleanupErr)
		}
		markRuntimeParamChanged(t, ctx)
	})
}

// withRuntimeParamAbsent removes one runtime parameter row for a test case and
// restores it afterward when necessary.
func withRuntimeParamAbsent(t *testing.T, key string) {
	t.Helper()

	ctx := context.Background()
	original, err := queryRuntimeParam(ctx, key)
	if err != nil {
		t.Fatalf("query runtime param %s: %v", key, err)
	}
	if original == nil {
		return
	}

	_, err = dao.SysConfig.Ctx(ctx).
		Unscoped().
		Where(do.SysConfig{Id: original.Id}).
		Delete()
	if err != nil {
		t.Fatalf("delete runtime param %s: %v", key, err)
	}
	markRuntimeParamChanged(t, ctx)

	t.Cleanup(func() {
		_, cleanupErr := dao.SysConfig.Ctx(ctx).Data(do.SysConfig{
			Name:      original.Name,
			Key:       original.Key,
			Value:     original.Value,
			IsBuiltin: original.IsBuiltin,
			Remark:    original.Remark,
		}).Insert()
		if cleanupErr != nil {
			t.Fatalf("restore deleted runtime param %s: %v", key, cleanupErr)
		}
		markRuntimeParamChanged(t, ctx)
	})
}

// withCachedRuntimeParamValue injects one process-local runtime snapshot value
// so tests can exercise override logic without touching sys_config.
func withCachedRuntimeParamValue(t *testing.T, key string, value string) {
	t.Helper()

	snapshot := &runtimeParamSnapshot{
		values:         map[string]string{key: value},
		durationValues: make(map[string]time.Duration),
		int64Values:    make(map[string]int64),
		parseErrors:    make(map[string]error),
	}
	switch key {
	case RuntimeParamKeyJWTExpire, RuntimeParamKeySessionTimeout:
		parsed, err := validatePositiveDurationValue(key, value)
		if err != nil {
			snapshot.parseErrors[key] = err
		} else {
			snapshot.durationValues[key] = parsed
		}
	case RuntimeParamKeyUploadMaxSize, RuntimeParamKeyLogRetentionDays:
		parsed, err := validatePositiveInt64Value(key, value)
		if err != nil {
			snapshot.parseErrors[key] = err
		} else {
			snapshot.int64Values[key] = parsed
		}
	}

	withCachedRuntimeParamSnapshot(t, snapshot)
}

// withCachedRuntimeParamParseError injects one runtime snapshot parse error so
// tests can exercise read-side fallback behavior.
func withCachedRuntimeParamParseError(t *testing.T, key string, parseErr error) {
	t.Helper()

	withCachedRuntimeParamSnapshot(t, &runtimeParamSnapshot{
		values:         map[string]string{key: "bad"},
		durationValues: make(map[string]time.Duration),
		int64Values:    make(map[string]int64),
		parseErrors:    map[string]error{key: parseErr},
	})
}

// withCachedRuntimeParamSnapshot injects one process-local runtime snapshot so
// tests can exercise fallback and override logic without shared sys_config state.
func withCachedRuntimeParamSnapshot(t *testing.T, snapshot *runtimeParamSnapshot) {
	t.Helper()

	ctx := context.Background()
	resetRuntimeParamCacheTestState(t)
	storeLocalRuntimeParamRevision(1)

	cached := &cachedRuntimeParamSnapshot{
		Revision:    1,
		RefreshedAt: time.Now(),
		Snapshot:    snapshot,
	}
	if cached.Snapshot == nil {
		cached.Snapshot = &runtimeParamSnapshot{}
	}
	if err := runtimeParamSnapshotCache.Set(
		ctx,
		scopedRuntimeParamSnapshotCacheKey(ctx),
		cached, runtimeParamSnapshotCacheTTL,
	); err != nil {
		t.Fatalf("seed runtime param snapshot cache: %v", err)
	}
}

// markRuntimeParamChanged bumps the runtime-parameter revision for test setup changes.
func markRuntimeParamChanged(t *testing.T, ctx context.Context) {
	t.Helper()

	if err := New().MarkRuntimeParamsChanged(ctx); err != nil {
		t.Fatalf("mark runtime params changed: %v", err)
	}
}

// clearRuntimeParamSnapshotCache clears the process-local runtime snapshot cache.
func clearRuntimeParamSnapshotCache(t *testing.T, ctx context.Context) {
	t.Helper()

	if _, err := runtimeParamSnapshotCache.Remove(ctx, scopedRuntimeParamSnapshotCacheKey(ctx)); err != nil {
		t.Fatalf("clear runtime param snapshot cache: %v", err)
	}
}

// resetRuntimeParamCacheTestState resets revision and snapshot cache state
// before and after a test case.
func resetRuntimeParamCacheTestState(t *testing.T) {
	t.Helper()

	ctx := context.Background()
	clearLocalRuntimeParamRevision()
	if _, err := runtimeParamSnapshotCache.Remove(ctx, scopedRuntimeParamSnapshotCacheKey(ctx)); err != nil {
		t.Fatalf("reset runtime param snapshot cache: %v", err)
	}
	t.Cleanup(func() {
		clearLocalRuntimeParamRevision()
		if _, err := runtimeParamSnapshotCache.Remove(ctx, scopedRuntimeParamSnapshotCacheKey(ctx)); err != nil {
			t.Fatalf("cleanup runtime param snapshot cache: %v", err)
		}
	})
}

// queryRuntimeParam loads one runtime parameter row directly from sys_config.
func queryRuntimeParam(ctx context.Context, key string) (*entity.SysConfig, error) {
	var runtimeParam *entity.SysConfig
	err := dao.SysConfig.Ctx(ctx).
		Unscoped().
		Where(do.SysConfig{Key: key}).
		Scan(&runtimeParam)
	if err != nil {
		return nil, err
	}
	return runtimeParam, nil
}

// fakeClusterRevisionCacheCoordService provides deterministic cachecoord behavior
// for clustered runtime-parameter revision tests.
type fakeClusterRevisionCacheCoordService struct {
	currentRevision int64
	currentErr      error
	currentCalls    int32
	markRevision    int64
	markErr         error
	markCalls       int32
}

// ConfigureDomain is a no-op because these tests configure domain metadata elsewhere.
func (f *fakeClusterRevisionCacheCoordService) ConfigureDomain(_ cachecoord.DomainSpec) error {
	return nil
}

// MarkChanged returns the configured changed revision or error.
func (f *fakeClusterRevisionCacheCoordService) MarkChanged(
	_ context.Context,
	_ cachecoord.Domain,
	_ cachecoord.Scope,
	_ cachecoord.ChangeReason,
) (int64, error) {
	atomic.AddInt32(&f.markCalls, 1)
	if f.markErr != nil {
		return 0, f.markErr
	}
	return f.markRevision, nil
}

// MarkTenantChanged returns the same configured revision as the global change
// path because runtime-parameter tests only verify revision coordination.
func (f *fakeClusterRevisionCacheCoordService) MarkTenantChanged(
	ctx context.Context,
	domain cachecoord.Domain,
	scope cachecoord.Scope,
	_ cachecoord.InvalidationScope,
	reason cachecoord.ChangeReason,
) (int64, error) {
	return f.MarkChanged(ctx, domain, scope, reason)
}

// EnsureFresh runs the refresher against the configured current revision.
func (f *fakeClusterRevisionCacheCoordService) EnsureFresh(
	ctx context.Context,
	domain cachecoord.Domain,
	scope cachecoord.Scope,
	refresher cachecoord.Refresher,
) (int64, error) {
	revision, err := f.CurrentRevision(ctx, domain, scope)
	if err != nil {
		return 0, err
	}
	if refresher != nil {
		if err = refresher(ctx, revision); err != nil {
			return 0, err
		}
	}
	return revision, nil
}

// CurrentRevision returns the configured revision value or the configured error.
func (f *fakeClusterRevisionCacheCoordService) CurrentRevision(
	_ context.Context,
	_ cachecoord.Domain,
	_ cachecoord.Scope,
) (int64, error) {
	atomic.AddInt32(&f.currentCalls, 1)
	if f.currentErr != nil {
		return 0, f.currentErr
	}
	return f.currentRevision, nil
}

// Snapshot is unused by runtime-parameter revision tests.
func (f *fakeClusterRevisionCacheCoordService) Snapshot(_ context.Context) ([]cachecoord.SnapshotItem, error) {
	return nil, nil
}

// fakeRuntimeParamRevisionController provides deterministic revision behavior
// for service-level cache helper tests.
type fakeRuntimeParamRevisionController struct {
	currentRevision int64
	syncRevision    int64
	markRevision    int64
	currentErr      error
	syncErr         error
	markErr         error
	markCalls       int32
}

// CurrentRevision returns the configured current revision or error.
func (f *fakeRuntimeParamRevisionController) CurrentRevision(_ context.Context) (int64, error) {
	if f.currentErr != nil {
		return 0, f.currentErr
	}
	return f.currentRevision, nil
}

// SyncRevision returns the configured synchronized revision or error.
func (f *fakeRuntimeParamRevisionController) SyncRevision(_ context.Context) (int64, error) {
	if f.syncErr != nil {
		return 0, f.syncErr
	}
	return f.syncRevision, nil
}

// MarkChanged returns the configured changed revision or error.
func (f *fakeRuntimeParamRevisionController) MarkChanged(_ context.Context) (int64, error) {
	atomic.AddInt32(&f.markCalls, 1)
	if f.markErr != nil {
		return 0, f.markErr
	}
	return f.markRevision, nil
}

// TestClusterRuntimeParamRevisionControllerCurrentRevisionEnsuresFreshValue verifies
// request-path revision checks consult cachecoord instead of indefinitely
// trusting the process-local copy.
func TestClusterRuntimeParamRevisionControllerCurrentRevisionEnsuresFreshValue(t *testing.T) {
	clearLocalRuntimeParamRevision()
	t.Cleanup(clearLocalRuntimeParamRevision)

	fakeCoord := &fakeClusterRevisionCacheCoordService{currentRevision: 7}
	controller := &clusterRuntimeParamRevisionController{cacheCoordSvc: fakeCoord}

	revision, err := controller.CurrentRevision(context.Background())
	if err != nil {
		t.Fatalf("load shared revision: %v", err)
	}
	if revision != 7 {
		t.Fatalf("expected shared revision 7, got %d", revision)
	}

	revision, err = controller.CurrentRevision(context.Background())
	if err != nil {
		t.Fatalf("reload shared revision: %v", err)
	}
	if revision != 7 {
		t.Fatalf("expected refreshed revision 7, got %d", revision)
	}
	if calls := atomic.LoadInt32(&fakeCoord.currentCalls); calls != 2 {
		t.Fatalf("expected two cachecoord read calls, got %d", calls)
	}
}

// TestClusterRuntimeParamRevisionControllerSyncRevisionRefreshesLocalState verifies
// explicit sync always refreshes from cachecoord and replaces the local revision.
func TestClusterRuntimeParamRevisionControllerSyncRevisionRefreshesLocalState(t *testing.T) {
	clearLocalRuntimeParamRevision()
	storeLocalRuntimeParamRevision(3)
	t.Cleanup(clearLocalRuntimeParamRevision)

	fakeCoord := &fakeClusterRevisionCacheCoordService{currentRevision: 9}
	controller := &clusterRuntimeParamRevisionController{cacheCoordSvc: fakeCoord}

	revision, err := controller.SyncRevision(context.Background())
	if err != nil {
		t.Fatalf("sync shared revision: %v", err)
	}
	if revision != 9 {
		t.Fatalf("expected synced revision 9, got %d", revision)
	}
	if local, ok := getLocalRuntimeParamRevision(); !ok || local != 9 {
		t.Fatalf("expected local revision to be updated to 9, got value=%d ok=%t", local, ok)
	}
}

// TestClusterRuntimeParamRevisionControllerMarkChangedStoresReturnedRevision verifies
// successful shared increments are mirrored locally for the writing node.
func TestClusterRuntimeParamRevisionControllerMarkChangedStoresReturnedRevision(t *testing.T) {
	clearLocalRuntimeParamRevision()
	t.Cleanup(clearLocalRuntimeParamRevision)

	fakeCoord := &fakeClusterRevisionCacheCoordService{markRevision: 11}
	controller := &clusterRuntimeParamRevisionController{cacheCoordSvc: fakeCoord}

	revision, err := controller.MarkChanged(context.Background())
	if err != nil {
		t.Fatalf("increment shared revision: %v", err)
	}
	if revision != 11 {
		t.Fatalf("expected incremented revision 11, got %d", revision)
	}
	if local, ok := getLocalRuntimeParamRevision(); !ok || local != 11 {
		t.Fatalf("expected local revision to be updated to 11, got value=%d ok=%t", local, ok)
	}
	if calls := atomic.LoadInt32(&fakeCoord.markCalls); calls != 1 {
		t.Fatalf("expected one cachecoord publish call, got %d", calls)
	}
}

// TestClusterRuntimeParamRevisionControllerPropagatesCacheCoordErrors verifies
// cachecoord read and publish failures surface to callers.
func TestClusterRuntimeParamRevisionControllerPropagatesCacheCoordErrors(t *testing.T) {
	clearLocalRuntimeParamRevision()
	t.Cleanup(clearLocalRuntimeParamRevision)

	var (
		readErr        = errors.New("read revision failed")
		readCoord      = &fakeClusterRevisionCacheCoordService{currentErr: readErr}
		readController = &clusterRuntimeParamRevisionController{cacheCoordSvc: readCoord}
	)
	if _, err := readController.CurrentRevision(context.Background()); !errors.Is(err, readErr) {
		t.Fatalf("expected CurrentRevision error %v, got %v", readErr, err)
	}
	if _, err := readController.SyncRevision(context.Background()); !errors.Is(err, readErr) {
		t.Fatalf("expected SyncRevision error %v, got %v", readErr, err)
	}

	var (
		writeErr        = errors.New("increment revision failed")
		writeCoord      = &fakeClusterRevisionCacheCoordService{markErr: writeErr}
		writeController = &clusterRuntimeParamRevisionController{cacheCoordSvc: writeCoord}
	)
	if _, err := writeController.MarkChanged(context.Background()); !errors.Is(err, writeErr) {
		t.Fatalf("expected MarkChanged error %v, got %v", writeErr, err)
	}
}

// TestClusterRuntimeParamRevisionControllerConsumesCrossInstanceRevision
// verifies a second controller instance can observe a revision published by
// another clustered writer through the persistent coordination row.
func TestClusterRuntimeParamRevisionControllerConsumesCrossInstanceRevision(t *testing.T) {
	ctx := context.Background()
	clearLocalRuntimeParamRevision()
	t.Cleanup(clearLocalRuntimeParamRevision)
	coordSvc := coordination.NewMemory(nil)

	publisher := &clusterRuntimeParamRevisionController{
		cacheCoordSvc: cachecoord.NewWithCoordination(cachecoord.NewStaticTopology(true), coordSvc),
	}
	consumer := &clusterRuntimeParamRevisionController{
		cacheCoordSvc: cachecoord.NewWithCoordination(cachecoord.NewStaticTopology(true), coordSvc),
	}

	revision, err := publisher.MarkChanged(ctx)
	if err != nil {
		t.Fatalf("publish runtime-param revision failed: %v", err)
	}
	clearLocalRuntimeParamRevision()

	observed, err := consumer.SyncRevision(ctx)
	if err != nil {
		t.Fatalf("consume runtime-param revision from second controller failed: %v", err)
	}
	if observed != revision {
		t.Fatalf("expected consumer revision %d, got %d", revision, observed)
	}
	if local, ok := getLocalRuntimeParamRevision(); !ok || local != revision {
		t.Fatalf("expected local runtime-param revision %d, got value=%d ok=%t", revision, local, ok)
	}
}

// TestNotifyRuntimeParamsChangedSwallowsRevisionErrors verifies the helper is
// best-effort and never panics when revision publication fails.
func TestNotifyRuntimeParamsChangedSwallowsRevisionErrors(t *testing.T) {
	svc := &serviceImpl{
		runtimeParamRevisionCtrl: &fakeRuntimeParamRevisionController{markErr: errors.New("boom")},
	}

	svc.NotifyRuntimeParamsChanged(context.Background())
}

// TestRuntimeParamSnapshotReturnsControllerUnavailableError verifies malformed
// service construction no longer relies on recover to hide missing revision wiring.
func TestRuntimeParamSnapshotReturnsControllerUnavailableError(t *testing.T) {
	ctx := context.Background()
	resetRuntimeParamCacheTestState(t)

	snapshot, err := (&serviceImpl{}).getRuntimeParamSnapshot(ctx)
	if err == nil {
		t.Fatal("expected missing runtime-param revision controller to return an error")
	}
	if snapshot != nil {
		t.Fatal("expected no snapshot when revision controller is unavailable")
	}
	if !bizerr.Is(err, CodeConfigRuntimeParamRevisionUnavailable) {
		t.Fatalf("expected runtime-param revision unavailable error, got %v", err)
	}
}

// TestRuntimeParamSnapshotPropagatesRevisionErrorWithCachedSnapshot verifies a
// stale process-local snapshot is not reused when freshness cannot be confirmed.
func TestRuntimeParamSnapshotPropagatesRevisionErrorWithCachedSnapshot(t *testing.T) {
	ctx := context.Background()
	resetRuntimeParamCacheTestState(t)

	currentErr := errors.New("current revision failed")
	svc := &serviceImpl{
		runtimeParamRevisionCtrl: &fakeRuntimeParamRevisionController{currentErr: currentErr},
	}
	cached := &cachedRuntimeParamSnapshot{
		Revision:    3,
		RefreshedAt: time.Now(),
		Snapshot: &runtimeParamSnapshot{
			values:         map[string]string{RuntimeParamKeyJWTExpire: "12h"},
			durationValues: map[string]time.Duration{RuntimeParamKeyJWTExpire: 12 * time.Hour},
			int64Values:    map[string]int64{},
			parseErrors:    map[string]error{},
		},
	}
	if err := runtimeParamSnapshotCache.Set(
		ctx,
		scopedRuntimeParamSnapshotCacheKey(ctx),
		cached,
		runtimeParamSnapshotCacheTTL,
	); err != nil {
		t.Fatalf("seed runtime param snapshot cache: %v", err)
	}

	snapshot, err := svc.getRuntimeParamSnapshot(ctx)
	if !errors.Is(err, currentErr) {
		t.Fatalf("expected revision error %v, got %v", currentErr, err)
	}
	if snapshot != nil {
		t.Fatal("expected no snapshot when revision freshness check fails")
	}
}

// TestRuntimeParamSnapshotWatcherInterval verifies the package helper returns
// the configured synchronization interval constant.
func TestRuntimeParamSnapshotWatcherInterval(t *testing.T) {
	if interval := runtimeParamSnapshotSyncIntervalForTest(); interval != 10*time.Second {
		t.Fatalf("expected runtime snapshot sync interval 10s, got %s", interval)
	}
}

// TestGetCachedRuntimeParamSnapshotRemovesInvalidEntries verifies malformed or
// stale cache entries are dropped before later reads try to reuse them.
func TestGetCachedRuntimeParamSnapshotRemovesInvalidEntries(t *testing.T) {
	ctx := context.Background()
	resetRuntimeParamCacheTestState(t)

	svc := New().(*serviceImpl)
	if err := runtimeParamSnapshotCache.Set(ctx, scopedRuntimeParamSnapshotCacheKey(ctx), "invalid", runtimeParamSnapshotCacheTTL); err != nil {
		t.Fatalf("seed invalid runtime snapshot cache entry: %v", err)
	}
	if cached := svc.getCachedRuntimeParamSnapshot(ctx, 1); cached != nil {
		t.Fatal("expected invalid cache entry to be rejected")
	}
	if cachedVar, err := runtimeParamSnapshotCache.Get(ctx, scopedRuntimeParamSnapshotCacheKey(ctx)); err != nil {
		t.Fatalf("get invalid cache entry after removal: %v", err)
	} else if cachedVar != nil {
		t.Fatal("expected invalid cache entry to be removed")
	}

	valid := &cachedRuntimeParamSnapshot{
		Revision:    2,
		RefreshedAt: time.Now(),
		Snapshot: &runtimeParamSnapshot{
			values:         map[string]string{},
			durationValues: map[string]time.Duration{},
			int64Values:    map[string]int64{},
			parseErrors:    map[string]error{},
		},
	}
	if err := runtimeParamSnapshotCache.Set(ctx, scopedRuntimeParamSnapshotCacheKey(ctx), valid, runtimeParamSnapshotCacheTTL); err != nil {
		t.Fatalf("seed stale runtime snapshot cache entry: %v", err)
	}
	if cached := svc.getCachedRuntimeParamSnapshot(ctx, 1); cached != nil {
		t.Fatal("expected revision-mismatched cache entry to be rejected")
	}
}

// TestExtractCachedRuntimeParamSnapshotRejectsBrokenValues verifies defensive
// cache decoding ignores wrong types and nil snapshots.
func TestExtractCachedRuntimeParamSnapshotRejectsBrokenValues(t *testing.T) {
	if cached := extractCachedRuntimeParamSnapshot("invalid"); cached != nil {
		t.Fatal("expected string cache payload to be rejected")
	}
	if cached := extractCachedRuntimeParamSnapshot(&cachedRuntimeParamSnapshot{}); cached != nil {
		t.Fatal("expected cache payload without snapshot to be rejected")
	}
}
