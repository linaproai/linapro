// This file defines dynamic-plugin lifecycle bridge contracts.

package contract

import (
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
)

// LifecycleOperation identifies one dynamic-plugin lifecycle callback.
type LifecycleOperation string

// Lifecycle operation constants intentionally mirror the source plugin
// LifecycleHook names exactly.
const (
	// LifecycleOperationBeforeInstall protects plugin install.
	LifecycleOperationBeforeInstall LifecycleOperation = "BeforeInstall"
	// LifecycleOperationAfterInstall observes successful plugin install.
	LifecycleOperationAfterInstall LifecycleOperation = "AfterInstall"
	// LifecycleOperationBeforeUpgrade protects plugin runtime upgrade.
	LifecycleOperationBeforeUpgrade LifecycleOperation = "BeforeUpgrade"
	// LifecycleOperationAfterUpgrade observes successful plugin runtime upgrade.
	LifecycleOperationAfterUpgrade LifecycleOperation = "AfterUpgrade"
	// LifecycleOperationBeforeDisable protects global plugin disable.
	LifecycleOperationBeforeDisable LifecycleOperation = "BeforeDisable"
	// LifecycleOperationAfterDisable observes successful global plugin disable.
	LifecycleOperationAfterDisable LifecycleOperation = "AfterDisable"
	// LifecycleOperationBeforeUninstall protects plugin uninstall.
	LifecycleOperationBeforeUninstall LifecycleOperation = "BeforeUninstall"
	// LifecycleOperationAfterUninstall observes successful plugin uninstall.
	LifecycleOperationAfterUninstall LifecycleOperation = "AfterUninstall"
	// LifecycleOperationBeforeTenantDisable protects tenant-scoped plugin disable.
	LifecycleOperationBeforeTenantDisable LifecycleOperation = "BeforeTenantDisable"
	// LifecycleOperationAfterTenantDisable observes successful tenant-scoped plugin disable.
	LifecycleOperationAfterTenantDisable LifecycleOperation = "AfterTenantDisable"
	// LifecycleOperationBeforeTenantDelete protects tenant deletion.
	LifecycleOperationBeforeTenantDelete LifecycleOperation = "BeforeTenantDelete"
	// LifecycleOperationAfterTenantDelete observes successful tenant deletion.
	LifecycleOperationAfterTenantDelete LifecycleOperation = "AfterTenantDelete"
	// LifecycleOperationBeforeInstallModeChange protects install-mode changes.
	LifecycleOperationBeforeInstallModeChange LifecycleOperation = "BeforeInstallModeChange"
	// LifecycleOperationAfterInstallModeChange observes successful install-mode changes.
	LifecycleOperationAfterInstallModeChange LifecycleOperation = "AfterInstallModeChange"
)

// LifecycleContract describes one executable dynamic-plugin lifecycle handler.
type LifecycleContract struct {
	// Operation is the exact source-compatible lifecycle operation name.
	Operation LifecycleOperation `json:"operation" yaml:"operation"`
	// RequestType is the guest controller request DTO name used for reflection dispatch.
	RequestType string `json:"requestType" yaml:"requestType"`
	// InternalPath is the guest-side path used as a dispatcher fallback.
	InternalPath string `json:"internalPath" yaml:"internalPath"`
	// TimeoutMs optionally overrides the host default per-callback timeout.
	TimeoutMs int `json:"timeoutMs,omitempty" yaml:"timeoutMs,omitempty"`
}

// LifecycleRequest is the JSON body sent to dynamic lifecycle handlers.
type LifecycleRequest struct {
	// PluginID is the lifecycle operation target plugin.
	PluginID string `json:"pluginId"`
	// Operation is the exact source-compatible lifecycle operation name.
	Operation string `json:"operation"`
	// FromVersion is the effective version before upgrade when applicable.
	FromVersion string `json:"fromVersion,omitempty"`
	// ToVersion is the target version for upgrade when applicable.
	ToVersion string `json:"toVersion,omitempty"`
	// TenantID is the tenant affected by tenant-scoped lifecycle operations.
	TenantID int `json:"tenantId,omitempty"`
	// FromMode is the previous install mode for install-mode changes.
	FromMode string `json:"fromMode,omitempty"`
	// ToMode is the target install mode for install-mode changes.
	ToMode string `json:"toMode,omitempty"`
}

