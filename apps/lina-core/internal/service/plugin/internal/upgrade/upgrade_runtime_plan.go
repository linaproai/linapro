// This file owns side-effect-free runtime upgrade planning helpers, including
// preview SQL summaries, host-service diffs, risk hints, and explicit upgrade
// lock naming.

package upgrade

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gogf/gf/v2/util/guid"

	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/store"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

const (
	// DistributedLockLease bounds orphaned upgrade locks after a node crash while
	// remaining longer than normal plugin SQL/governance phases.
	DistributedLockLease = 30 * time.Minute
	// DistributedLockReason records the owner purpose in coordination backends.
	DistributedLockReason = "plugin-runtime-upgrade"
)

// SQLSummary summarizes manifest SQL assets visible to preview.
type SQLSummary struct {
	// InstallSQLCount is the number of target install/upgrade SQL assets.
	InstallSQLCount int
	// UninstallSQLCount is the number of target uninstall SQL assets.
	UninstallSQLCount int
	// MockSQLCount is the number of target mock SQL assets excluded from upgrade.
	MockSQLCount int
	// RuntimeSQLAssetCount is the dynamic artifact SQL section count when present.
	RuntimeSQLAssetCount int
}

// HostServicesDiff summarizes service-level hostServices drift.
type HostServicesDiff struct {
	// Added contains services declared by the target manifest but not by the effective snapshot.
	Added []*HostServiceChange
	// Removed contains services no longer requested by the target manifest.
	Removed []*HostServiceChange
	// Changed contains services whose methods or governed targets changed.
	Changed []*HostServiceChange
	// AuthorizationRequired reports whether the target manifest needs host confirmation.
	AuthorizationRequired bool
	// AuthorizationChanged reports whether requested host service scope changed.
	AuthorizationChanged bool
}

// HostServiceChange summarizes one service-level hostServices change.
type HostServiceChange struct {
	// Service is the logical host service identifier.
	Service string
	// FromMethods is the effective method set before upgrade.
	FromMethods []string
	// ToMethods is the target method set after upgrade.
	ToMethods []string
	// FromResourceCount is the number of governed targets before upgrade.
	FromResourceCount int
	// ToResourceCount is the number of governed targets after upgrade.
	ToResourceCount int
	// FromTables is the effective data-table set before upgrade.
	FromTables []string
	// ToTables is the target data-table set after upgrade.
	ToTables []string
	// FromPaths is the effective storage path set before upgrade.
	FromPaths []string
	// ToPaths is the target storage path set after upgrade.
	ToPaths []string
	// FromKeys is the effective authorized host config key set before upgrade.
	FromKeys []string
	// ToKeys is the target authorized host config key set after upgrade.
	ToKeys []string
}

// RiskHintKeys carries stable operator-facing risk hint keys owned by the root facade.
type RiskHintKeys struct {
	// UpgradeSQLRequiresReview warns that upgrade SQL should be reviewed before confirmation.
	UpgradeSQLRequiresReview string
	// MockSQLExcluded warns that mock SQL is never loaded by upgrade.
	MockSQLExcluded string
	// HostServiceAuthorizationChanged warns that hostServices changed.
	HostServiceAuthorizationChanged string
	// DependencyBlockers warns that dependency checks found hard blockers.
	DependencyBlockers string
}

// SnapshotInvalidError reports invalid hostServices in a persisted or target
// manifest snapshot while preserving the snapshot identity for root bizerr wrapping.
type SnapshotInvalidError struct {
	// PluginID is the snapshot plugin identifier.
	PluginID string
	// Version is the snapshot plugin version.
	Version string
	// Cause is the normalization error returned by the protocol package.
	Cause error
}

// Error returns a stable diagnostic for internal wrapping.
func (e *SnapshotInvalidError) Error() string {
	if e == nil || e.Cause == nil {
		return "runtime upgrade snapshot invalid"
	}
	return fmt.Sprintf("runtime upgrade snapshot invalid: %v", e.Cause)
}

