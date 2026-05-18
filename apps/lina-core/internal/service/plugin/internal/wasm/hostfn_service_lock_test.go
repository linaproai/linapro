// This file tests lock host service dispatch, ticket validation, and authorization.

package wasm

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/service/coordination"
	"lina-core/internal/service/hostlock"
	"lina-core/internal/service/locker"
	"lina-core/pkg/dialect"
	"lina-core/pkg/pluginbridge"
)

// trackingLockService records lock operations while returning deterministic
// tickets for shared-instance wiring tests.
type trackingLockService struct {
	acquireCalls int
	renewCalls   int
	releaseCalls int
	lastInput    hostlock.AcquireInput
}

// Acquire records one lock acquisition request.
func (s *trackingLockService) Acquire(_ context.Context, in hostlock.AcquireInput) (*hostlock.AcquireOutput, error) {
	s.acquireCalls++
	s.lastInput = in
	return &hostlock.AcquireOutput{
		Acquired: true,
		Ticket:   "shared-lock-ticket",
		ExpireAt: gtime.Now().Add(5 * time.Second),
	}, nil
}

// Renew records one lock renewal request.
func (s *trackingLockService) Renew(context.Context, string, int64, string, string) (*gtime.Time, error) {
	s.renewCalls++
	expireAt := gtime.Now().Add(5 * time.Second)
	return expireAt, nil
}

// Release records one lock release request.
func (s *trackingLockService) Release(context.Context, string, int64, string, string) error {
	s.releaseCalls++
	return nil
}

// createPluginLockerTableSQL prepares the governed lock table for tests.
const createPluginLockerTableSQL = `
CREATE TABLE IF NOT EXISTS sys_locker (
    id          INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name        VARCHAR(128) NOT NULL,
    reason      VARCHAR(255) NOT NULL DEFAULT '',
    holder      VARCHAR(128) NOT NULL DEFAULT '',
    expire_time TIMESTAMP NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_locker_name ON sys_locker (name);
CREATE INDEX IF NOT EXISTS idx_sys_locker_expire_time ON sys_locker (expire_time);
`