// LifecycleDecision is the JSON response body returned by dynamic lifecycle handlers.
type LifecycleDecision struct {
	// OK reports whether the host may continue the protected lifecycle operation.
	OK bool `json:"ok"`
	// Reason is a stable i18n reason key or short diagnostic used when OK is false.
	Reason string `json:"reason,omitempty"`
}

// NormalizeLifecycleOperation converts one raw operation string to its canonical value.
func NormalizeLifecycleOperation(value string) LifecycleOperation {
	switch strings.TrimSpace(value) {
	case LifecycleOperationBeforeInstall.String():
		return LifecycleOperationBeforeInstall
	case LifecycleOperationAfterInstall.String():
		return LifecycleOperationAfterInstall
	case LifecycleOperationBeforeUpgrade.String():
		return LifecycleOperationBeforeUpgrade
	case LifecycleOperationAfterUpgrade.String():
		return LifecycleOperationAfterUpgrade
	case LifecycleOperationBeforeDisable.String():
		return LifecycleOperationBeforeDisable
	case LifecycleOperationAfterDisable.String():
		return LifecycleOperationAfterDisable
	case LifecycleOperationBeforeUninstall.String():
		return LifecycleOperationBeforeUninstall
	case LifecycleOperationAfterUninstall.String():
		return LifecycleOperationAfterUninstall
	case LifecycleOperationBeforeTenantDisable.String():
		return LifecycleOperationBeforeTenantDisable
	case LifecycleOperationAfterTenantDisable.String():
		return LifecycleOperationAfterTenantDisable
	case LifecycleOperationBeforeTenantDelete.String():
		return LifecycleOperationBeforeTenantDelete
	case LifecycleOperationAfterTenantDelete.String():
		return LifecycleOperationAfterTenantDelete
	case LifecycleOperationBeforeInstallModeChange.String():
		return LifecycleOperationBeforeInstallModeChange
	case LifecycleOperationAfterInstallModeChange.String():
		return LifecycleOperationAfterInstallModeChange
	default:
		return LifecycleOperation("")
	}
}

// String returns the canonical lifecycle operation name.
func (operation LifecycleOperation) String() string {
	return string(operation)
}

// IsSupportedLifecycleOperation reports whether value names a supported lifecycle operation.
func IsSupportedLifecycleOperation(value string) bool {
	return NormalizeLifecycleOperation(value) != ""
}

// ValidateLifecycleContracts validates and normalizes dynamic lifecycle declarations.
func ValidateLifecycleContracts(pluginID string, contracts []*LifecycleContract) error {
	seen := make(map[LifecycleOperation]struct{}, len(contracts))
	for _, contract := range contracts {
		if contract == nil {
			return gerror.New("dynamic lifecycle contract cannot be nil")
		}
		NormalizeLifecycleContract(contract)
		if contract.Operation == "" {
			return gerror.Newf("dynamic lifecycle contract operation is unsupported for plugin %s", pluginID)
		}
		if _, ok := seen[contract.Operation]; ok {
			return gerror.Newf("dynamic lifecycle contract operation is duplicated for plugin %s: %s", pluginID, contract.Operation)
		}
		seen[contract.Operation] = struct{}{}
		if strings.TrimSpace(contract.RequestType) == "" {
			return gerror.Newf("dynamic lifecycle contract requestType cannot be empty for plugin %s operation %s", pluginID, contract.Operation)
		}
		if strings.TrimSpace(contract.InternalPath) == "" {
			return gerror.Newf("dynamic lifecycle contract internalPath cannot be empty for plugin %s operation %s", pluginID, contract.Operation)
		}
		if contract.TimeoutMs < 0 {
			return gerror.Newf("dynamic lifecycle contract timeoutMs cannot be negative for plugin %s operation %s", pluginID, contract.Operation)
		}
	}
	return nil
}

// NormalizeLifecycleContract trims and canonicalizes one lifecycle contract in-place.
func NormalizeLifecycleContract(contract *LifecycleContract) {
	if contract == nil {
		return
	}
	contract.Operation = NormalizeLifecycleOperation(contract.Operation.String())
	contract.RequestType = strings.TrimSpace(contract.RequestType)
	contract.InternalPath = normalizeLifecycleInternalPath(contract.InternalPath)
}

// normalizeLifecycleInternalPath returns a canonical absolute internal path.
func normalizeLifecycleInternalPath(value string) string {
	path := strings.TrimSpace(value)
	if path == "" {
		return ""
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if len(path) > 1 {
		path = strings.TrimSuffix(path, "/")
	}
	return path
}
