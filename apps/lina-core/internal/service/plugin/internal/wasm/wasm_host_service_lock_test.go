// This file tests lock host service dispatch, ticket validation, and authorization.

package wasm

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/service/coordination"
	"lina-core/internal/service/hostlock"
	"lina-core/internal/service/locker"
	"lina-core/pkg/dialect"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/lockcap"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
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
		ExpireAt: timePtr(time.Now().Add(5 * time.Second)),
	}, nil
}

// Renew records one lock renewal request.
func (s *trackingLockService) Renew(context.Context, string, int64, string, string) (*time.Time, error) {
	s.renewCalls++
	return timePtr(time.Now().Add(5 * time.Second)), nil
}

// timePtr returns a pointer to value for hostlock test responses.
func timePtr(value time.Time) *time.Time {
	return &value
}

// Release records one lock release request.
func (s *trackingLockService) Release(context.Context, string, int64, string, string) error {
	s.releaseCalls++
	return nil
}

// lockDomainTestService adapts the existing hostlock service into the
// plugin-visible lockcap.Service used by the unified domain dispatcher.
type lockDomainTestService struct {
	service  hostlock.Service
	pluginID string
}

// Acquire records or acquires one plugin-scoped lock through hostlock.
func (s *lockDomainTestService) Acquire(ctx context.Context, in lockcap.AcquireInput) (*lockcap.AcquireOutput, error) {
	output, err := s.service.Acquire(ctx, hostlock.AcquireInput{
		PluginID:    s.pluginID,
		TenantID:    lockDomainTenantID(ctx),
		ResourceRef: in.Name,
		LeaseMillis: in.Lease.Milliseconds(),
	})
	if err != nil {
		return nil, err
	}
	if output == nil {
		return &lockcap.AcquireOutput{Acquired: false}, nil
	}
	return &lockcap.AcquireOutput{
		Acquired: output.Acquired,
		Ticket:   output.Ticket,
		ExpireAt: output.ExpireAt,
	}, nil
}

// Renew extends one plugin-scoped lock through hostlock.
func (s *lockDomainTestService) Renew(ctx context.Context, in lockcap.RenewInput) (*lockcap.RenewOutput, error) {
	expireAt, err := s.service.Renew(ctx, s.pluginID, lockDomainTenantID(ctx), in.Name, in.Ticket)
	if err != nil {
		return nil, err
	}
	return &lockcap.RenewOutput{ExpireAt: expireAt}, nil
}

// Release releases one plugin-scoped lock through hostlock.
func (s *lockDomainTestService) Release(ctx context.Context, in lockcap.ReleaseInput) error {
	return s.service.Release(ctx, s.pluginID, lockDomainTenantID(ctx), in.Name, in.Ticket)
}

