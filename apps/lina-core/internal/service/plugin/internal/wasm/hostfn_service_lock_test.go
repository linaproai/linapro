// This file tests lock host service dispatch, ticket validation, and authorization.

package wasm

import (
	"context"
	"strings"
	"testing"

	"github.com/gogf/gf/v2/frame/g"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/pkg/pluginbridge"
)

const createPluginLockerTableSQL = `
CREATE TABLE IF NOT EXISTS sys_locker (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    name VARCHAR(64) NOT NULL COMMENT '锁名称，唯一标识',
    reason VARCHAR(255) DEFAULT '' COMMENT '获取锁的原因',
    holder VARCHAR(64) DEFAULT '' COMMENT '锁持有者标识（节点名）',
    expire_time DATETIME NOT NULL COMMENT '锁过期时间',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY uk_name (name),
    INDEX idx_expire_time (expire_time)
) ENGINE=MEMORY DEFAULT CHARSET=utf8mb4 COMMENT='分布式锁表';
`

func TestHandleHostServiceInvokeLockLifecycle(t *testing.T) {
	ctx := context.Background()
	ensurePluginLockerTable(t, ctx)

	pluginID := "test-plugin-lock"
	lockName := "orders-sync"
	cleanupPluginLock(t, ctx, buildPluginLockName(pluginID, lockName))
	t.Cleanup(func() {
		cleanupPluginLock(t, ctx, buildPluginLockName(pluginID, lockName))
	})

	hcc := newLockHostCallContext(pluginID, lockName)

	acquireResponse := invokeLockHostService(
		t,
		hcc,
		pluginbridge.HostServiceMethodLockAcquire,
		lockName,
		pluginbridge.MarshalHostServiceLockAcquireRequest(&pluginbridge.HostServiceLockAcquireRequest{LeaseMillis: 5000}),
	)
	if acquireResponse.Status != pluginbridge.HostCallStatusSuccess {
		t.Fatalf("acquire: expected success, got status=%d payload=%s", acquireResponse.Status, string(acquireResponse.Payload))
	}
	acquirePayload, err := pluginbridge.UnmarshalHostServiceLockAcquireResponse(acquireResponse.Payload)
	if err != nil {
		t.Fatalf("acquire payload decode failed: %v", err)
	}
	if !acquirePayload.Acquired || strings.TrimSpace(acquirePayload.Ticket) == "" {
		t.Fatalf("acquire payload: got %#v", acquirePayload)
	}

	duplicateAcquireResponse := invokeLockHostService(
		t,
		hcc,
		pluginbridge.HostServiceMethodLockAcquire,
		lockName,
		pluginbridge.MarshalHostServiceLockAcquireRequest(&pluginbridge.HostServiceLockAcquireRequest{LeaseMillis: 5000}),
	)
	if duplicateAcquireResponse.Status != pluginbridge.HostCallStatusSuccess {
		t.Fatalf("duplicate acquire: expected success envelope, got status=%d payload=%s", duplicateAcquireResponse.Status, string(duplicateAcquireResponse.Payload))
	}
	duplicateAcquirePayload, err := pluginbridge.UnmarshalHostServiceLockAcquireResponse(duplicateAcquireResponse.Payload)
	if err != nil {
		t.Fatalf("duplicate acquire payload decode failed: %v", err)
	}
	if duplicateAcquirePayload.Acquired {
		t.Fatalf("expected duplicate acquire to be rejected by lock holder, got %#v", duplicateAcquirePayload)
	}

	renewResponse := invokeLockHostService(
		t,
		hcc,
		pluginbridge.HostServiceMethodLockRenew,
		lockName,
		pluginbridge.MarshalHostServiceLockRenewRequest(&pluginbridge.HostServiceLockRenewRequest{Ticket: acquirePayload.Ticket}),
	)
	if renewResponse.Status != pluginbridge.HostCallStatusSuccess {
		t.Fatalf("renew: expected success, got status=%d payload=%s", renewResponse.Status, string(renewResponse.Payload))
	}
	renewPayload, err := pluginbridge.UnmarshalHostServiceLockRenewResponse(renewResponse.Payload)
	if err != nil {
		t.Fatalf("renew payload decode failed: %v", err)
	}
	if strings.TrimSpace(renewPayload.ExpireAt) == "" {
		t.Fatalf("renew payload: got %#v", renewPayload)
	}

	releaseResponse := invokeLockHostService(
		t,
		hcc,
		pluginbridge.HostServiceMethodLockRelease,
		lockName,
		pluginbridge.MarshalHostServiceLockReleaseRequest(&pluginbridge.HostServiceLockReleaseRequest{Ticket: acquirePayload.Ticket}),
	)
	if releaseResponse.Status != pluginbridge.HostCallStatusSuccess {
		t.Fatalf("release: expected success, got status=%d payload=%s", releaseResponse.Status, string(releaseResponse.Payload))
	}

	reacquireResponse := invokeLockHostService(
		t,
		hcc,
		pluginbridge.HostServiceMethodLockAcquire,
		lockName,
		pluginbridge.MarshalHostServiceLockAcquireRequest(&pluginbridge.HostServiceLockAcquireRequest{LeaseMillis: 5000}),
	)
	if reacquireResponse.Status != pluginbridge.HostCallStatusSuccess {
		t.Fatalf("reacquire: expected success, got status=%d payload=%s", reacquireResponse.Status, string(reacquireResponse.Payload))
	}
	reacquirePayload, err := pluginbridge.UnmarshalHostServiceLockAcquireResponse(reacquireResponse.Payload)
	if err != nil {
		t.Fatalf("reacquire payload decode failed: %v", err)
	}
	if !reacquirePayload.Acquired {
		t.Fatalf("expected released lock to be acquirable again, got %#v", reacquirePayload)
	}
}

