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

	"lina-core/internal/service/cluster"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/store"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

const (
	// distributedLockLease bounds orphaned upgrade locks after a node crash while
	// remaining longer than normal plugin SQL/governance phases.
	distributedLockLease = 30 * time.Minute
	// distributedLockReason records the owner purpose in coordination backends.
	distributedLockReason = "plugin-runtime-upgrade"
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

// HostServicesDiff summarizes owner/service/version-level hostServices drift.
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

// HostServiceChange summarizes one owner/service/version-level hostServices change.
type HostServiceChange struct {
	// Owner is the owner plugin ID for plugin-owned host services.
	Owner string
	// Service is the logical host service identifier.
	Service string
	// Version is the owner capability protocol version for plugin-owned host services.
	Version string
	// FromMethods is the effective method set before upgrade.
	FromMethods []string
	// ToMethods is the target method set after upgrade.
	ToMethods []string
	// FromResourceCount is the number of governed targets before upgrade.
	FromResourceCount int
	// ToResourceCount is the number of governed targets after upgrade.
	ToResourceCount int
	// FromResourceRefs is the effective governed resource ref set before upgrade.
	FromResourceRefs []string
	// ToResourceRefs is the target governed resource ref set after upgrade.
	ToResourceRefs []string
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

// canExecute reports whether the explicit management upgrade endpoint may run
// side effects for a runtime-upgrade state.
func canExecute(state plugintypes.RuntimeUpgradeState) bool {
	return state == plugintypes.RuntimeUpgradeStatePendingUpgrade ||
		state == plugintypes.RuntimeUpgradeStateUpgradeFailed
}

// distributedLockName builds the cluster-wide lock name for one plugin upgrade.
func distributedLockName(pluginID string) string {
	return "plugin-runtime-upgrade:" + strings.TrimSpace(pluginID)
}

// distributedLockOwner builds a unique owner for one acquisition so concurrent
// requests from the same node cannot re-enter the same lock.
func distributedLockOwner(topology cluster.Service) string {
	nodeID := "local-node"
	if topology != nil && strings.TrimSpace(topology.NodeID()) != "" {
		nodeID = strings.TrimSpace(topology.NodeID())
	}
	return nodeID + ":" + guid.S()
}

// buildSQLSummary derives a count-only SQL preview from the target snapshot.
func buildSQLSummary(snapshot *store.ManifestSnapshot) SQLSummary {
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

// buildHostServicesDiff compares effective and target requested hostServices at owner/service/version level.
func buildHostServicesDiff(
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

	identitySet := make(map[string]struct{}, len(fromServices)+len(toServices))
	for identity := range fromServices {
		identitySet[identity] = struct{}{}
	}
	for identity := range toServices {
		identitySet[identity] = struct{}{}
	}
	identities := make([]string, 0, len(identitySet))
	for identity := range identitySet {
		identities = append(identities, identity)
	}
	sort.Strings(identities)

	for _, identity := range identities {
		var (
			fromSpec = fromServices[identity]
			toSpec   = toServices[identity]
			change   = buildHostServiceChange(fromSpec, toSpec)
		)
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

// buildRiskHints returns stable i18n keys for operator risk hints.
func buildRiskHints(
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

// normalizeHostServices normalizes hostServices into one identity-keyed map.
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
		result[protocol.HostServiceSpecIdentity(spec)] = spec
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
	fromSpec *protocol.HostServiceSpec,
	toSpec *protocol.HostServiceSpec,
) *HostServiceChange {
	identitySpec := toSpec
	if identitySpec == nil {
		identitySpec = fromSpec
	}
	return &HostServiceChange{
		Owner:             hostServiceOwner(identitySpec),
		Service:           hostServiceName(identitySpec),
		Version:           hostServiceVersion(identitySpec),
		FromMethods:       cloneSortedStrings(hostServiceMethods(fromSpec)),
		ToMethods:         cloneSortedStrings(hostServiceMethods(toSpec)),
		FromResourceCount: len(hostServiceResources(fromSpec)),
		ToResourceCount:   len(hostServiceResources(toSpec)),
		FromResourceRefs:  cloneSortedStrings(hostServiceResourceRefs(fromSpec)),
		ToResourceRefs:    cloneSortedStrings(hostServiceResourceRefs(toSpec)),
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
	if protocol.HostServiceSpecIdentity(fromSpec) != protocol.HostServiceSpecIdentity(toSpec) {
		return true
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

// hostServiceOwner returns the normalized owner plugin ID from one spec.
func hostServiceOwner(spec *protocol.HostServiceSpec) string {
	if spec == nil {
		return ""
	}
	return strings.TrimSpace(spec.Owner)
}

// hostServiceName returns the normalized service name from one spec.
func hostServiceName(spec *protocol.HostServiceSpec) string {
	if spec == nil {
		return ""
	}
	return strings.TrimSpace(spec.Service)
}

// hostServiceVersion returns the normalized owner capability version from one spec.
func hostServiceVersion(spec *protocol.HostServiceSpec) string {
	if spec == nil {
		return ""
	}
	return strings.TrimSpace(spec.Version)
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
