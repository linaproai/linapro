// This file defines structured dynamic-upgrade failure diagnostics returned to
// the unified upgrade owner. Runtime still owns release rollback, while upgrade
// owns migration-ledger projection for management diagnostics.

package runtime

import (
	"errors"

	"lina-core/internal/service/plugin/internal/plugintypes"
)

// dynamicUpgradePhaseError tags a dynamic upgrade failure with the lifecycle
// phase that failed before rollback diagnostics were appended.
type dynamicUpgradePhaseError struct {
	// phase is the stable runtime-upgrade failure phase.
	phase plugintypes.RuntimeUpgradeFailurePhase
	// err is the underlying dynamic upgrade failure.
	err error
}

// Error implements error.
func (e *dynamicUpgradePhaseError) Error() string {
	if e == nil || e.err == nil {
		return ""
	}
	return e.err.Error()
}

// Unwrap returns the underlying failure.
func (e *dynamicUpgradePhaseError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.err
}

// tagDynamicUpgradeFailure annotates one dynamic upgrade error with a stable
// phase unless the error already carries a phase from a deeper operation.
func tagDynamicUpgradeFailure(
	phase plugintypes.RuntimeUpgradeFailurePhase,
	err error,
) error {
	if err == nil {
		return nil
	}
	if DynamicUpgradeFailurePhase(err) != "" {
		return err
	}
	return &dynamicUpgradePhaseError{phase: phase, err: err}
}

// DynamicUpgradeFailurePhase returns the stable failure phase tagged on a
// dynamic upgrade error. Empty means the caller should fall back to release.
func DynamicUpgradeFailurePhase(err error) plugintypes.RuntimeUpgradeFailurePhase {
	var phaseErr *dynamicUpgradePhaseError
	if errors.As(err, &phaseErr) && phaseErr != nil {
		return phaseErr.phase
	}
	return ""
}