// lockDomainTenantID resolves the tenant scope injected by the host-call dispatcher.
func lockDomainTenantID(ctx context.Context) int64 {
	current := bizctxcap.CurrentFromContext(ctx)
	if current.TenantID > 0 {
		return int64(current.TenantID)
	}
	return int64(tenantcap.PLATFORM)
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
	configureDefaultLockDomainService(t)

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
		protocol.HostServiceMethodLockAcquire,
		lockName,
		protocol.MarshalHostServiceLockAcquireRequest(&protocol.HostServiceLockAcquireRequest{LeaseMillis: 5000}),
	)
	if acquireResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("acquire: expected success, got status=%d payload=%s", acquireResponse.Status, string(acquireResponse.Payload))
	}
	acquirePayload, err := protocol.UnmarshalHostServiceLockAcquireResponse(acquireResponse.Payload)
	if err != nil {
		t.Fatalf("acquire payload decode failed: %v", err)
	}
	if !acquirePayload.Acquired || strings.TrimSpace(acquirePayload.Ticket) == "" {
		t.Fatalf("acquire payload: got %#v", acquirePayload)
	}
	if _, err := time.Parse(time.RFC3339Nano, acquirePayload.ExpireAt); err != nil {
		t.Fatalf("acquire expireAt should use RFC3339Nano wire format, got %q: %v", acquirePayload.ExpireAt, err)
	}

	duplicateAcquireResponse := invokeLockHostService(
		t,
		hcc,
		protocol.HostServiceMethodLockAcquire,
		lockName,
		protocol.MarshalHostServiceLockAcquireRequest(&protocol.HostServiceLockAcquireRequest{LeaseMillis: 5000}),
	)
	if duplicateAcquireResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("duplicate acquire: expected success envelope, got status=%d payload=%s", duplicateAcquireResponse.Status, string(duplicateAcquireResponse.Payload))
	}
	duplicateAcquirePayload, err := protocol.UnmarshalHostServiceLockAcquireResponse(duplicateAcquireResponse.Payload)
	if err != nil {
		t.Fatalf("duplicate acquire payload decode failed: %v", err)
	}
	if duplicateAcquirePayload.Acquired {
		t.Fatalf("expected duplicate acquire to be rejected by lock holder, got %#v", duplicateAcquirePayload)
	}

	renewResponse := invokeLockHostService(
		t,
		hcc,
		protocol.HostServiceMethodLockRenew,
		lockName,
		protocol.MarshalHostServiceLockRenewRequest(&protocol.HostServiceLockRenewRequest{Ticket: acquirePayload.Ticket}),
	)
	if renewResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("renew: expected success, got status=%d payload=%s", renewResponse.Status, string(renewResponse.Payload))
	}
	renewPayload, err := protocol.UnmarshalHostServiceLockRenewResponse(renewResponse.Payload)
	if err != nil {
		t.Fatalf("renew payload decode failed: %v", err)
	}
	if strings.TrimSpace(renewPayload.ExpireAt) == "" {
		t.Fatalf("renew payload: got %#v", renewPayload)
	}
	if _, err := time.Parse(time.RFC3339Nano, renewPayload.ExpireAt); err != nil {
		t.Fatalf("renew expireAt should use RFC3339Nano wire format, got %q: %v", renewPayload.ExpireAt, err)
	}

	releaseResponse := invokeLockHostService(
		t,
		hcc,
		protocol.HostServiceMethodLockRelease,
		lockName,
		protocol.MarshalHostServiceLockReleaseRequest(&protocol.HostServiceLockReleaseRequest{Ticket: acquirePayload.Ticket}),
	)
	if releaseResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("release: expected success, got status=%d payload=%s", releaseResponse.Status, string(releaseResponse.Payload))
	}

	reacquireResponse := invokeLockHostService(
		t,
		hcc,
		protocol.HostServiceMethodLockAcquire,
		lockName,
		protocol.MarshalHostServiceLockAcquireRequest(&protocol.HostServiceLockAcquireRequest{LeaseMillis: 5000}),
	)
	if reacquireResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("reacquire: expected success, got status=%d payload=%s", reacquireResponse.Status, string(reacquireResponse.Payload))
	}
	reacquirePayload, err := protocol.UnmarshalHostServiceLockAcquireResponse(reacquireResponse.Payload)
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
	configureDefaultLockDomainService(t)

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
		protocol.HostServiceMethodLockAcquire,
		lockName,
		protocol.MarshalHostServiceLockAcquireRequest(&protocol.HostServiceLockAcquireRequest{LeaseMillis: 5000}),
	)
	if acquireResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("acquire: expected success, got status=%d payload=%s", acquireResponse.Status, string(acquireResponse.Payload))
	}
	acquirePayload, err := protocol.UnmarshalHostServiceLockAcquireResponse(acquireResponse.Payload)
	if err != nil {
		t.Fatalf("acquire payload decode failed: %v", err)
	}

	mismatchResponse := invokeLockHostService(
		t,
		hcc,
		protocol.HostServiceMethodLockRenew,
		otherLockName,
		protocol.MarshalHostServiceLockRenewRequest(&protocol.HostServiceLockRenewRequest{Ticket: acquirePayload.Ticket}),
	)
	if mismatchResponse.Status != protocol.HostCallStatusInvalidRequest {
		t.Fatalf("expected invalid request for mismatched ticket, got status=%d payload=%s", mismatchResponse.Status, string(mismatchResponse.Payload))
	}
}

