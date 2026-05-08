// This file defines the shared dynamic-plugin cron declaration contract used by
// guest-side code registration, host-side discovery, and scheduled-job projection.

package contract

import (
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
)

const (
	// DefaultCronContractTimezone is the fallback timezone applied to declared
	// dynamic-plugin cron jobs when the artifact omits an explicit timezone.
	DefaultCronContractTimezone = "Asia/Shanghai"
	// DefaultCronContractTimeoutSeconds is the fallback timeout used for one
	// dynamic-plugin cron execution when the artifact omits an explicit value.
	DefaultCronContractTimeoutSeconds = 300
	// DeclaredCronRouteBasePath is the synthetic runtime route prefix reserved
	// for declared dynamic-plugin cron jobs.
	DeclaredCronRouteBasePath = "/@cron"
	// DeclaredCronRegistrationInternalPath is the reserved guest controller
	// internal path invoked by the host to collect dynamic-plugin cron
	// declarations.
	DeclaredCronRegistrationInternalPath = "/register-crons"
	// DeclaredCronRegistrationRequestType is the reflected guest request type
	// name used by the default guest controller dispatcher for cron discovery.
	DeclaredCronRegistrationRequestType = "RegisterCronsReq"
)

// CronScope identifies where one declared plugin cron job is allowed to run.
type CronScope string

// Supported plugin cron scope values.
const (
	// CronScopeMasterOnly limits execution to the primary node.
	CronScopeMasterOnly CronScope = "master_only"
	// CronScopeAllNode allows execution on every node.
	CronScopeAllNode CronScope = "all_node"
)

// String returns the canonical cron scope value.
func (s CronScope) String() string {
	return string(s)
}

// IsValid reports whether the cron scope is supported.
func (s CronScope) IsValid() bool {
	switch s {
	case CronScopeMasterOnly, CronScopeAllNode:
		return true
	default:
		return false
	}
}

// CronConcurrency identifies the overlap policy for one declared plugin cron
// job.
type CronConcurrency string

// Supported plugin cron concurrency values.
const (
	// CronConcurrencySingleton skips overlapping executions.
	CronConcurrencySingleton CronConcurrency = "singleton"
	// CronConcurrencyParallel allows overlaps up to maxConcurrency.
	CronConcurrencyParallel CronConcurrency = "parallel"
)

// String returns the canonical cron concurrency value.
func (c CronConcurrency) String() string {
	return string(c)
}

// IsValid reports whether the cron concurrency is supported.
func (c CronConcurrency) IsValid() bool {
	switch c {
	case CronConcurrencySingleton, CronConcurrencyParallel:
		return true
	default:
		return false
	}
}

// CronContract defines one dynamic-plugin built-in scheduled-job declaration
// registered from guest code through the governed cron host service.
type CronContract struct {
	// Name is the stable plugin-local cron job identifier.
	Name string `json:"name" yaml:"name"`
	// DisplayName is the UI-facing cron job title shown in task management.
	DisplayName string `json:"displayName,omitempty" yaml:"displayName,omitempty"`
	// Description explains the cron job purpose for operators.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// Pattern is the raw gcron expression declared by the plugin.
	Pattern string `json:"pattern" yaml:"pattern"`
	// Timezone is the optional UI display timezone for cron-style patterns.
	Timezone string `json:"timezone,omitempty" yaml:"timezone,omitempty"`
	// Scope selects master-only or all-node execution.
	Scope CronScope `json:"scope,omitempty" yaml:"scope,omitempty"`
	// Concurrency selects singleton or parallel overlap handling.
	Concurrency CronConcurrency `json:"concurrency,omitempty" yaml:"concurrency,omitempty"`
	// MaxConcurrency limits overlaps when Concurrency=parallel.
	MaxConcurrency int `json:"maxConcurrency,omitempty" yaml:"maxConcurrency,omitempty"`
	// TimeoutSeconds bounds one execution in whole seconds.
	TimeoutSeconds int `json:"timeoutSeconds,omitempty" yaml:"timeoutSeconds,omitempty"`
	// RequestType is the optional reflected guest request type used by the
	// default guest controller dispatcher.
	RequestType string `json:"requestType,omitempty" yaml:"requestType,omitempty"`
	// InternalPath is the optional guest-internal handler path used as a
	// fallback dispatch key for custom guest runtimes.
	InternalPath string `json:"internalPath,omitempty" yaml:"internalPath,omitempty"`
}

// NormalizeCronScope normalizes one raw cron scope string.
func NormalizeCronScope(value string) CronScope {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "", CronScopeAllNode.String():
		return CronScopeAllNode
	case CronScopeMasterOnly.String():
		return CronScopeMasterOnly
	default:
		return ""
	}
}

// NormalizeCronConcurrency normalizes one raw cron concurrency string.
func NormalizeCronConcurrency(value string) CronConcurrency {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "", CronConcurrencySingleton.String():
		return CronConcurrencySingleton
	case CronConcurrencyParallel.String():
		return CronConcurrencyParallel
	default:
		return ""
	}
}

