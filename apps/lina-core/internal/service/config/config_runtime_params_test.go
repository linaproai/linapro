// This file verifies built-in runtime parameter validation and sys_config
// overrides for host config getters.

package config

import (
	"context"
	"testing"
	"time"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
)

func TestValidateRuntimeParamValue(t *testing.T) {
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
	}

	for _, testCase := range testCases {
		err := ValidateRuntimeParamValue(testCase.key, testCase.value)
		if testCase.shouldErr && err == nil {
			t.Fatalf("expected validation error for %s=%q", testCase.key, testCase.value)
		}
		if !testCase.shouldErr && err != nil {
			t.Fatalf("expected validation success for %s=%q, got %v", testCase.key, testCase.value, err)
		}
	}
}

func TestGetJwtPrefersRuntimeParamOverride(t *testing.T) {
	withRuntimeParamValue(t, RuntimeParamKeyJWTExpire, "12h")

	svc := New()
	cfg := svc.GetJwt(context.Background())

	if cfg.Expire != 12*time.Hour {
		t.Fatalf("expected runtime param jwt expire to be 12h, got %s", cfg.Expire)
	}
	if expire := svc.GetJwtExpire(context.Background()); expire != 12*time.Hour {
		t.Fatalf("expected runtime getter jwt expire to be 12h, got %s", expire)
	}
}

func TestGetSessionPrefersRuntimeParamTimeout(t *testing.T) {
	withRuntimeParamValue(t, RuntimeParamKeySessionTimeout, "2h")

	svc := New()
	cfg := svc.GetSession(context.Background())

	if cfg.Timeout != 2*time.Hour {
		t.Fatalf("expected runtime param session timeout to be 2h, got %s", cfg.Timeout)
	}
	if cfg.CleanupInterval <= 0 {
		t.Fatalf("expected cleanup interval to remain positive, got %s", cfg.CleanupInterval)
	}
	if timeout := svc.GetSessionTimeout(context.Background()); timeout != 2*time.Hour {
		t.Fatalf("expected runtime getter session timeout to be 2h, got %s", timeout)
	}
}

func TestGetUploadPrefersRuntimeParamMaxSize(t *testing.T) {
	withRuntimeParamValue(t, RuntimeParamKeyUploadMaxSize, "8")

	svc := New()
	cfg := svc.GetUpload(context.Background())

	if cfg.MaxSize != 8 {
		t.Fatalf("expected runtime param upload max size to be 8, got %d", cfg.MaxSize)
	}
	if maxSize := svc.GetUploadMaxSize(context.Background()); maxSize != 8 {
		t.Fatalf("expected runtime getter upload max size to be 8, got %d", maxSize)
	}
}

func TestGetLoginUsesRuntimeBlacklist(t *testing.T) {
	withRuntimeParamValue(t, RuntimeParamKeyLoginBlackIPList, "127.0.0.1;10.0.0.0/8")

	svc := New()
	cfg := svc.GetLogin(context.Background())

	if !cfg.IsBlacklisted("127.0.0.1") {
		t.Fatal("expected 127.0.0.1 to be blacklisted")
	}
	if !cfg.IsBlacklisted("10.1.2.3") {
		t.Fatal("expected 10.1.2.3 to match blacklisted CIDR")
	}
	if cfg.IsBlacklisted("192.168.1.10") {
		t.Fatal("expected 192.168.1.10 not to be blacklisted")
	}
	if !svc.IsLoginIPBlacklisted(context.Background(), "10.1.2.3") {
		t.Fatal("expected runtime blacklist getter to match 10.1.2.3")
	}
	if svc.IsLoginIPBlacklisted(context.Background(), "192.168.1.10") {
		t.Fatal("expected runtime blacklist getter not to match 192.168.1.10")
	}
}

func TestRuntimeParamSnapshotReloadsAfterRevisionChange(t *testing.T) {
	ctx := context.Background()
	withRuntimeParamValue(t, RuntimeParamKeyJWTExpire, "12h")
	clearRuntimeParamSnapshotCache(t, ctx)

	svc := New()
	if cfg := svc.GetJwt(ctx); cfg.Expire != 12*time.Hour {
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

	if cfg := svc.GetJwt(ctx); cfg.Expire != 12*time.Hour {
		t.Fatalf("expected cached jwt expire to remain 12h before revision bump, got %s", cfg.Expire)
	}

	markRuntimeParamChanged(t, ctx)

	if cfg := svc.GetJwt(ctx); cfg.Expire != 6*time.Hour {
		t.Fatalf("expected jwt expire to reload to 6h after revision bump, got %s", cfg.Expire)
	}
}

func TestSyncRuntimeParamSnapshotKeepsCachedValueWhenRevisionUnchanged(t *testing.T) {
	ctx := context.Background()
	withRuntimeParamValue(t, RuntimeParamKeyJWTExpire, "12h")
	clearRuntimeParamSnapshotCache(t, ctx)

	svc := New()
	if err := svc.SyncRuntimeParamSnapshot(ctx); err != nil {
		t.Fatalf("initial runtime param sync failed: %v", err)
	}
	if cfg := svc.GetJwt(ctx); cfg.Expire != 12*time.Hour {
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
	if cfg := svc.GetJwt(ctx); cfg.Expire != 12*time.Hour {
		t.Fatalf("expected cached jwt expire to remain 12h when revision is unchanged, got %s", cfg.Expire)
	}
}

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
	if cfg := svc.GetJwt(ctx); cfg.Expire != 6*time.Hour {
		t.Fatalf("expected jwt expire to reload to 6h after watcher sync, got %s", cfg.Expire)
	}
}

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
				Name:   original.Name,
				Key:    original.Key,
				Value:  original.Value,
				Remark: original.Remark,
			}).
			Update()
		if cleanupErr != nil {
			t.Fatalf("restore runtime param %s: %v", key, cleanupErr)
		}
		markRuntimeParamChanged(t, ctx)
	})
}

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
			Id:     original.Id,
			Name:   original.Name,
			Key:    original.Key,
			Value:  original.Value,
			Remark: original.Remark,
		}).Insert()
		if cleanupErr != nil {
			t.Fatalf("restore deleted runtime param %s: %v", key, cleanupErr)
		}
		markRuntimeParamChanged(t, ctx)
	})
}

func markRuntimeParamChanged(t *testing.T, ctx context.Context) {
	t.Helper()

	if err := New().MarkRuntimeParamsChanged(ctx); err != nil {
		t.Fatalf("mark runtime params changed: %v", err)
	}
}

func clearRuntimeParamSnapshotCache(t *testing.T, ctx context.Context) {
	t.Helper()

	if _, err := runtimeParamSnapshotCache.Remove(ctx, runtimeParamSnapshotCacheKey); err != nil {
		t.Fatalf("clear runtime param snapshot cache: %v", err)
	}
}

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
