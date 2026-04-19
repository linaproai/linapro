// This file verifies runtime authentication behaviors driven by managed
// sys_config parameters.

package auth

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	hostconfig "lina-core/internal/service/config"
)

// TestLoginRejectsBlacklistedIP verifies managed login IP blacklist settings
// are enforced before user lookup.
func TestLoginRejectsBlacklistedIP(t *testing.T) {
	withRuntimeParamValue(t, hostconfig.RuntimeParamKeyLoginBlackIPList, "127.0.0.1")

	username := fmt.Sprintf("blacklist-test-%s", t.Name())
	ctx := newRequestContext(t, "127.0.0.1:18080")

	_, err := New().Login(ctx, LoginInput{
		Username: username,
		Password: "ignored",
	})
	if err == nil {
		t.Fatal("expected blacklisted login attempt to fail")
	}
	if err.Error() != "登录IP已被禁止" {
		t.Fatalf("expected blacklisted login error, got %v", err)
	}

	t.Cleanup(func() {
		if _, cleanupErr := dao.SysLoginLog.Ctx(context.Background()).
			Where(do.SysLoginLog{UserName: username}).
			Delete(); cleanupErr != nil {
			t.Fatalf("cleanup login logs for %s: %v", username, cleanupErr)
		}
	})
}

// newRequestContext builds one request-backed context carrying the supplied
// remote address for auth service tests.
func newRequestContext(t *testing.T, remoteAddr string) context.Context {
	t.Helper()

	httpReq, err := http.NewRequest(http.MethodPost, "http://localhost/api/v1/auth/login", nil)
	if err != nil {
		t.Fatalf("build http request: %v", err)
	}
	httpReq.RemoteAddr = remoteAddr
	httpReq.Header.Set("User-Agent", "runtime-param-test")

	req := &ghttp.Request{Request: httpReq}
	return req.Context()
}

// withRuntimeParamValue temporarily overrides one protected runtime parameter
// and restores the original sys_config record during cleanup.
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

// markRuntimeParamChanged bumps the runtime-parameter revision for tests after
// direct sys_config mutations.
func markRuntimeParamChanged(t *testing.T, ctx context.Context) {
	t.Helper()

	if err := hostconfig.New().MarkRuntimeParamsChanged(ctx); err != nil {
		t.Fatalf("mark runtime params changed: %v", err)
	}
}

// queryRuntimeParam loads one sys_config record by protected runtime-parameter key.
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
