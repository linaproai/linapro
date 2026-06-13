// This file re-exports catalog-owned runtime-upgrade state types for runtime
// list projections without coupling callers to catalog internals.

package runtime

import "lina-core/internal/service/plugin/internal/plugintypes"

type (
	// RuntimeUpgradeState identifies whether discovered plugin files match the effective state.
	RuntimeUpgradeState = plugintypes.RuntimeUpgradeState
	// RuntimeUpgradeAbnormalReason identifies why a plugin cannot be treated as normally upgradeable.
	RuntimeUpgradeAbnormalReason = plugintypes.RuntimeUpgradeAbnormalReason
	// RuntimeUpgradeFailurePhase identifies the phase associated with the latest failure.
	RuntimeUpgradeFailurePhase = plugintypes.RuntimeUpgradeFailurePhase
	// RuntimeUpgradeFailure exposes the latest observable runtime-upgrade failure.
	RuntimeUpgradeFailure = plugintypes.RuntimeUpgradeFailure
)

const (
	// RuntimeUpgradeStateNormal means the effective and discovered metadata are aligned.
	RuntimeUpgradeStateNormal = plugintypes.RuntimeUpgradeStateNormal
	// RuntimeUpgradeStatePendingUpgrade means discovered files are newer than the effective version.
	RuntimeUpgradeStatePendingUpgrade = plugintypes.RuntimeUpgradeStatePendingUpgrade
	// RuntimeUpgradeStateAbnormal means discovered files are older or cannot be safely compared.
	RuntimeUpgradeStateAbnormal = plugintypes.RuntimeUpgradeStateAbnormal
	// RuntimeUpgradeStateUpgradeRunning means a runtime upgrade transition is reconciling.
	RuntimeUpgradeStateUpgradeRunning = plugintypes.RuntimeUpgradeStateUpgradeRunning
	// RuntimeUpgradeStateUpgradeFailed means the latest target release failed before becoming effective.
	RuntimeUpgradeStateUpgradeFailed = plugintypes.RuntimeUpgradeStateUpgradeFailed
	// RuntimeUpgradeAbnormalReasonDiscoveredVersionLowerThanEffective means the file version is lower than the DB version.
	RuntimeUpgradeAbnormalReasonDiscoveredVersionLowerThanEffective = plugintypes.RuntimeUpgradeAbnormalReasonDiscoveredVersionLowerThanEffective
	// RuntimeUpgradeAbnormalReasonVersionCompareFailed means at least one version string is not semver-compatible.
	RuntimeUpgradeAbnormalReasonVersionCompareFailed = plugintypes.RuntimeUpgradeAbnormalReasonVersionCompareFailed
)
