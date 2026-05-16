// This file re-exports catalog-owned runtime-upgrade state types for runtime
// list projections without coupling callers to catalog internals.

package runtime

import "lina-core/internal/service/plugin/internal/catalog"

type (
	// RuntimeUpgradeState identifies whether discovered plugin files match the effective state.
	RuntimeUpgradeState = catalog.RuntimeUpgradeState
	// RuntimeUpgradeAbnormalReason identifies why a plugin cannot be treated as normally upgradeable.
	RuntimeUpgradeAbnormalReason = catalog.RuntimeUpgradeAbnormalReason
	// RuntimeUpgradeFailurePhase identifies the phase associated with the latest failure.
	RuntimeUpgradeFailurePhase = catalog.RuntimeUpgradeFailurePhase
	// RuntimeUpgradeFailure exposes the latest observable runtime-upgrade failure.
	RuntimeUpgradeFailure = catalog.RuntimeUpgradeFailure
)

const (
	// RuntimeUpgradeStateNormal means the effective and discovered metadata are aligned.
	RuntimeUpgradeStateNormal = catalog.RuntimeUpgradeStateNormal
	// RuntimeUpgradeStatePendingUpgrade means discovered files are newer than the effective version.
	RuntimeUpgradeStatePendingUpgrade = catalog.RuntimeUpgradeStatePendingUpgrade
	// RuntimeUpgradeStateAbnormal means discovered files are older or cannot be safely compared.
	RuntimeUpgradeStateAbnormal = catalog.RuntimeUpgradeStateAbnormal
	// RuntimeUpgradeStateUpgradeRunning means a runtime upgrade transition is reconciling.
	RuntimeUpgradeStateUpgradeRunning = catalog.RuntimeUpgradeStateUpgradeRunning
	// RuntimeUpgradeStateUpgradeFailed means the latest target release failed before becoming effective.
	RuntimeUpgradeStateUpgradeFailed = catalog.RuntimeUpgradeStateUpgradeFailed
	// RuntimeUpgradeAbnormalReasonDiscoveredVersionLowerThanEffective means the file version is lower than the DB version.
	RuntimeUpgradeAbnormalReasonDiscoveredVersionLowerThanEffective = catalog.RuntimeUpgradeAbnormalReasonDiscoveredVersionLowerThanEffective
	// RuntimeUpgradeAbnormalReasonVersionCompareFailed means at least one version string is not semver-compatible.
	RuntimeUpgradeAbnormalReasonVersionCompareFailed = catalog.RuntimeUpgradeAbnormalReasonVersionCompareFailed
)
