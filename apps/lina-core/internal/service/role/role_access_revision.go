// This file defines deployment-aware permission-topology revision controllers
// selected during role-service construction.

package role

import (
	"context"
	"time"

	"lina-core/internal/service/cachecoord"
	"lina-core/internal/service/datascope"
)

// Permission-access cache coordination reasons.
const (
	// accessTopologyCacheDomain coordinates token-scoped permission access snapshots.
	accessTopologyCacheDomain cachecoord.Domain = "permission-access"
	// accessTopologyCacheChangeReason records permission-topology mutations.
	accessTopologyCacheChangeReason cachecoord.ChangeReason = "access_topology_changed"
	// accessTopologyCacheMaxStale is the permission-access freshness budget.
	accessTopologyCacheMaxStale = 3 * time.Second
)

// accessRevisionController hides the single-node and clustered revision
// synchronization strategies behind one common contract.
type accessRevisionController interface {
	// CurrentRevision returns the effective permission-topology revision currently visible to the process.
	CurrentRevision(ctx context.Context) (int64, error)
	// SyncRevision refreshes the process-local revision from the active source and clears stale token snapshots when needed.
	SyncRevision(ctx context.Context, onRevisionChange func()) (int64, error)
	// MarkChanged records one topology mutation and returns the new effective revision.
	MarkChanged(ctx context.Context) (int64, error)
}

// localAccessRevisionController keeps permission-topology invalidation fully
// in process memory for single-node deployments.
type localAccessRevisionController struct{}

// clusterAccessRevisionController synchronizes topology revision through
// cachecoord while preserving a local hot-path copy for request-time permission checks.
type clusterAccessRevisionController struct {
	cacheCoordSvc cachecoord.Service
}

// newCacheCoordAccessRevisionController selects the deployment-specific
// revision strategy backed by cachecoord in cluster mode.
func newCacheCoordAccessRevisionController(clusterEnabled bool) accessRevisionController {
	if clusterEnabled {
		cacheCoordSvc := cachecoord.Default(cachecoord.NewStaticTopology(true))
		configureAccessTopologyCacheDomain(cacheCoordSvc)
		return &clusterAccessRevisionController{
			cacheCoordSvc: cacheCoordSvc,
		}
	}
	return &localAccessRevisionController{}
}

// configureAccessTopologyCacheDomain declares the permission-access consistency
// contract in the role module that owns the cache semantics.
func configureAccessTopologyCacheDomain(cacheCoordSvc cachecoord.Service) {
	if cacheCoordSvc == nil {
		return
	}
	if err := cacheCoordSvc.ConfigureDomain(cachecoord.DomainSpec{
		Domain:           accessTopologyCacheDomain,
		AuthoritySource:  "sys_role, sys_role_menu, sys_user_role, sys_menu, plugin permissions",
		ConsistencyModel: cachecoord.ConsistencySharedRevision,
		MaxStale:         accessTopologyCacheMaxStale,
		SyncMechanism:    "persistent sys_cache_revision plus request or watcher refresh",
		FailureStrategy:  cachecoord.FailureStrategyFailClosed,
	}); err != nil {
		panic(err)
	}
}

// CurrentRevision reuses the last in-process value or lazily initializes one
// so single-node permission checks never need a distributed coordination step.
func (c *localAccessRevisionController) CurrentRevision(_ context.Context) (int64, error) {
	if revision, ok := getLocalAccessRevisionForce(); ok {
		return revision, nil
	}
	revision := bumpLocalAccessRevision()
	return revision, nil
}

// SyncRevision is the same as CurrentRevision in single-node mode because no
// other node can advance the permission topology independently.
func (c *localAccessRevisionController) SyncRevision(ctx context.Context, _ func()) (int64, error) {
	return c.CurrentRevision(ctx)
}

// MarkChanged advances the local revision immediately after one topology write
// so token snapshots created afterward bind to the new version.
func (c *localAccessRevisionController) MarkChanged(_ context.Context) (int64, error) {
	return bumpLocalAccessRevision(), nil
}

// CurrentRevision first tries the short-lived local copy and only re-reads
// cachecoord when this process needs to resynchronize.
func (c *clusterAccessRevisionController) CurrentRevision(ctx context.Context) (int64, error) {
	if revision, ok := getLocalAccessRevision(); ok {
		return revision, nil
	}

	return c.ensureFresh(ctx)
}

// SyncRevision is used by the background watcher, so it must always consult
// cachecoord and optionally trigger token-cache eviction when another node
// wrote a newer topology revision.
func (c *clusterAccessRevisionController) SyncRevision(
	ctx context.Context,
	onRevisionChange func(),
) (int64, error) {
	localRevision, hadLocalRevision := getLocalAccessRevisionForce()
	revision, err := c.ensureFresh(ctx)
	if err != nil {
		return 0, err
	}

	if hadLocalRevision && localRevision != revision && onRevisionChange != nil {
		onRevisionChange()
	}
	return revision, nil
}

// MarkChanged publishes one shared revision bump and then updates the local
// copy so the writing node observes its own topology mutation immediately.
func (c *clusterAccessRevisionController) MarkChanged(ctx context.Context) (int64, error) {
	revision, err := c.cacheCoordSvc.MarkTenantChanged(
		ctx,
		accessTopologyCacheDomain,
		cachecoord.ScopeGlobal,
		accessRevisionInvalidationScope(ctx),
		accessTopologyCacheChangeReason,
	)
	if err != nil {
		return 0, err
	}
	storeLocalAccessRevision(revision)
	return revision, nil
}

// ensureFresh confirms that the local permission snapshot has consumed the
// latest coordinated revision or fails closed after the stale window.
func (c *clusterAccessRevisionController) ensureFresh(ctx context.Context) (int64, error) {
	revision, err := c.cacheCoordSvc.EnsureFresh(
		ctx,
		accessTopologyCacheDomain,
		accessRevisionScope(ctx),
		func(_ context.Context, revision int64) error {
			storeLocalAccessRevision(revision)
			return nil
		},
	)
	if err != nil {
		return 0, err
	}
	storeLocalAccessRevision(revision)
	return revision, nil
}

// accessRevisionScope returns the tenant-partitioned permission revision scope
// used by clustered permission access snapshots.
func accessRevisionScope(ctx context.Context) cachecoord.Scope {
	return cachecoord.ScopedScope(
		cachecoord.ScopeGlobal,
		accessRevisionInvalidationScope(ctx),
	)
}

// accessRevisionInvalidationScope describes which tenant bucket one permission
// topology mutation affects. Platform changes cascade because platform menus and
// defaults can alter every tenant's effective permission set.
func accessRevisionInvalidationScope(ctx context.Context) cachecoord.InvalidationScope {
	tenantID := datascope.CurrentTenantID(ctx)
	return cachecoord.InvalidationScope{
		TenantID:         cachecoord.TenantID(tenantID),
		CascadeToTenants: tenantID == datascope.PlatformTenantID,
	}
}