// TestHandleHostServiceInvokeLockRejectsUnauthorizedResource verifies unauthorized lock names are rejected.
func TestHandleHostServiceInvokeLockRejectsUnauthorizedResource(t *testing.T) {
	hcc := newLockHostCallContext("test-plugin-lock-denied", "orders-sync")
	response := invokeLockHostService(
		t,
		hcc,
		protocol.HostServiceMethodLockAcquire,
		"inventory-sync",
		protocol.MarshalHostServiceLockAcquireRequest(&protocol.HostServiceLockAcquireRequest{LeaseMillis: 5000}),
	)
	if response.Status != protocol.HostCallStatusCapabilityDenied {
		t.Fatalf("expected capability denied for unauthorized lock name, got status=%d payload=%s", response.Status, string(response.Payload))
	}
}

// TestHandleHostServiceInvokeLockRequiresConfiguredService verifies missing
// domain lock wiring fails explicitly instead of using a package default.
func TestHandleHostServiceInvokeLockRequiresConfiguredService(t *testing.T) {
	configureDomainHostServicesForCapabilityTest(t, &capabilityHostServiceTestServices{})

	hcc := newLockHostCallContext("test-plugin-lock-unconfigured", "orders-sync")
	response := invokeLockHostService(
		t,
		hcc,
		protocol.HostServiceMethodLockAcquire,
		"orders-sync",
		protocol.MarshalHostServiceLockAcquireRequest(&protocol.HostServiceLockAcquireRequest{LeaseMillis: 5000}),
	)
	if response.Status != protocol.HostCallStatusInternalError {
		t.Fatalf("expected internal error for unconfigured lock service, got status=%d payload=%s", response.Status, string(response.Payload))
	}
}

// TestHandleHostServiceInvokeLockUsesConfiguredSharedService verifies lock
// host service dispatch reuses the explicitly configured shared instance.
func TestHandleHostServiceInvokeLockUsesConfiguredSharedService(t *testing.T) {
	lockSvc := &trackingLockService{}
	configureLockDomainServiceForTest(t, &lockDomainTestService{service: lockSvc})

	hcc := newTenantLockHostCallContext("test-plugin-lock-shared", 88, "orders-sync")
	response := invokeLockHostService(
		t,
		hcc,
		protocol.HostServiceMethodLockAcquire,
		"orders-sync",
		protocol.MarshalHostServiceLockAcquireRequest(&protocol.HostServiceLockAcquireRequest{LeaseMillis: 5000}),
	)
	if response.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("acquire through shared lock: expected success, got status=%d payload=%s", response.Status, string(response.Payload))
	}
	payload, err := protocol.UnmarshalHostServiceLockAcquireResponse(response.Payload)
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
	lockSvc, err := hostlock.New(locker.New())
	if err != nil {
		t.Fatalf("create host lock service failed: %v", err)
	}
	configureLockDomainServiceForTest(t, &lockDomainTestService{service: lockSvc})
	t.Cleanup(func() {
		locker.ConfigureCoordination(nil)
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
		protocol.HostServiceMethodLockAcquire,
		lockName,
		protocol.MarshalHostServiceLockAcquireRequest(&protocol.HostServiceLockAcquireRequest{LeaseMillis: 5000}),
	)
	if duplicateResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("duplicate acquire expected success envelope, got status=%d payload=%s", duplicateResponse.Status, string(duplicateResponse.Payload))
	}
	duplicatePayload, err := protocol.UnmarshalHostServiceLockAcquireResponse(duplicateResponse.Payload)
	if err != nil {
		t.Fatalf("decode duplicate acquire payload: %v", err)
	}
	if duplicatePayload.Acquired {
		t.Fatalf("expected duplicate same plugin/tenant lock acquire to miss, got %#v", duplicatePayload)
	}

	wrongRelease := invokeLockHostService(
		t,
		tenantOne,
		protocol.HostServiceMethodLockRelease,
		lockName,
		protocol.MarshalHostServiceLockReleaseRequest(&protocol.HostServiceLockReleaseRequest{Ticket: tenantTwoTicket}),
	)
	if wrongRelease.Status != protocol.HostCallStatusInvalidRequest {
		t.Fatalf("expected cross-tenant release to fail, got status=%d payload=%s", wrongRelease.Status, string(wrongRelease.Payload))
	}

	releaseLockTicket(t, tenantOne, lockName, tenantOneTicket)
	reacquiredTicket := acquireLockTicket(t, tenantOne, lockName)
	if reacquiredTicket == "" {
		t.Fatal("expected reacquired ticket")
	}
}

