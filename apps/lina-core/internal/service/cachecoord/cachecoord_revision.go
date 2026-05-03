// This file implements local and shared cache revision coordination.

package cachecoord

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
)

// Cache revision identity and storage limits mirror sys_cache_revision columns.
const (
	maxDomainBytes = 64
	maxScopeBytes  = 128
	maxReasonBytes = 255
)

// revisionKey uniquely identifies one cache domain/scope revision row.
type revisionKey struct {
	domain Domain
	scope  Scope
}

// processLocalRevisions shares single-node revision state across service
// instances because several host services construct their own dependencies.
var processLocalRevisions = struct {
	sync.Mutex
	values map[revisionKey]int64
}{
	values: make(map[revisionKey]int64),
}

// processCoordinationStatuses aggregates observable status across cachecoord
// instances in the current process without changing per-instance freshness
// decisions.
var processCoordinationStatuses = struct {
	sync.RWMutex
	values map[revisionKey]coordinationStatus
}{
	values: make(map[revisionKey]coordinationStatus),
}

// ConfigureDomain configures or replaces one cache domain consistency contract.
func (s *serviceImpl) ConfigureDomain(spec DomainSpec) error {
	normalizedDomain := Domain(strings.TrimSpace(string(spec.Domain)))
	if normalizedDomain == "" || len([]byte(normalizedDomain)) > maxDomainBytes {
		return bizerr.NewCode(CodeCacheCoordDomainInvalid)
	}
	if spec.MaxStale <= 0 {
		spec.MaxStale = DefaultDomainMaxStale
	}
	if spec.ConsistencyModel == "" {
		spec.ConsistencyModel = ConsistencySharedRevision
	}
	if spec.FailureStrategy == "" {
		spec.FailureStrategy = FailureStrategyReturnVisibleError
	}
	spec.Domain = normalizedDomain

	s.mu.Lock()
	s.domains[normalizedDomain] = spec
	s.mu.Unlock()
	return nil
}

// MarkChanged publishes one explicit cache domain/scope revision change.
func (s *serviceImpl) MarkChanged(
	ctx context.Context,
	domain Domain,
	scope Scope,
	reason ChangeReason,
) (int64, error) {
	key, err := s.resolveKey(domain, scope)
	if err != nil {
		return 0, err
	}

	if !s.clusterEnabled() {
		revision := bumpLocalRevision(key)
		s.storeObservedRevision(key, revision)
		s.recordSuccess(key, revision, revision, time.Now())
		return revision, nil
	}

	revision, err := s.bumpSharedRevision(ctx, key, reason)
	if err != nil {
		s.recordFailure(key, err)
		logger.Warningf(
			ctx,
			"publish cache coordination revision failed domain=%s scope=%s err=%v",
			key.domain,
			key.scope,
			err,
		)
		return 0, bizerr.WrapCode(
			err,
			CodeCacheCoordPublishFailed,
			bizerr.P("domain", key.domain),
			bizerr.P("scope", key.scope),
		)
	}

	s.storeObservedRevision(key, revision)
	s.recordSuccess(key, revision, revision, time.Now())
	return revision, nil
}

// EnsureFresh refreshes local state if the shared or local revision advanced.
func (s *serviceImpl) EnsureFresh(
	ctx context.Context,
	domain Domain,
	scope Scope,
	refresher Refresher,
) (int64, error) {
	key, err := s.resolveKey(domain, scope)
	if err != nil {
		return 0, err
	}

	revision, err := s.CurrentRevision(ctx, key.domain, key.scope)
	if err != nil {
		if s.isWithinStaleWindow(key) {
			logger.Warningf(
				ctx,
				"cache coordination revision read failed within stale window domain=%s scope=%s err=%v",
				key.domain,
				key.scope,
				err,
			)
			return s.observedRevision(key), nil
		}
		s.recordFailure(key, err)
		return 0, bizerr.WrapCode(
			err,
			CodeCacheCoordFreshnessUnavailable,
			bizerr.P("domain", key.domain),
			bizerr.P("scope", key.scope),
		)
	}

	if s.observedRevision(key) == revision {
		s.recordSuccess(key, revision, revision, time.Now())
		return revision, nil
	}
	if refresher != nil {
		if err = refresher(ctx, revision); err != nil {
			s.recordFailure(key, err)
			return 0, err
		}
	}
	s.storeObservedRevision(key, revision)
	s.recordSuccess(key, revision, revision, time.Now())
	return revision, nil
}

