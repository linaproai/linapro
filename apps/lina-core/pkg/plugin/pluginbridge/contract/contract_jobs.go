// This file defines dynamic-plugin and source-plugin scheduled-job declaration
// metadata shared by guest registration, host discovery, and handler projection.

package contract

import (
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
)

const (
	// DefaultJobContractTimezone is the fallback timezone applied to declared
	// dynamic-plugin jobs when the artifact omits an explicit timezone.
	DefaultJobContractTimezone = "Asia/Shanghai"
	// DefaultJobContractTimeoutSeconds is the fallback timeout used for one
	// dynamic-plugin job execution when the artifact omits an explicit value.
	DefaultJobContractTimeoutSeconds = 300
	// DeclaredJobRouteBasePath is the synthetic runtime route prefix reserved
	// for declared dynamic-plugin jobs.
	DeclaredJobRouteBasePath = "/@jobs"
	// DeclaredJobRegistrationInternalPath is the reserved guest controller
	// internal path invoked by the host to collect dynamic-plugin job declarations.
	DeclaredJobRegistrationInternalPath = "/register-jobs"
	// DeclaredJobRegistrationRequestType is the reflected guest request type
	// name used by the default guest controller dispatcher for Jobs discovery.
	DeclaredJobRegistrationRequestType = "RegisterJobsReq"
)

// JobScope identifies where one declared plugin job is allowed to run.
type JobScope string

// Supported plugin job scope values.
const (
	// JobScopeMasterOnly limits execution to the primary node.
	JobScopeMasterOnly JobScope = "master_only"
	// JobScopeAllNode allows execution on every node.
	JobScopeAllNode JobScope = "all_node"
)

// String returns the canonical job scope value.
func (s JobScope) String() string {
	return string(s)
}

// IsValid reports whether the job scope is supported.
func (s JobScope) IsValid() bool {
	switch s {
	case JobScopeMasterOnly, JobScopeAllNode:
		return true
	default:
		return false
	}
}

// JobConcurrency identifies the overlap policy for one declared plugin job.
type JobConcurrency string

// Supported plugin job concurrency values.
const (
	// JobConcurrencySingleton skips overlapping executions.
	JobConcurrencySingleton JobConcurrency = "singleton"
	// JobConcurrencyParallel allows overlaps up to maxConcurrency.
	JobConcurrencyParallel JobConcurrency = "parallel"
)

// String returns the canonical job concurrency value.
func (c JobConcurrency) String() string {
	return string(c)
}

// IsValid reports whether the job concurrency is supported.
func (c JobConcurrency) IsValid() bool {
	switch c {
	case JobConcurrencySingleton, JobConcurrencyParallel:
		return true
	default:
		return false
	}
}

// JobContract defines one dynamic-plugin built-in scheduled-job declaration
// registered from guest code through the governed Jobs host service.
type JobContract struct {
	// Name is the stable plugin-local job identifier.
	Name string `json:"name" yaml:"name"`
	// DisplayName is the UI-facing job title shown in task management.
	DisplayName string `json:"displayName,omitempty" yaml:"displayName,omitempty"`
	// Description explains the job purpose for operators.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// Pattern is the raw gcron expression declared by the plugin.
	Pattern string `json:"pattern" yaml:"pattern"`
	// Timezone is the optional UI display timezone for cron-style patterns.
	Timezone string `json:"timezone,omitempty" yaml:"timezone,omitempty"`
	// Scope selects master-only or all-node execution.
	Scope JobScope `json:"scope,omitempty" yaml:"scope,omitempty"`
	// Concurrency selects singleton or parallel overlap handling.
	Concurrency JobConcurrency `json:"concurrency,omitempty" yaml:"concurrency,omitempty"`
	// MaxConcurrency limits overlaps when Concurrency=parallel.
	MaxConcurrency int `json:"maxConcurrency,omitempty" yaml:"maxConcurrency,omitempty"`
	// TimeoutSeconds bounds one execution in whole seconds.
	TimeoutSeconds int `json:"timeoutSeconds,omitempty" yaml:"timeoutSeconds,omitempty"`
	// RequestType is the reflected guest request type used by the guest controller dispatcher.
	RequestType string `json:"requestType" yaml:"requestType"`
	// InternalPath is the optional guest-internal route metadata for the scheduled job.
	InternalPath string `json:"internalPath,omitempty" yaml:"internalPath,omitempty"`
}

// NormalizeJobScope normalizes one raw job scope string.
func NormalizeJobScope(value string) JobScope {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "", JobScopeAllNode.String():
		return JobScopeAllNode
	case JobScopeMasterOnly.String():
		return JobScopeMasterOnly
	default:
		return ""
	}
}

// NormalizeJobConcurrency normalizes one raw job concurrency string.
func NormalizeJobConcurrency(value string) JobConcurrency {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "", JobConcurrencySingleton.String():
		return JobConcurrencySingleton
	case JobConcurrencyParallel.String():
		return JobConcurrencyParallel
	default:
		return ""
	}
}