// TestHandleHostServiceInvokeLockLifecycle verifies acquire, renew, and release lock flows.
func TestHandleHostServiceInvokeLockLifecycle(t *testing.T) {
	ctx := context.Background()
	ensurePluginLockerTable(t, ctx)
	configureDefaultLockHostService(t)

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

// TestHandleHostServiceInvokeLockRejectsTicketMismatch verifies mismatched tickets are rejected.
func TestHandleHostServiceInvokeLockRejectsTicketMismatch(t *testing.T) {
	ctx := context.Background()
	ensurePluginLockerTable(t, ctx)
	configureDefaultLockHostService(t)

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

// TestHandleHostServiceInvokeLockRejectsUnauthorizedResource verifies unauthorized lock names are rejected.
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

// TestHandleHostServiceInvokeLockUsesConfiguredSharedService verifies lock
// host service dispatch reuses the explicitly configured shared instance.
func TestHandleHostServiceInvokeLockUsesConfiguredSharedService(t *testing.T) {
	lockSvc := &trackingLockService{}
	previousLockSvc := lockHostService
	if err := ConfigureLockHostService(lockSvc); err != nil {
		t.Fatalf("configure lock host service failed: %v", err)
	}
	t.Cleanup(func() {
		lockHostService = previousLockSvc
	})

	hcc := newTenantLockHostCallContext("test-plugin-lock-shared", 88, "orders-sync")
	response := invokeLockHostService(
		t,
		hcc,
		pluginbridge.HostServiceMethodLockAcquire,
		"orders-sync",
		pluginbridge.MarshalHostServiceLockAcquireRequest(&pluginbridge.HostServiceLockAcquireRequest{LeaseMillis: 5000}),
	)
	if response.Status != pluginbridge.HostCallStatusSuccess {
		t.Fatalf("acquire through shared lock: expected success, got status=%d payload=%s", response.Status, string(response.Payload))
	}
	payload, err := pluginbridge.UnmarshalHostServiceLockAcquireResponse(response.Payload)
	if err != nil {
		t.Fatalf("decode shared lock acquire: %v", err)
	}
	if payload.Ticket != "shared-lock-ticket" {
		t.Fatalf("expected shared lock ticket, got %#v", payload)
	}
	if lockSvc.acquireCalls != 1 {
		t.Fatalf("expected shared lock service to receive one acquire, got %d", lockSvc.acquireCalls)
	}
	if lockSvc.lastInput.PluginID != hcc.pluginID || lockSvc.lastInput.TenantID != 88 || lockSvc.lastInput.ResourceRef != "orders-sync" {
		t.Fatalf("expected plugin/tenant/resource to be forwarded, got %#v", lockSvc.lastInput)
	}
}

// TestHandleHostServiceInvokeLockUsesCoordinationAndTenantIsolation verifies
// Wasm plugin locks use coordination locking and tenant-scoped lock names.
func TestHandleHostServiceInvokeLockUsesCoordinationAndTenantIsolation(t *testing.T) {
	coordSvc := coordination.NewMemory(nil)
	locker.ConfigureCoordination(coordSvc)
	previousLockSvc := lockHostService
	lockSvc, err := hostlock.New(locker.New())
	if err != nil {
		t.Fatalf("create host lock service failed: %v", err)
	}
	if err = ConfigureLockHostService(lockSvc); err != nil {
		t.Fatalf("configure host lock service failed: %v", err)
	}
	t.Cleanup(func() {
		locker.ConfigureCoordination(nil)
		lockHostService = previousLockSvc
	})

	lockName := "orders-sync"
	tenantOne := newTenantLockHostCallContext("test-plugin-lock-tenant", 11, lockName)
	tenantTwo := newTenantLockHostCallContext("test-plugin-lock-tenant", 22, lockName)
	otherPlugin := newTenantLockHostCallContext("test-plugin-lock-other", 11, lockName)

	tenantOneTicket := acquireLockTicket(t, tenantOne, lockName)
	tenantTwoTicket := acquireLockTicket(t, tenantTwo, lockName)
	otherPluginTicket := acquireLockTicket(t, otherPlugin, lockName)
	if tenantOneTicket == tenantTwoTicket || tenantOneTicket == otherPluginTicket {
		t.Fatal("expected isolated coordination lock tickets")
	}

	duplicateResponse := invokeLockHostService(
		t,
		tenantOne,
		pluginbridge.HostServiceMethodLockAcquire,
		lockName,
		pluginbridge.MarshalHostServiceLockAcquireRequest(&pluginbridge.HostServiceLockAcquireRequest{LeaseMillis: 5000}),
	)
	if duplicateResponse.Status != pluginbridge.HostCallStatusSuccess {
		t.Fatalf("duplicate acquire expected success envelope, got status=%d payload=%s", duplicateResponse.Status, string(duplicateResponse.Payload))
	}
	duplicatePayload, err := pluginbridge.UnmarshalHostServiceLockAcquireResponse(duplicateResponse.Payload)
	if err != nil {
		t.Fatalf("decode duplicate acquire payload: %v", err)
	}
	if duplicatePayload.Acquired {
		t.Fatalf("expected duplicate same plugin/tenant lock acquire to miss, got %#v", duplicatePayload)
	}

	wrongRelease := invokeLockHostService(
		t,
		tenantOne,
		pluginbridge.HostServiceMethodLockRelease,
		lockName,
		pluginbridge.MarshalHostServiceLockReleaseRequest(&pluginbridge.HostServiceLockReleaseRequest{Ticket: tenantTwoTicket}),
	)
	if wrongRelease.Status != pluginbridge.HostCallStatusInvalidRequest {
		t.Fatalf("expected cross-tenant release to fail, got status=%d payload=%s", wrongRelease.Status, string(wrongRelease.Payload))
	}

	releaseLockTicket(t, tenantOne, lockName, tenantOneTicket)
	reacquiredTicket := acquireLockTicket(t, tenantOne, lockName)
	if reacquiredTicket == "" {
		t.Fatal("expected reacquired ticket")
	}
}

// TestConfigureLockHostServiceRejectsNil verifies missing runtime lock
// injection returns an error instead of silently constructing an isolated backend.
func TestConfigureLockHostServiceRejectsNil(t *testing.T) {
	if err := ConfigureLockHostService(nil); err == nil {
		t.Fatal("expected nil lock host service to return an error")
	}
}

// configureDefaultLockHostService wires tests to the same locker-backed host
// lock service shape used by startup and restores the previous package state.
func configureDefaultLockHostService(t *testing.T) {
	t.Helper()
	previousLockSvc := lockHostService
	lockSvc, err := hostlock.New(locker.New())
	if err != nil {
		t.Fatalf("create host lock service failed: %v", err)
	}
	if err = ConfigureLockHostService(lockSvc); err != nil {
		t.Fatalf("configure host lock service failed: %v", err)
	}
	t.Cleanup(func() {
		lockHostService = previousLockSvc
	})
}

// ensurePluginLockerTable creates the lock table needed by lock host call tests.
func ensurePluginLockerTable(t *testing.T, ctx context.Context) {
	t.Helper()
	for _, statement := range dialect.SplitSQLStatements(createPluginLockerTableSQL) {
		if _, err := g.DB().Exec(ctx, statement); err != nil {
			t.Fatalf("expected sys_locker table to be created, got error: %v\nSQL:\n%s", err, statement)
		}
	}
}

// cleanupPluginLock removes lock rows created by the current test.
func cleanupPluginLock(t *testing.T, ctx context.Context, lockName string) {
	t.Helper()
	if _, err := dao.SysLocker.Ctx(ctx).Where(do.SysLocker{Name: lockName}).Delete(); err != nil {
		t.Fatalf("failed to cleanup plugin lock %s: %v", lockName, err)
	}
}

// buildPluginLockName builds the fully qualified lock name used by the backend.
func buildPluginLockName(pluginID string, lockName string) string {
	return "plugin:" + pluginID + ":tenant=0:" + lockName
}

// newLockHostCallContext builds a host call context authorized for the given lock names.
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

// newTenantLockHostCallContext builds a host call context with a tenant identity.
func newTenantLockHostCallContext(pluginID string, tenantID int32, lockNames ...string) *hostCallContext {
	hcc := newLockHostCallContext(pluginID, lockNames...)
	hcc.identity = &pluginbridge.IdentitySnapshotV1{TenantId: tenantID, UserID: 1, Username: "admin"}
	return hcc
}

// acquireLockTicket acquires one lock and returns the issued ticket.
func acquireLockTicket(t *testing.T, hcc *hostCallContext, lockName string) string {
	t.Helper()
	response := invokeLockHostService(
		t,
		hcc,
		pluginbridge.HostServiceMethodLockAcquire,
		lockName,
		pluginbridge.MarshalHostServiceLockAcquireRequest(&pluginbridge.HostServiceLockAcquireRequest{LeaseMillis: 5000}),
	)
	if response.Status != pluginbridge.HostCallStatusSuccess {
		t.Fatalf("acquire coordination lock expected success, got status=%d payload=%s", response.Status, string(response.Payload))
	}
	payload, err := pluginbridge.UnmarshalHostServiceLockAcquireResponse(response.Payload)
	if err != nil {
		t.Fatalf("decode acquire payload: %v", err)
	}
	if !payload.Acquired || strings.TrimSpace(payload.Ticket) == "" {
		t.Fatalf("expected acquired lock ticket, got %#v", payload)
	}
	return payload.Ticket
}

// releaseLockTicket releases one lock ticket through the host service.
func releaseLockTicket(t *testing.T, hcc *hostCallContext, lockName string, ticket string) {
	t.Helper()
	response := invokeLockHostService(
		t,
		hcc,
		pluginbridge.HostServiceMethodLockRelease,
		lockName,
		pluginbridge.MarshalHostServiceLockReleaseRequest(&pluginbridge.HostServiceLockReleaseRequest{Ticket: ticket}),
	)
	if response.Status != pluginbridge.HostCallStatusSuccess {
		t.Fatalf("release coordination lock expected success, got status=%d payload=%s", response.Status, string(response.Payload))
	}
}

// invokeLockHostService marshals and dispatches one lock host service request.
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
