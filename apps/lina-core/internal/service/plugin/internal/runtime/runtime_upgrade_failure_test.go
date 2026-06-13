// This file verifies structured dynamic-upgrade failure phase propagation
// remains readable after rollback diagnostics wrap or join the original error.

package runtime

import (
	"errors"
	"testing"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/service/plugin/internal/plugintypes"
)

// TestDynamicUpgradeFailurePhaseSurvivesJoinedDiagnostics verifies rollback
// diagnostics do not hide the original upgrade phase from the unified upgrade
// owner.
func TestDynamicUpgradeFailurePhaseSurvivesJoinedDiagnostics(t *testing.T) {
	original := tagDynamicUpgradeFailure(
		plugintypes.RuntimeUpgradeFailurePhaseSQL,
		gerror.New("upgrade SQL failed"),
	)
	joined := errors.Join(original, gerror.New("rollback SQL failed"))

	if phase := DynamicUpgradeFailurePhase(joined); phase != plugintypes.RuntimeUpgradeFailurePhaseSQL {
		t.Fatalf("expected SQL failure phase through joined diagnostics, got %s", phase)
	}
}

// TestTagDynamicUpgradeFailureKeepsExistingPhase verifies outer wrappers do not
// overwrite a more precise phase reported by a deeper operation.
func TestTagDynamicUpgradeFailureKeepsExistingPhase(t *testing.T) {
	original := tagDynamicUpgradeFailure(
		plugintypes.RuntimeUpgradeFailurePhaseUpgradeCallback,
		gerror.New("upgrade callback failed"),
	)
	wrapped := tagDynamicUpgradeFailure(plugintypes.RuntimeUpgradeFailurePhaseRelease, original)

	if phase := DynamicUpgradeFailurePhase(wrapped); phase != plugintypes.RuntimeUpgradeFailurePhaseUpgradeCallback {
		t.Fatalf("expected original callback phase to survive, got %s", phase)
	}
}