// NormalizeCronContract normalizes one declared cron contract in place.
func NormalizeCronContract(contract *CronContract) {
	if contract == nil {
		return
	}
	contract.Name = strings.TrimSpace(contract.Name)
	contract.DisplayName = strings.TrimSpace(contract.DisplayName)
	contract.Description = strings.TrimSpace(contract.Description)
	contract.Pattern = strings.TrimSpace(contract.Pattern)
	contract.Timezone = strings.TrimSpace(contract.Timezone)
	if contract.Timezone == "" {
		contract.Timezone = DefaultCronContractTimezone
	}
	contract.Scope = NormalizeCronScope(contract.Scope.String())
	contract.Concurrency = NormalizeCronConcurrency(contract.Concurrency.String())
	if contract.MaxConcurrency <= 0 {
		contract.MaxConcurrency = 1
	}
	if contract.TimeoutSeconds <= 0 {
		contract.TimeoutSeconds = DefaultCronContractTimeoutSeconds
	}
	contract.RequestType = strings.TrimSpace(contract.RequestType)
	contract.InternalPath = strings.TrimSpace(contract.InternalPath)
	if contract.InternalPath != "" && !strings.HasPrefix(contract.InternalPath, "/") {
		contract.InternalPath = "/" + contract.InternalPath
	}
}

// BuildPluginCronHandlerRef returns the synthetic handler reference used for
// one plugin-owned built-in cron job.
func BuildPluginCronHandlerRef(pluginID string, name string) (string, error) {
	trimmedPluginID := strings.TrimSpace(pluginID)
	trimmedName := strings.TrimSpace(name)
	if trimmedPluginID == "" {
		return "", gerror.New("plugin ID cannot be empty")
	}
	if trimmedName == "" {
		return "", gerror.New("plugin built-in cron job name cannot be empty")
	}
	return fmt.Sprintf("plugin:%s/cron:%s", trimmedPluginID, trimmedName), nil
}

// BuildDeclaredCronRoutePath returns the synthetic runtime route path used to
// execute one declared dynamic-plugin cron job through the bridge.
func BuildDeclaredCronRoutePath(contract *CronContract) string {
	if contract == nil {
		return DeclaredCronRouteBasePath
	}
	if internalPath := strings.TrimSpace(contract.InternalPath); internalPath != "" {
		if strings.HasPrefix(internalPath, "/") {
			return internalPath
		}
		return "/" + internalPath
	}
	trimmedName := strings.TrimSpace(contract.Name)
	if trimmedName == "" {
		return DeclaredCronRouteBasePath
	}
	return DeclaredCronRouteBasePath + "/" + trimmedName
}

// ValidateCronContracts validates one plugin's declared cron contracts in place.
func ValidateCronContracts(pluginID string, contracts []*CronContract) error {
	seen := make(map[string]struct{}, len(contracts))
	for _, contract := range contracts {
		if contract == nil {
			return gerror.New("dynamic plugin cron declaration cannot be nil")
		}
		NormalizeCronContract(contract)
		if contract.Name == "" {
			return gerror.Newf("dynamic plugin %s cron job is missing name", strings.TrimSpace(pluginID))
		}
		if contract.Pattern == "" {
			return gerror.Newf("dynamic plugin %s cron job %s is missing pattern", strings.TrimSpace(pluginID), contract.Name)
		}
		if len(contract.Pattern) > 128 {
			return gerror.Newf("dynamic plugin %s cron job %s pattern cannot exceed 128 characters", strings.TrimSpace(pluginID), contract.Name)
		}
		if !contract.Scope.IsValid() {
			return gerror.Newf("dynamic plugin %s cron job %s has invalid scope", strings.TrimSpace(pluginID), contract.Name)
		}
		if !contract.Concurrency.IsValid() {
			return gerror.Newf("dynamic plugin %s cron job %s has invalid concurrency", strings.TrimSpace(pluginID), contract.Name)
		}
		if contract.TimeoutSeconds <= 0 || contract.TimeoutSeconds > int((24*time.Hour).Seconds()) {
			return gerror.Newf("dynamic plugin %s cron job %s timeoutSeconds is out of range", strings.TrimSpace(pluginID), contract.Name)
		}
		if contract.Timezone != "" {
			if _, err := time.LoadLocation(contract.Timezone); err != nil {
				return gerror.Newf(
					"dynamic plugin %s cron job %s has invalid timezone: %s",
					strings.TrimSpace(pluginID),
					contract.Name,
					contract.Timezone,
				)
			}
		}
		if contract.RequestType == "" && contract.InternalPath == "" {
			return gerror.Newf(
				"dynamic plugin %s cron job %s must declare requestType or internalPath",
				strings.TrimSpace(pluginID),
				contract.Name,
			)
		}
		if _, ok := seen[contract.Name]; ok {
			return gerror.Newf("dynamic plugin %s cron job name is duplicated: %s", strings.TrimSpace(pluginID), contract.Name)
		}
		seen[contract.Name] = struct{}{}
	}
	return nil
}