// CurrentRevision returns the latest visible revision for one domain/scope.
func (s *serviceImpl) CurrentRevision(ctx context.Context, domain Domain, scope Scope) (int64, error) {
	key, err := s.resolveKey(domain, scope)
	if err != nil {
		return 0, err
	}
	if !s.clusterEnabled() {
		revision := currentLocalRevision(key)
		s.recordSuccess(key, revision, revision, time.Now())
		return revision, nil
	}

	revision, err := s.currentSharedRevision(ctx, key)
	if err != nil {
		s.recordFailure(key, err)
		return 0, bizerr.WrapCode(
			err,
			CodeCacheCoordRevisionUnavailable,
			bizerr.P("domain", key.domain),
			bizerr.P("scope", key.scope),
		)
	}
	s.recordSharedRevision(key, revision)
	return revision, nil
}

// Snapshot returns observable status for configured cache domains and touched scopes.
func (s *serviceImpl) Snapshot(ctx context.Context) ([]SnapshotItem, error) {
	s.mu.RLock()
	keys := make([]revisionKey, 0, len(s.status)+len(s.domains))
	seen := make(map[revisionKey]struct{}, len(s.status)+len(s.domains))
	for key := range s.status {
		keys = append(keys, key)
		seen[key] = struct{}{}
	}
	for domain := range s.domains {
		key := revisionKey{domain: domain, scope: ScopeGlobal}
		if _, ok := seen[key]; !ok {
			keys = append(keys, key)
		}
	}
	s.mu.RUnlock()

	processCoordinationStatuses.RLock()
	for key := range processCoordinationStatuses.values {
		if _, ok := seen[key]; ok {
			continue
		}
		keys = append(keys, key)
		seen[key] = struct{}{}
	}
	processCoordinationStatuses.RUnlock()

	items := make([]SnapshotItem, 0, len(keys))
	for _, key := range keys {
		sharedRevision := int64(0)
		if s.clusterEnabled() {
			revision, err := s.currentSharedRevision(ctx, key)
			if err != nil {
				s.recordFailure(key, err)
			} else {
				sharedRevision = revision
				s.recordSharedRevision(key, revision)
			}
		} else {
			sharedRevision = currentLocalRevision(key)
		}
		items = append(items, s.snapshotItem(key, sharedRevision))
	}
	return items, nil
}

// bumpSharedRevision increments one persistent revision row under a transaction
// and row lock so concurrent publishers cannot lose increments.
func (s *serviceImpl) bumpSharedRevision(
	ctx context.Context,
	key revisionKey,
	reason ChangeReason,
) (int64, error) {
	normalizedReason := normalizeReason(reason)
	_, err := dao.SysCacheRevision.Ctx(ctx).Data(do.SysCacheRevision{
		Domain:   key.domain,
		Scope:    key.scope,
		Revision: 0,
		Reason:   normalizedReason,
	}).InsertIgnore()
	if err != nil {
		return 0, err
	}

	var revision int64
	err = dao.SysCacheRevision.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		var row *entity.SysCacheRevision
		err = dao.SysCacheRevision.Ctx(ctx).Where(do.SysCacheRevision{
			Domain: key.domain,
			Scope:  key.scope,
		}).LockUpdate().Scan(&row)
		if err != nil {
			return err
		}
		if row == nil {
			return bizerr.NewCode(
				CodeCacheCoordRevisionUnavailable,
				bizerr.P("domain", key.domain),
				bizerr.P("scope", key.scope),
			)
		}

		revision = row.Revision + 1
		_, err = dao.SysCacheRevision.Ctx(ctx).Where(do.SysCacheRevision{Id: row.Id}).Data(do.SysCacheRevision{
			Revision: revision,
			Reason:   normalizedReason,
		}).Update()
		return err
	})
	if err != nil {
		return 0, err
	}
	return revision, nil
}

// currentSharedRevision reads the persistent revision for one domain/scope.
func (s *serviceImpl) currentSharedRevision(ctx context.Context, key revisionKey) (int64, error) {
	var row *entity.SysCacheRevision
	err := dao.SysCacheRevision.Ctx(ctx).Where(do.SysCacheRevision{
		Domain: key.domain,
		Scope:  key.scope,
	}).Scan(&row)
	if err != nil {
		return 0, err
	}
	if row == nil {
		return 0, nil
	}
	return row.Revision, nil
}

// resolveKey validates one domain/scope pair and returns the normalized key.
func (s *serviceImpl) resolveKey(domain Domain, scope Scope) (revisionKey, error) {
	key := revisionKey{
		domain: Domain(strings.TrimSpace(string(domain))),
		scope:  Scope(strings.TrimSpace(string(scope))),
	}
	if key.scope == "" {
		key.scope = ScopeGlobal
	}
	if key.domain == "" || len([]byte(key.domain)) > maxDomainBytes {
		return revisionKey{}, bizerr.NewCode(CodeCacheCoordDomainInvalid)
	}
	if len([]byte(key.scope)) > maxScopeBytes {
		return revisionKey{}, bizerr.NewCode(CodeCacheCoordScopeInvalid)
	}
	return key, nil
}

