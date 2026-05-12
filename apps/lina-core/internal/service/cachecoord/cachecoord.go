// Package cachecoord provides topology-aware revision coordination for
// process-local cache domains.
package cachecoord

import (
	"context"
	"sync"
	"time"

	"lina-core/internal/service/coordination"
)

// Domain identifies one cache domain coordinated by the host.
type Domain string

// Scope identifies one explicit invalidation scope inside a cache domain.
type Scope string

// ConsistencyModel names the coordination model used by one cache domain.
type ConsistencyModel string

// FailureStrategy names how callers should degrade when freshness cannot be
// confirmed inside the declared stale window.
type FailureStrategy string

// ChangeReason describes why a cache domain revision was published.
type ChangeReason string

// TenantID identifies the tenant scope of one cache invalidation message.
type TenantID int

// Cache scope constants centralize stable invalidation scopes.
const (
	// ScopeGlobal invalidates the whole cache domain.
	ScopeGlobal Scope = "global"
	// ScopeReconciler invalidates dynamic-plugin runtime reconciler wake-up state.
	ScopeReconciler Scope = "reconciler"
)

// Cache consistency model constants describe runtime review metadata.
const (
	// ConsistencyLocalOnly keeps coordination in the current process.
	ConsistencyLocalOnly ConsistencyModel = "local-only"
	// ConsistencySharedRevision uses a persistent shared revision row.
	ConsistencySharedRevision ConsistencyModel = "shared-revision"
)

// Cache failure strategy constants describe caller-visible degradation behavior.
const (
	// FailureStrategyFailClosed rejects uncertain access after the stale window.
	FailureStrategyFailClosed FailureStrategy = "fail-closed"
	// FailureStrategyReturnVisibleError returns a visible runtime error to callers.
	FailureStrategyReturnVisibleError FailureStrategy = "return-visible-error"
	// FailureStrategyConservativeHide hides uncertain plugin capabilities.
	FailureStrategyConservativeHide FailureStrategy = "conservative-hide"
)

// DefaultDomainMaxStale is the freshness budget used by domains that do not
// configure a domain-specific stale window.
const (
	DefaultDomainMaxStale = 5 * time.Second
)

// Refresher rebuilds or invalidates one process-local cache domain after a
// newer revision is observed.
type Refresher func(ctx context.Context, revision int64) error

// Topology exposes the cluster switch and node metadata needed by cachecoord.
type Topology interface {
	// IsEnabled reports whether clustered deployment mode is enabled.
	IsEnabled() bool
	// IsPrimary reports whether this node owns primary-only background work.
	IsPrimary() bool
	// NodeID returns the stable identifier of the current host node.
	NodeID() string
}

// DomainSpec optionally declares the reviewable consistency contract for one
// cache domain.
type DomainSpec struct {
	Domain           Domain           // Domain is the stable cache domain identifier.
	AuthoritySource  string           // AuthoritySource describes the canonical data source.
	ConsistencyModel ConsistencyModel // ConsistencyModel describes local or shared revision coordination.
	MaxStale         time.Duration    // MaxStale is the maximum acceptable local stale window.
	SyncMechanism    string           // SyncMechanism describes cross-node synchronization.
	FailureStrategy  FailureStrategy  // FailureStrategy describes degradation after MaxStale.
}

// SnapshotItem exposes one cache domain and scope coordination status.
type SnapshotItem struct {
	Domain           Domain           // Domain is the cache domain identifier.
	Scope            Scope            // Scope is the explicit invalidation scope.
	AuthoritySource  string           // AuthoritySource is the canonical data source.
	ConsistencyModel ConsistencyModel // ConsistencyModel is the declared consistency model.
	MaxStale         time.Duration    // MaxStale is the configured stale window.
	FailureStrategy  FailureStrategy  // FailureStrategy is the configured degradation behavior.
	LocalRevision    int64            // LocalRevision is the latest revision consumed locally.
	SharedRevision   int64            // SharedRevision is the latest shared revision when cluster mode is enabled.
	LastSyncedAt     time.Time        // LastSyncedAt records the latest successful local sync.
	Backend                coordination.BackendName // Backend is the active coordination backend for this snapshot.
	CoordinationHealthy   bool      // CoordinationHealthy reports the backend health snapshot when clustered coordination is active.
	EventSubscriberRunning bool      // EventSubscriberRunning reports whether the backend event consumer is active.
	LastEventReceivedAt    time.Time // LastEventReceivedAt records the latest consumed backend event time.
	RecentError      string           // RecentError records the latest coordination failure.
	StaleSeconds     int64            // StaleSeconds reports seconds elapsed since LastSyncedAt.
}

// InvalidationScope declares the tenant range for one cache invalidation.
type InvalidationScope struct {
	TenantID         TenantID // TenantID is the target tenant, 0 platform, or -1 all tenants.
	CascadeToTenants bool     // CascadeToTenants invalidates tenant buckets after platform default changes.
}

