// This file contains scheduled-job enum validation and i18n key normalization
// helpers used by management and registry code.

package jobmeta

import (
	"strings"
)

// messageKeyPath converts backend anchors into dotted i18n key fragments.
func messageKeyPath(value string) string {
	trimmed := strings.ToLower(strings.TrimSpace(value))
	if trimmed == "" {
		return ""
	}

	var builder strings.Builder
	previousDot := false
	for _, current := range trimmed {
		switch {
		case current >= 'a' && current <= 'z',
			current >= '0' && current <= '9',
			current == '-':
			builder.WriteRune(current)
			previousDot = false
		default:
			if !previousDot {
				builder.WriteRune('.')
				previousDot = true
			}
		}
	}
	return strings.Trim(builder.String(), ".")
}

// IsValid reports whether the task type is supported.
func (t TaskType) IsValid() bool {
	switch t {
	case TaskTypeHandler, TaskTypeShell:
		return true
	}
	return false
}

// IsValid reports whether the job scope is supported.
func (s JobScope) IsValid() bool {
	switch s {
	case JobScopeMasterOnly, JobScopeAllNode:
		return true
	}
	return false
}

// IsValid reports whether the job concurrency mode is supported.
func (c JobConcurrency) IsValid() bool {
	switch c {
	case JobConcurrencySingleton, JobConcurrencyParallel:
		return true
	}
	return false
}

// IsValid reports whether the job status is supported.
func (s JobStatus) IsValid() bool {
	switch s {
	case JobStatusEnabled, JobStatusDisabled, JobStatusPausedByPlugin:
		return true
	}
	return false
}

// IsValid reports whether the trigger type is supported.
func (t TriggerType) IsValid() bool {
	switch t {
	case TriggerTypeCron, TriggerTypeManual:
		return true
	}
	return false
}

// IsValid reports whether the log status is supported.
func (s LogStatus) IsValid() bool {
	switch s {
	case LogStatusRunning,
		LogStatusSuccess,
		LogStatusFailed,
		LogStatusCancelled,
		LogStatusTimeout,
		LogStatusSkippedNotPrimary,
		LogStatusSkippedSingleton,
		LogStatusSkippedMaxConcurrency:
		return true
	}
	return false
}

// IsValid reports whether the handler source is supported.
func (s HandlerSource) IsValid() bool {
	switch s {
	case HandlerSourceHost, HandlerSourcePlugin:
		return true
	}
	return false
}

// IsValid reports whether the retention mode is supported.
func (m RetentionMode) IsValid() bool {
	switch m {
	case RetentionModeDays, RetentionModeCount, RetentionModeNone:
		return true
	}
	return false
}