// clusterEnabled reports whether shared persistent coordination is active.
func (s *serviceImpl) clusterEnabled() bool {
	topology := s.topologySnapshot()
	return topology != nil && topology.IsEnabled()
}

// topologySnapshot returns the current topology view without exposing mutable
// coordinator state to callers.
func (s *serviceImpl) topologySnapshot() Topology {
	if s == nil {
		return nil
	}
	s.topologyMu.RLock()
	topology := s.topology
	s.topologyMu.RUnlock()
	return topology
}

// recordSuccess updates local observable state after one successful sync.
func (s *serviceImpl) recordSuccess(
	key revisionKey,
	localRevision int64,
	sharedRevision int64,
	now time.Time,
) {
	s.mu.Lock()
	status := s.ensureStatusLocked(key)
	status.localRevision = localRevision
	status.sharedRevision = sharedRevision
	status.lastSyncedAt = now
	status.recentError = ""
	status.recentErrorAt = time.Time{}
	s.mu.Unlock()
	recordProcessSuccess(key, localRevision, sharedRevision, now)
}

// recordSharedRevision stores a shared revision read without changing observed freshness.
func (s *serviceImpl) recordSharedRevision(key revisionKey, sharedRevision int64) {
	s.mu.Lock()
	status := s.ensureStatusLocked(key)
	status.sharedRevision = sharedRevision
	s.mu.Unlock()
	recordProcessSharedRevision(key, sharedRevision)
}

// recordFailure stores the latest coordination failure for diagnostics.
func (s *serviceImpl) recordFailure(key revisionKey, err error) {
	if err == nil {
		return
	}
	s.mu.Lock()
	status := s.ensureStatusLocked(key)
	status.recentError = err.Error()
	status.recentErrorAt = time.Now()
	s.mu.Unlock()
	recordProcessFailure(key, err)
}

// ensureStatusLocked returns the mutable status object for a key. Caller must hold s.mu.
func (s *serviceImpl) ensureStatusLocked(key revisionKey) *coordinationStatus {
	status := s.status[key]
	if status == nil {
		status = &coordinationStatus{}
		s.status[key] = status
	}
	return status
}

// isWithinStaleWindow reports whether the latest successful sync is still
// inside the cache domain's failure window.
func (s *serviceImpl) isWithinStaleWindow(key revisionKey) bool {
	s.mu.RLock()
	status := s.status[key]
	s.mu.RUnlock()
	if status == nil || status.lastSyncedAt.IsZero() {
		return false
	}
	spec := s.domainSpec(key.domain)
	return time.Since(status.lastSyncedAt) <= spec.MaxStale
}

// observedRevision returns the service-local consumed revision for a key.
func (s *serviceImpl) observedRevision(key revisionKey) int64 {
	s.mu.RLock()
	revision := s.observed[key]
	s.mu.RUnlock()
	return revision
}

// storeObservedRevision records the revision consumed by this service instance.
func (s *serviceImpl) storeObservedRevision(key revisionKey, revision int64) {
	s.mu.Lock()
	s.observed[key] = revision
	status := s.ensureStatusLocked(key)
	status.localRevision = revision
	status.lastSyncedAt = time.Now()
	sharedRevision := status.sharedRevision
	lastSyncedAt := status.lastSyncedAt
	s.mu.Unlock()
	recordProcessSuccess(key, revision, sharedRevision, lastSyncedAt)
}

// snapshotItem builds one detached observable snapshot row.
func (s *serviceImpl) snapshotItem(key revisionKey, sharedRevision int64) SnapshotItem {
	s.mu.RLock()
	status := s.status[key]
	var localStatus coordinationStatus
	if status != nil {
		localStatus = *status
	}
	s.mu.RUnlock()
	spec := s.domainSpec(key.domain)
	processStatus := processStatusSnapshot(key)
	if localStatus.lastSyncedAt.IsZero() && !processStatus.lastSyncedAt.IsZero() {
		localStatus = processStatus
	} else {
		if processStatus.localRevision > localStatus.localRevision {
			localStatus.localRevision = processStatus.localRevision
		}
		if processStatus.sharedRevision > localStatus.sharedRevision {
			localStatus.sharedRevision = processStatus.sharedRevision
		}
		if processStatus.lastSyncedAt.After(localStatus.lastSyncedAt) {
			localStatus.lastSyncedAt = processStatus.lastSyncedAt
		}
		if processStatus.recentErrorAt.After(localStatus.recentErrorAt) {
			localStatus.recentError = processStatus.recentError
			localStatus.recentErrorAt = processStatus.recentErrorAt
		}
	}

	staleSeconds := int64(0)
	if !localStatus.lastSyncedAt.IsZero() {
		staleSeconds = int64(time.Since(localStatus.lastSyncedAt).Seconds())
	}
	if !s.clusterEnabled() && localStatus.localRevision == 0 {
		localStatus.localRevision = sharedRevision
	}
	return SnapshotItem{
		Domain:           key.domain,
		Scope:            key.scope,
		AuthoritySource:  spec.AuthoritySource,
		ConsistencyModel: spec.ConsistencyModel,
		MaxStale:         spec.MaxStale,
		FailureStrategy:  spec.FailureStrategy,
		LocalRevision:    localStatus.localRevision,
		SharedRevision:   maxInt64(sharedRevision, localStatus.sharedRevision),
		LastSyncedAt:     localStatus.lastSyncedAt,
		RecentError:      localStatus.recentError,
		StaleSeconds:     staleSeconds,
	}
}