// Service defines the cache coordination contract.
type Service interface {
	// ConfigureDomain configures or replaces one cache domain consistency contract.
	ConfigureDomain(spec DomainSpec) error
	// MarkChanged publishes one explicit cache domain/scope revision change.
	MarkChanged(ctx context.Context, domain Domain, scope Scope, reason ChangeReason) (int64, error)
	// MarkTenantChanged publishes one tenant-scoped cache domain/scope revision change.
	MarkTenantChanged(ctx context.Context, domain Domain, scope Scope, tenantScope InvalidationScope, reason ChangeReason) (int64, error)
	// EnsureFresh refreshes local state if the shared or local revision advanced.
	EnsureFresh(ctx context.Context, domain Domain, scope Scope, refresher Refresher) (int64, error)
	// CurrentRevision returns the latest visible revision for one domain/scope.
	CurrentRevision(ctx context.Context, domain Domain, scope Scope) (int64, error)
	// Snapshot returns observable status for configured cache domains and touched scopes.
	Snapshot(ctx context.Context) ([]SnapshotItem, error)
}

// Interface compliance assertion for the default cachecoord service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	topologyMu sync.RWMutex
	topology   Topology
	coordMu    sync.RWMutex
	coord      coordination.Service
	mu         sync.RWMutex
	domains    map[Domain]DomainSpec
	observed   map[revisionKey]int64
	status     map[revisionKey]*coordinationStatus
}

// coordinationStatus stores local observable state for one domain/scope.
type coordinationStatus struct {
	localRevision  int64
	sharedRevision int64
	lastSyncedAt   time.Time
	recentError    string
	recentErrorAt  time.Time
}

// processDefaultService stores the host-wide coordinator used by production
// services that need to share one cache freshness view inside the same process.
var processDefaultService = struct {
	sync.Mutex
	service *serviceImpl
}{}

// New creates an isolated cache coordination service. Production service
// constructors should normally use Default so process-local cache freshness
// state stays shared.
func New(topology Topology) Service {
	return newServiceImpl(topology)
}

// NewWithCoordination creates an isolated cache coordination service that uses
// the provided coordination backend for clustered shared revisions.
func NewWithCoordination(topology Topology, coordinationSvc coordination.Service) Service {
	service := newServiceImpl(topology)
	service.setCoordination(coordinationSvc)
	return service
}

// Default returns the process-wide cache coordination service. When a later
// startup phase provides a richer topology, the existing coordinator is kept
// and only its topology view is updated.
func Default(topology Topology) Service {
	processDefaultService.Lock()
	defer processDefaultService.Unlock()

	if processDefaultService.service == nil {
		processDefaultService.service = newServiceImpl(topology)
		return processDefaultService.service
	}
	if shouldReplaceDefaultTopology(processDefaultService.service.topologySnapshot(), topology) {
		processDefaultService.service.setTopology(topology)
	}
	return processDefaultService.service
}

// DefaultWithCoordination returns the process-wide cache coordination service
// and wires the active distributed coordination backend when one is available.
func DefaultWithCoordination(topology Topology, coordinationSvc coordination.Service) Service {
	processDefaultService.Lock()
	defer processDefaultService.Unlock()

	if processDefaultService.service == nil {
		processDefaultService.service = newServiceImpl(topology)
		processDefaultService.service.setCoordination(coordinationSvc)
		return processDefaultService.service
	}
	if shouldReplaceDefaultTopology(processDefaultService.service.topologySnapshot(), topology) {
		processDefaultService.service.setTopology(topology)
	}
	if coordinationSvc != nil {
		processDefaultService.service.setCoordination(coordinationSvc)
	}
	return processDefaultService.service
}

// newServiceImpl allocates one cache coordination implementation.
func newServiceImpl(topology Topology) *serviceImpl {
	if topology == nil {
		topology = NewStaticTopology(false)
	}
	service := &serviceImpl{
		topology: topology,
		domains:  make(map[Domain]DomainSpec),
		observed: make(map[revisionKey]int64),
		status:   make(map[revisionKey]*coordinationStatus),
	}
	return service
}

// setTopology replaces the coordinator topology without resetting cache-domain
// observations or diagnostic state.
func (s *serviceImpl) setTopology(topology Topology) {
	if s == nil {
		return
	}
	if topology == nil {
		topology = NewStaticTopology(false)
	}
	s.topologyMu.Lock()
	s.topology = topology
	s.topologyMu.Unlock()
}

// setCoordination replaces the distributed coordination backend used in
// clustered mode without resetting local cache observations.
func (s *serviceImpl) setCoordination(coordinationSvc coordination.Service) {
	if s == nil {
		return
	}
	s.coordMu.Lock()
	s.coord = coordinationSvc
	s.coordMu.Unlock()
}

// shouldReplaceDefaultTopology keeps the real cluster topology once it has
// been wired while still allowing early static placeholders to be upgraded.
func shouldReplaceDefaultTopology(current Topology, next Topology) bool {
	if next == nil {
		return false
	}
	if current == nil {
		return true
	}
	if current.IsEnabled() && !next.IsEnabled() {
		return false
	}
	if !current.IsEnabled() && next.IsEnabled() {
		return true
	}
	_, currentStatic := current.(staticTopology)
	_, nextStatic := next.(staticTopology)
	if !currentStatic && nextStatic {
		return false
	}
	return true
}