// configureDefaultLockDomainService wires tests to the same locker-backed host
// lock service shape used by startup.
func configureDefaultLockDomainService(t *testing.T) {
	t.Helper()
	lockSvc, err := hostlock.New(locker.New())
	if err != nil {
		t.Fatalf("create host lock service failed: %v", err)
	}
	configureLockDomainServiceForTest(t, &lockDomainTestService{service: lockSvc})
}

// configureLockDomainServiceForTest installs one lockcap service directory.
func configureLockDomainServiceForTest(t *testing.T, service lockcap.Service) {
	t.Helper()
	configureDomainHostServicesForCapabilityTest(t, &capabilityHostServiceTestServices{lock: service})
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
	resources := make([]*protocol.HostServiceResourceSpec, 0, len(lockNames))
	for _, lockName := range lockNames {
		resources = append(resources, &protocol.HostServiceResourceSpec{Ref: lockName})
	}
	return &hostCallContext{
		pluginID: pluginID,
		capabilities: map[string]struct{}{
			protocol.CapabilityLock: {},
		},
		hostServices: []*protocol.HostServiceSpec{{
			Service: protocol.HostServiceLock,
			Methods: []string{
				protocol.HostServiceMethodLockAcquire,
				protocol.HostServiceMethodLockRelease,
				protocol.HostServiceMethodLockRenew,
			},
			Resources: resources,
		}},
	}
}

// newTenantLockHostCallContext builds a host call context with a tenant identity.
func newTenantLockHostCallContext(pluginID string, tenantID int32, lockNames ...string) *hostCallContext {
	hcc := newLockHostCallContext(pluginID, lockNames...)
	hcc.identity = &protocol.IdentitySnapshotV1{TenantId: tenantID, UserID: 1, Username: "admin"}
	return hcc
}

// acquireLockTicket acquires one lock and returns the issued ticket.
func acquireLockTicket(t *testing.T, hcc *hostCallContext, lockName string) string {
	t.Helper()
	response := invokeLockHostService(
		t,
		hcc,
		protocol.HostServiceMethodLockAcquire,
		lockName,
		protocol.MarshalHostServiceLockAcquireRequest(&protocol.HostServiceLockAcquireRequest{LeaseMillis: 5000}),
	)
	if response.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("acquire coordination lock expected success, got status=%d payload=%s", response.Status, string(response.Payload))
	}
	payload, err := protocol.UnmarshalHostServiceLockAcquireResponse(response.Payload)
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
		protocol.HostServiceMethodLockRelease,
		lockName,
		protocol.MarshalHostServiceLockReleaseRequest(&protocol.HostServiceLockReleaseRequest{Ticket: ticket}),
	)
	if response.Status != protocol.HostCallStatusSuccess {
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
) *protocol.HostCallResponseEnvelope {
	t.Helper()

	request := &protocol.HostServiceRequestEnvelope{
		Service:     protocol.HostServiceLock,
		Method:      method,
		ResourceRef: lockName,
		Payload:     payload,
	}
	return handleHostServiceInvoke(
		context.Background(),
		withTestHostCallRuntime(t, hcc),
		protocol.MarshalHostServiceRequestEnvelope(request),
	)
}