// domainSpec returns the configured domain contract or the default contract for
// free-form domains that have not configured custom behavior.
func (s *serviceImpl) domainSpec(domain Domain) DomainSpec {
	s.mu.RLock()
	spec, ok := s.domains[domain]
	s.mu.RUnlock()
	if ok {
		return spec
	}
	return defaultDomainSpec(domain)
}

// defaultDomainSpec returns the built-in fallback consistency contract used by
// any valid cache domain that has not configured domain-specific metadata.
func defaultDomainSpec(domain Domain) DomainSpec {
	return DomainSpec{
		Domain:           domain,
		AuthoritySource:  "caller-owned cache domain",
		ConsistencyModel: ConsistencySharedRevision,
		MaxStale:         DefaultDomainMaxStale,
		SyncMechanism:    "persistent MySQL sys_cache_revision plus request or watcher refresh",
		FailureStrategy:  FailureStrategyReturnVisibleError,
	}
}

// recordProcessSuccess updates process-level diagnostics after a successful
// local freshness observation.
func recordProcessSuccess(
	key revisionKey,
	localRevision int64,
	sharedRevision int64,
	now time.Time,
) {
	processCoordinationStatuses.Lock()
	status := processCoordinationStatuses.values[key]
	status.localRevision = maxInt64(status.localRevision, localRevision)
	status.sharedRevision = maxInt64(status.sharedRevision, sharedRevision)
	status.lastSyncedAt = now
	status.recentError = ""
	status.recentErrorAt = time.Time{}
	processCoordinationStatuses.values[key] = status
	processCoordinationStatuses.Unlock()
}

// recordProcessSharedRevision updates process-level diagnostics for shared
// revision reads that do not imply local refresh completion.
func recordProcessSharedRevision(key revisionKey, sharedRevision int64) {
	processCoordinationStatuses.Lock()
	status := processCoordinationStatuses.values[key]
	status.sharedRevision = maxInt64(status.sharedRevision, sharedRevision)
	processCoordinationStatuses.values[key] = status
	processCoordinationStatuses.Unlock()
}

// recordProcessFailure stores the latest process-level coordination failure.
func recordProcessFailure(key revisionKey, err error) {
	if err == nil {
		return
	}
	processCoordinationStatuses.Lock()
	status := processCoordinationStatuses.values[key]
	status.recentError = err.Error()
	status.recentErrorAt = time.Now()
	processCoordinationStatuses.values[key] = status
	processCoordinationStatuses.Unlock()
}

// processStatusSnapshot returns the current process-level diagnostic row for a
// domain/scope key.
func processStatusSnapshot(key revisionKey) coordinationStatus {
	processCoordinationStatuses.RLock()
	status := processCoordinationStatuses.values[key]
	processCoordinationStatuses.RUnlock()
	return status
}

// bumpLocalRevision increments one process-local domain/scope revision.
func bumpLocalRevision(key revisionKey) int64 {
	processLocalRevisions.Lock()
	defer processLocalRevisions.Unlock()
	processLocalRevisions.values[key]++
	if processLocalRevisions.values[key] <= 0 {
		processLocalRevisions.values[key] = 1
	}
	return processLocalRevisions.values[key]
}

// currentLocalRevision returns one initialized process-local revision.
func currentLocalRevision(key revisionKey) int64 {
	processLocalRevisions.Lock()
	defer processLocalRevisions.Unlock()
	revision := processLocalRevisions.values[key]
	if revision <= 0 {
		revision = 1
		processLocalRevisions.values[key] = revision
	}
	return revision
}

// normalizeReason trims and bounds the diagnostic reason string.
func normalizeReason(reason ChangeReason) string {
	value := strings.TrimSpace(string(reason))
	if len([]byte(value)) <= maxReasonBytes {
		return value
	}

	bytes := []byte(value)
	return string(bytes[:maxReasonBytes])
}

// maxInt64 returns the larger of two revision counters.
func maxInt64(a int64, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