// NormalizeJobContract normalizes one declared job contract in place.
func NormalizeJobContract(contract *JobContract) {
	if contract == nil {
		return
	}
	contract.Name = strings.TrimSpace(contract.Name)
	contract.DisplayName = strings.TrimSpace(contract.DisplayName)
	contract.Description = strings.TrimSpace(contract.Description)
	contract.Pattern = strings.TrimSpace(contract.Pattern)
	contract.Timezone = strings.TrimSpace(contract.Timezone)
	if contract.Timezone == "" {
		contract.Timezone = DefaultJobContractTimezone
	}
	contract.Scope = NormalizeJobScope(contract.Scope.String())
	contract.Concurrency = NormalizeJobConcurrency(contract.Concurrency.String())
	if contract.MaxConcurrency <= 0 {
		contract.MaxConcurrency = 1
	}
	if contract.TimeoutSeconds <= 0 {
		contract.TimeoutSeconds = DefaultJobContractTimeoutSeconds
	}
	contract.RequestType = strings.TrimSpace(contract.RequestType)
	contract.InternalPath = strings.TrimSpace(contract.InternalPath)
	if contract.InternalPath != "" && !strings.HasPrefix(contract.InternalPath, "/") {
		contract.InternalPath = "/" + contract.InternalPath
	}
}

// BuildPluginJobHandlerRef returns the synthetic handler reference used for
// one plugin-owned built-in scheduled job.
func BuildPluginJobHandlerRef(pluginID string, name string) (string, error) {
	trimmedPluginID := strings.TrimSpace(pluginID)
	trimmedName := strings.TrimSpace(name)
	if trimmedPluginID == "" {
		return "", gerror.New("plugin ID cannot be empty")
	}
	if trimmedName == "" {
		return "", gerror.New("plugin built-in job name cannot be empty")
	}
	return fmt.Sprintf("plugin:%s/jobs:%s", trimmedPluginID, trimmedName), nil
}

// BuildDeclaredJobRoutePath returns the synthetic runtime route path used to
// execute one declared dynamic-plugin job through the bridge.
func BuildDeclaredJobRoutePath(contract *JobContract) string {
	if contract == nil {
		return DeclaredJobRouteBasePath
	}
	if internalPath := strings.TrimSpace(contract.InternalPath); internalPath != "" {
		if strings.HasPrefix(internalPath, "/") {
			return internalPath
		}
		return "/" + internalPath
	}
	trimmedName := strings.TrimSpace(contract.Name)
	if trimmedName == "" {
		return DeclaredJobRouteBasePath
	}
	return DeclaredJobRouteBasePath + "/" + trimmedName
}

// ValidateJobContracts validates one plugin's declared job contracts in place.
func ValidateJobContracts(pluginID string, contracts []*JobContract) error {
	seen := make(map[string]struct{}, len(contracts))
	for _, contract := range contracts {
		if contract == nil {
			return gerror.New("dynamic plugin job declaration cannot be nil")
		}
		NormalizeJobContract(contract)
		if contract.Name == "" {
			return gerror.Newf("dynamic plugin %s job is missing name", strings.TrimSpace(pluginID))
		}
		if contract.Pattern == "" {
			return gerror.Newf("dynamic plugin %s job %s is missing pattern", strings.TrimSpace(pluginID), contract.Name)
		}
		if len(contract.Pattern) > 128 {
			return gerror.Newf("dynamic plugin %s job %s pattern cannot exceed 128 characters", strings.TrimSpace(pluginID), contract.Name)
		}
		if !contract.Scope.IsValid() {
			return gerror.Newf("dynamic plugin %s job %s has invalid scope", strings.TrimSpace(pluginID), contract.Name)
		}
		if !contract.Concurrency.IsValid() {
			return gerror.Newf("dynamic plugin %s job %s has invalid concurrency", strings.TrimSpace(pluginID), contract.Name)
		}
		if contract.TimeoutSeconds <= 0 || contract.TimeoutSeconds > int((24*time.Hour).Seconds()) {
			return gerror.Newf("dynamic plugin %s job %s timeoutSeconds is out of range", strings.TrimSpace(pluginID), contract.Name)
		}
		if contract.Timezone != "" {
			if _, err := time.LoadLocation(contract.Timezone); err != nil {
				return gerror.Newf(
					"dynamic plugin %s job %s has invalid timezone: %s",
					strings.TrimSpace(pluginID),
					contract.Name,
					contract.Timezone,
				)
			}
		}
		if contract.RequestType == "" {
			return gerror.Newf(
				"dynamic plugin %s job %s must declare requestType",
				strings.TrimSpace(pluginID),
				contract.Name,
			)
		}
		if _, ok := seen[contract.Name]; ok {
			return gerror.Newf("dynamic plugin %s job name is duplicated: %s", strings.TrimSpace(pluginID), contract.Name)
		}
		seen[contract.Name] = struct{}{}
	}
	return nil
}