// Unwrap returns the underlying protocol normalization error.
func (e *SnapshotInvalidError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// CanExecute reports whether the explicit management upgrade endpoint may run
// side effects for a runtime-upgrade state.
func CanExecute(state plugintypes.RuntimeUpgradeState) bool {
	return state == plugintypes.RuntimeUpgradeStatePendingUpgrade ||
		state == plugintypes.RuntimeUpgradeStateUpgradeFailed
}

// DistributedLockName builds the cluster-wide lock name for one plugin upgrade.
func DistributedLockName(pluginID string) string {
	return "plugin-runtime-upgrade:" + strings.TrimSpace(pluginID)
}

// DistributedLockOwner builds a unique owner for one acquisition so concurrent
// requests from the same node cannot re-enter the same lock.
func DistributedLockOwner(topology Topology) string {
	nodeID := "local-node"
	if topology != nil && strings.TrimSpace(topology.NodeID()) != "" {
		nodeID = strings.TrimSpace(topology.NodeID())
	}
	return nodeID + ":" + guid.S()
}

// BuildSQLSummary derives a count-only SQL preview from the target snapshot.
func BuildSQLSummary(snapshot *store.ManifestSnapshot) SQLSummary {
	if snapshot == nil {
		return SQLSummary{}
	}
	return SQLSummary{
		InstallSQLCount:      snapshot.InstallSQLCount,
		UninstallSQLCount:    snapshot.UninstallSQLCount,
		MockSQLCount:         snapshot.MockSQLCount,
		RuntimeSQLAssetCount: snapshot.RuntimeSQLAssetCount,
	}
}

// BuildHostServicesDiff compares effective and target requested hostServices at service level.
func BuildHostServicesDiff(
	fromSnapshot *store.ManifestSnapshot,
	toSnapshot *store.ManifestSnapshot,
) (HostServicesDiff, error) {
	fromServices, err := normalizeHostServices(fromSnapshot, requestedHostServices(fromSnapshot))
	if err != nil {
		return HostServicesDiff{}, err
	}
	toServices, err := normalizeHostServices(toSnapshot, requestedHostServices(toSnapshot))
	if err != nil {
		return HostServicesDiff{}, err
	}
	diff := HostServicesDiff{}
	diff.AuthorizationRequired = toSnapshot != nil && toSnapshot.HostServiceAuthRequired

	serviceSet := make(map[string]struct{}, len(fromServices)+len(toServices))
	for service := range fromServices {
		serviceSet[service] = struct{}{}
	}
	for service := range toServices {
		serviceSet[service] = struct{}{}
	}
	services := make([]string, 0, len(serviceSet))
	for service := range serviceSet {
		services = append(services, service)
	}
	sort.Strings(services)

	for _, service := range services {
		fromSpec := fromServices[service]
		toSpec := toServices[service]
		change := buildHostServiceChange(service, fromSpec, toSpec)
		switch {
		case fromSpec == nil && toSpec != nil:
			diff.Added = append(diff.Added, change)
		case fromSpec != nil && toSpec == nil:
			diff.Removed = append(diff.Removed, change)
		case hostServiceChanged(fromSpec, toSpec):
			diff.Changed = append(diff.Changed, change)
		}
	}
	diff.AuthorizationChanged = len(diff.Added) > 0 || len(diff.Removed) > 0 || len(diff.Changed) > 0
	return diff, nil
}

// BuildRiskHints returns stable i18n keys for operator risk hints.
func BuildRiskHints(
	sqlSummary SQLSummary,
	hostServicesDiff HostServicesDiff,
	dependencyBlocked bool,
	keys RiskHintKeys,
) []string {
	hints := make([]string, 0, 4)
	if sqlSummary.InstallSQLCount > 0 || sqlSummary.RuntimeSQLAssetCount > 0 {
		hints = append(hints, keys.UpgradeSQLRequiresReview)
	}
	if sqlSummary.MockSQLCount > 0 {
		hints = append(hints, keys.MockSQLExcluded)
	}
	if hostServicesDiff.AuthorizationChanged || hostServicesDiff.AuthorizationRequired {
		hints = append(hints, keys.HostServiceAuthorizationChanged)
	}
	if dependencyBlocked {
		hints = append(hints, keys.DependencyBlockers)
	}
	return hints
}

// requestedHostServices returns requested hostServices from one snapshot.
func requestedHostServices(snapshot *store.ManifestSnapshot) []*protocol.HostServiceSpec {
	if snapshot == nil {
		return nil
	}
	return snapshot.RequestedHostServices
}

// normalizeHostServices normalizes hostServices into one map.
func normalizeHostServices(
	snapshot *store.ManifestSnapshot,
	specs []*protocol.HostServiceSpec,
) (map[string]*protocol.HostServiceSpec, error) {
	normalized, err := protocol.NormalizeHostServiceSpecsForPlugin(snapshotPluginID(snapshot), specs)
	if err != nil {
		return nil, &SnapshotInvalidError{
			PluginID: snapshotPluginID(snapshot),
			Version:  snapshotVersion(snapshot),
			Cause:    err,
		}
	}
	result := make(map[string]*protocol.HostServiceSpec, len(normalized))
	for _, spec := range normalized {
		if spec == nil || strings.TrimSpace(spec.Service) == "" {
			continue
		}
		result[strings.TrimSpace(spec.Service)] = spec
	}
	return result, nil
}

// snapshotPluginID safely extracts the plugin ID for diagnostics.
func snapshotPluginID(snapshot *store.ManifestSnapshot) string {
	if snapshot == nil {
		return ""
	}
	return snapshot.ID
}

// snapshotVersion safely extracts the plugin version for diagnostics.
func snapshotVersion(snapshot *store.ManifestSnapshot) string {
	if snapshot == nil {
		return ""
	}
	return snapshot.Version
}

// buildHostServiceChange converts two service specs into a stable summary for API projection.
func buildHostServiceChange(
	service string,
	fromSpec *protocol.HostServiceSpec,
	toSpec *protocol.HostServiceSpec,
) *HostServiceChange {
	return &HostServiceChange{
		Service:           service,
		FromMethods:       cloneSortedStrings(hostServiceMethods(fromSpec)),
		ToMethods:         cloneSortedStrings(hostServiceMethods(toSpec)),
		FromResourceCount: len(hostServiceResources(fromSpec)),
		ToResourceCount:   len(hostServiceResources(toSpec)),
		FromTables:        cloneSortedStrings(hostServiceTables(fromSpec)),
		ToTables:          cloneSortedStrings(hostServiceTables(toSpec)),
		FromPaths:         cloneSortedStrings(hostServicePaths(fromSpec)),
		ToPaths:           cloneSortedStrings(hostServicePaths(toSpec)),
		FromKeys:          cloneSortedStrings(hostServiceKeys(fromSpec)),
		ToKeys:            cloneSortedStrings(hostServiceKeys(toSpec)),
	}
}

// hostServiceChanged reports whether two normalized specs differ in methods or governed targets.
func hostServiceChanged(fromSpec *protocol.HostServiceSpec, toSpec *protocol.HostServiceSpec) bool {
	if fromSpec == nil || toSpec == nil {
		return fromSpec != toSpec
	}
	if !stringSlicesEqual(hostServiceMethods(fromSpec), hostServiceMethods(toSpec)) {
		return true
	}
	if !stringSlicesEqual(hostServiceTables(fromSpec), hostServiceTables(toSpec)) {
		return true
	}
	if !stringSlicesEqual(hostServicePaths(fromSpec), hostServicePaths(toSpec)) {
		return true
	}
	if !stringSlicesEqual(hostServiceKeys(fromSpec), hostServiceKeys(toSpec)) {
		return true
	}
	return !stringSlicesEqual(hostServiceResourceRefs(fromSpec), hostServiceResourceRefs(toSpec))
}

// hostServiceMethods returns normalized methods from one spec.
func hostServiceMethods(spec *protocol.HostServiceSpec) []string {
	if spec == nil {
		return nil
	}
	return spec.Methods
}

// hostServiceTables returns normalized data tables from one spec.
func hostServiceTables(spec *protocol.HostServiceSpec) []string {
	if spec == nil {
		return nil
	}
	return spec.Tables
}

// hostServicePaths returns normalized storage paths from one spec.
func hostServicePaths(spec *protocol.HostServiceSpec) []string {
	if spec == nil {
		return nil
	}
	return spec.Paths
}

// hostServiceKeys returns normalized authorized host config keys from one spec.
func hostServiceKeys(spec *protocol.HostServiceSpec) []string {
	if spec == nil {
		return nil
	}
	return spec.Keys
}

// hostServiceResources returns governed resources from one spec.
func hostServiceResources(spec *protocol.HostServiceSpec) []*protocol.HostServiceResourceSpec {
	if spec == nil {
		return nil
	}
	return spec.Resources
}

// hostServiceResourceRefs returns normalized resource refs from one spec.
func hostServiceResourceRefs(spec *protocol.HostServiceSpec) []string {
	resources := hostServiceResources(spec)
	refs := make([]string, 0, len(resources))
	for _, resource := range resources {
		if resource == nil || strings.TrimSpace(resource.Ref) == "" {
			continue
		}
		refs = append(refs, strings.TrimSpace(resource.Ref))
	}
	sort.Strings(refs)
	return refs
}

// cloneSortedStrings copies and sorts a string slice before exposing it.
func cloneSortedStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		normalized := strings.TrimSpace(value)
		if normalized == "" {
			continue
		}
		out = append(out, normalized)
	}
	sort.Strings(out)
	return out
}

// stringSlicesEqual compares slices as normalized sorted sets.
func stringSlicesEqual(left []string, right []string) bool {
	leftValues := cloneSortedStrings(left)
	rightValues := cloneSortedStrings(right)
	if len(leftValues) != len(rightValues) {
		return false
	}
	for index := range leftValues {
		if leftValues[index] != rightValues[index] {
			return false
		}
	}
	return true
}