func TestHandleHostServiceInvokeLockRejectsTicketMismatch(t *testing.T) {
	ctx := context.Background()
	ensurePluginLockerTable(t, ctx)

	pluginID := "test-plugin-lock-mismatch"
	lockName := "orders-sync"
	otherLockName := "inventory-sync"
	cleanupPluginLock(t, ctx, buildPluginLockName(pluginID, lockName))
	cleanupPluginLock(t, ctx, buildPluginLockName(pluginID, otherLockName))
	t.Cleanup(func() {
		cleanupPluginLock(t, ctx, buildPluginLockName(pluginID, lockName))
		cleanupPluginLock(t, ctx, buildPluginLockName(pluginID, otherLockName))
	})

	hcc := newLockHostCallContext(pluginID, lockName, otherLockName)
	acquireResponse := invokeLockHostService(
		t,
		hcc,
		pluginbridge.HostServiceMethodLockAcquire,
		lockName,
		pluginbridge.MarshalHostServiceLockAcquireRequest(&pluginbridge.HostServiceLockAcquireRequest{LeaseMillis: 5000}),
	)
	if acquireResponse.Status != pluginbridge.HostCallStatusSuccess {
		t.Fatalf("acquire: expected success, got status=%d payload=%s", acquireResponse.Status, string(acquireResponse.Payload))
	}
	acquirePayload, err := pluginbridge.UnmarshalHostServiceLockAcquireResponse(acquireResponse.Payload)
	if err != nil {
		t.Fatalf("acquire payload decode failed: %v", err)
	}

	mismatchResponse := invokeLockHostService(
		t,
		hcc,
		pluginbridge.HostServiceMethodLockRenew,
		otherLockName,
		pluginbridge.MarshalHostServiceLockRenewRequest(&pluginbridge.HostServiceLockRenewRequest{Ticket: acquirePayload.Ticket}),
	)
	if mismatchResponse.Status != pluginbridge.HostCallStatusInvalidRequest {
		t.Fatalf("expected invalid request for mismatched ticket, got status=%d payload=%s", mismatchResponse.Status, string(mismatchResponse.Payload))
	}
}

func TestHandleHostServiceInvokeLockRejectsUnauthorizedResource(t *testing.T) {
	hcc := newLockHostCallContext("test-plugin-lock-denied", "orders-sync")
	response := invokeLockHostService(
		t,
		hcc,
		pluginbridge.HostServiceMethodLockAcquire,
		"inventory-sync",
		pluginbridge.MarshalHostServiceLockAcquireRequest(&pluginbridge.HostServiceLockAcquireRequest{LeaseMillis: 5000}),
	)
	if response.Status != pluginbridge.HostCallStatusCapabilityDenied {
		t.Fatalf("expected capability denied for unauthorized lock name, got status=%d payload=%s", response.Status, string(response.Payload))
	}
}

func ensurePluginLockerTable(t *testing.T, ctx context.Context) {
	t.Helper()
	if _, err := g.DB().Exec(ctx, createPluginLockerTableSQL); err != nil {
		t.Fatalf("expected sys_locker table to be created, got error: %v", err)
	}
}

func cleanupPluginLock(t *testing.T, ctx context.Context, lockName string) {
	t.Helper()
	if _, err := dao.SysLocker.Ctx(ctx).Where(do.SysLocker{Name: lockName}).Delete(); err != nil {
		t.Fatalf("failed to cleanup plugin lock %s: %v", lockName, err)
	}
}

func buildPluginLockName(pluginID string, lockName string) string {
	return "plugin:" + pluginID + ":" + lockName
}

func newLockHostCallContext(pluginID string, lockNames ...string) *hostCallContext {
	resources := make([]*pluginbridge.HostServiceResourceSpec, 0, len(lockNames))
	for _, lockName := range lockNames {
		resources = append(resources, &pluginbridge.HostServiceResourceSpec{Ref: lockName})
	}
	return &hostCallContext{
		pluginID: pluginID,
		capabilities: map[string]struct{}{
			pluginbridge.CapabilityLock: {},
		},
		hostServices: []*pluginbridge.HostServiceSpec{{
			Service: pluginbridge.HostServiceLock,
			Methods: []string{
				pluginbridge.HostServiceMethodLockAcquire,
				pluginbridge.HostServiceMethodLockRelease,
				pluginbridge.HostServiceMethodLockRenew,
			},
			Resources: resources,
		}},
	}
}

func invokeLockHostService(
	t *testing.T,
	hcc *hostCallContext,
	method string,
	lockName string,
	payload []byte,
) *pluginbridge.HostCallResponseEnvelope {
	t.Helper()

	request := &pluginbridge.HostServiceRequestEnvelope{
		Service:     pluginbridge.HostServiceLock,
		Method:      method,
		ResourceRef: lockName,
		Payload:     payload,
	}
	return handleHostServiceInvoke(
		context.Background(),
		hcc,
		pluginbridge.MarshalHostServiceRequestEnvelope(request),
	)
}
