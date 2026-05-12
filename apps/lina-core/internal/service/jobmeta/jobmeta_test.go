// This file tests shared scheduled-job domain enums, i18n key helpers,
// and retention policy parsing/serialization.

package jobmeta

import (
	"testing"

	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

// TestNormalizeFunctions verifies every Normalize* helper trims whitespace
// and preserves the underlying type.
func TestNormalizeFunctions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		fn       func(string) any
		input    string
		expected string
	}{
		{"TaskType trims whitespace", func(s string) any { return string(NormalizeTaskType(s)) }, "  handler  ", "handler"},
		{"TaskType preserves valid", func(s string) any { return string(NormalizeTaskType(s)) }, "shell", "shell"},
		{"TaskType handles empty", func(s string) any { return string(NormalizeTaskType(s)) }, "", ""},

		{"JobScope trims whitespace", func(s string) any { return string(NormalizeJobScope(s)) }, "  all_node  ", "all_node"},
		{"JobScope preserves valid", func(s string) any { return string(NormalizeJobScope(s)) }, "master_only", "master_only"},

		{"JobConcurrency trims whitespace", func(s string) any { return string(NormalizeJobConcurrency(s)) }, "  parallel  ", "parallel"},

		{"JobStatus trims whitespace", func(s string) any { return string(NormalizeJobStatus(s)) }, "  disabled  ", "disabled"},

		{"TriggerType trims whitespace", func(s string) any { return string(NormalizeTriggerType(s)) }, "  manual  ", "manual"},

		{"LogStatus trims whitespace", func(s string) any { return string(NormalizeLogStatus(s)) }, "  success  ", "success"},

		{"HandlerSource trims whitespace", func(s string) any { return string(NormalizeHandlerSource(s)) }, "  plugin  ", "plugin"},

		{"RetentionMode trims whitespace", func(s string) any { return string(NormalizeRetentionMode(s)) }, "  days  ", "days"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.fn(tc.input); got != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}

// TestTaskType_IsValid verifies the TaskType validity check.
func TestTaskType_IsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    TaskType
		expected bool
	}{
		{"handler is valid", TaskTypeHandler, true},
		{"shell is valid", TaskTypeShell, true},
		{"empty is invalid", TaskType(""), false},
		{"unknown is invalid", TaskType("cron"), false},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.value.IsValid(); got != tc.expected {
				t.Fatalf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}

// TestJobScope_IsValid verifies the JobScope validity check.
func TestJobScope_IsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    JobScope
		expected bool
	}{
		{"master_only is valid", JobScopeMasterOnly, true},
		{"all_node is valid", JobScopeAllNode, true},
		{"empty is invalid", JobScope(""), false},
		{"unknown is invalid", JobScope("single_node"), false},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.value.IsValid(); got != tc.expected {
				t.Fatalf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}

// TestJobConcurrency_IsValid verifies the JobConcurrency validity check.
func TestJobConcurrency_IsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    JobConcurrency
		expected bool
	}{
		{"singleton is valid", JobConcurrencySingleton, true},
		{"parallel is valid", JobConcurrencyParallel, true},
		{"empty is invalid", JobConcurrency(""), false},
		{"unknown is invalid", JobConcurrency("exclusive"), false},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.value.IsValid(); got != tc.expected {
				t.Fatalf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}

// TestJobStatus_IsValid verifies the JobStatus validity check.
func TestJobStatus_IsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    JobStatus
		expected bool
	}{
		{"enabled is valid", JobStatusEnabled, true},
		{"disabled is valid", JobStatusDisabled, true},
		{"paused_by_plugin is valid", JobStatusPausedByPlugin, true},
		{"empty is invalid", JobStatus(""), false},
		{"unknown is invalid", JobStatus("running"), false},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.value.IsValid(); got != tc.expected {
				t.Fatalf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}

// TestTriggerType_IsValid verifies the TriggerType validity check.
func TestTriggerType_IsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    TriggerType
		expected bool
	}{
		{"cron is valid", TriggerTypeCron, true},
		{"manual is valid", TriggerTypeManual, true},
		{"empty is invalid", TriggerType(""), false},
		{"unknown is invalid", TriggerType("scheduled"), false},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.value.IsValid(); got != tc.expected {
				t.Fatalf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}

// TestLogStatus_IsValid verifies every log status variant is recognized and
// unknown values are rejected.
func TestLogStatus_IsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    LogStatus
		expected bool
	}{
		{"running is valid", LogStatusRunning, true},
		{"success is valid", LogStatusSuccess, true},
		{"failed is valid", LogStatusFailed, true},
		{"cancelled is valid", LogStatusCancelled, true},
		{"timeout is valid", LogStatusTimeout, true},
		{"skipped_not_primary is valid", LogStatusSkippedNotPrimary, true},
		{"skipped_singleton is valid", LogStatusSkippedSingleton, true},
		{"skipped_max_concurrency is valid", LogStatusSkippedMaxConcurrency, true},
		{"empty is invalid", LogStatus(""), false},
		{"unknown is invalid", LogStatus("pending"), false},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.value.IsValid(); got != tc.expected {
				t.Fatalf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}

// TestHandlerSource_IsValid verifies the HandlerSource validity check.
func TestHandlerSource_IsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    HandlerSource
		expected bool
	}{
		{"host is valid", HandlerSourceHost, true},
		{"plugin is valid", HandlerSourcePlugin, true},
		{"empty is invalid", HandlerSource(""), false},
		{"unknown is invalid", HandlerSource("wasm"), false},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.value.IsValid(); got != tc.expected {
				t.Fatalf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}

// TestRetentionMode_IsValid verifies the RetentionMode validity check.
func TestRetentionMode_IsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    RetentionMode
		expected bool
	}{
		{"days is valid", RetentionModeDays, true},
		{"count is valid", RetentionModeCount, true},
		{"none is valid", RetentionModeNone, true},
		{"empty is invalid", RetentionMode(""), false},
		{"unknown is invalid", RetentionMode("forever"), false},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.value.IsValid(); got != tc.expected {
				t.Fatalf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}

// TestHandlerI18nKey verifies stable i18n key construction from handler
// reference and field name anchors.
func TestHandlerI18nKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		ref      string
		field    string
		expected string
	}{
		{
			"builds handler display key",
			"clean-stale-sessions", "name",
			"job.handler.clean-stale-sessions.name",
		},
		{
			"handles camelCase ref",
			"CleanStaleSessions", "DisplayName",
			"job.handler.cleanstalesessions.displayname",
		},
		{
			"empty ref returns empty",
			"", "name",
			"",
		},
		{
			"empty field returns empty",
			"my-handler", "",
			"",
		},
		{
			"whitespace-only ref returns empty",
			"   ", "name",
			"",
		},
		{
			"special characters become dots",
			"my_handler", "display name",
			"job.handler.my.handler.display.name",
		},
		{
			"multiple special chars collapse to single dot",
			"a@@b", "c##d",
			"job.handler.a.b.c.d",
		},
		{
			"leading and trailing dots trimmed",
			"!test!", "!field!",
			"job.handler.test.field",
		},
		{
			"preserves hyphens and digits",
			"job-001", "label-v2",
			"job.handler.job-001.label-v2",
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := HandlerI18nKey(tc.ref, tc.field)
			if got != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}

// TestParseRetentionOption verifies retention policy JSON parsing and
// validation across valid inputs, edge cases, and error conditions.
func TestParseRetentionOption(t *testing.T) {
	t.Parallel()

	t.Run("empty string returns nil", func(t *testing.T) {
		t.Parallel()
		opt, err := ParseRetentionOption("")
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if opt != nil {
			t.Fatalf("expected nil option, got %+v", opt)
		}
	})

	t.Run("whitespace-only returns nil", func(t *testing.T) {
		t.Parallel()
		opt, err := ParseRetentionOption("   ")
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if opt != nil {
			t.Fatalf("expected nil option, got %+v", opt)
		}
	})

	t.Run("valid days policy", func(t *testing.T) {
		t.Parallel()
		opt, err := ParseRetentionOption(`{"mode":"days","value":30}`)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if opt.Mode != RetentionModeDays {
			t.Fatalf("expected mode days, got %q", opt.Mode)
		}
		if opt.Value != 30 {
			t.Fatalf("expected value 30, got %d", opt.Value)
		}
	})

	t.Run("valid count policy", func(t *testing.T) {
		t.Parallel()
		opt, err := ParseRetentionOption(`{"mode":"count","value":100}`)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if opt.Mode != RetentionModeCount {
			t.Fatalf("expected mode count, got %q", opt.Mode)
		}
		if opt.Value != 100 {
			t.Fatalf("expected value 100, got %d", opt.Value)
		}
	})

	t.Run("none policy zeroes value", func(t *testing.T) {
		t.Parallel()
		opt, err := ParseRetentionOption(`{"mode":"none","value":5}`)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if opt.Mode != RetentionModeNone {
			t.Fatalf("expected mode none, got %q", opt.Mode)
		}
		if opt.Value != 0 {
			t.Fatalf("expected value 0 (zeroed for none mode), got %d", opt.Value)
		}
	})

	t.Run("mode with whitespace is trimmed", func(t *testing.T) {
		t.Parallel()
		opt, err := ParseRetentionOption(`{"mode":"  days  ","value":30}`)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if opt.Mode != RetentionModeDays {
			t.Fatalf("expected mode days, got %q", opt.Mode)
		}
	})

	t.Run("invalid JSON returns CodeJobRetentionParseFailed", func(t *testing.T) {
		t.Parallel()
		_, err := ParseRetentionOption("not-json")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !bizerr.Is(err, CodeJobRetentionParseFailed) {
			t.Fatalf("expected error to match CodeJobRetentionParseFailed, got %v", err)
		}
	})

	t.Run("unsupported mode returns CodeJobRetentionModeUnsupported", func(t *testing.T) {
		t.Parallel()
		_, err := ParseRetentionOption(`{"mode":"forever","value":1}`)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !bizerr.Is(err, CodeJobRetentionModeUnsupported) {
			t.Fatalf("expected error to match CodeJobRetentionModeUnsupported, got %v", err)
		}
	})

	t.Run("zero value for days returns CodeJobRetentionValueInvalid", func(t *testing.T) {
		t.Parallel()
		_, err := ParseRetentionOption(`{"mode":"days","value":0}`)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !bizerr.Is(err, CodeJobRetentionValueInvalid) {
			t.Fatalf("expected error to match CodeJobRetentionValueInvalid, got %v", err)
		}
	})

	t.Run("negative value returns CodeJobRetentionValueInvalid", func(t *testing.T) {
		t.Parallel()
		_, err := ParseRetentionOption(`{"mode":"count","value":-1}`)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !bizerr.Is(err, CodeJobRetentionValueInvalid) {
			t.Fatalf("expected error to match CodeJobRetentionValueInvalid, got %v", err)
		}
	})
}

// TestMustMarshalRetentionOption verifies retention policy serialization
// including nil-safety and round-trip fidelity with ParseRetentionOption.
func TestMustMarshalRetentionOption(t *testing.T) {
	t.Parallel()

	t.Run("nil option returns empty", func(t *testing.T) {
		t.Parallel()
		raw, err := MustMarshalRetentionOption(nil)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if raw != "" {
			t.Fatalf("expected empty string, got %q", raw)
		}
	})

	t.Run("valid option marshals to JSON", func(t *testing.T) {
		t.Parallel()
		raw, err := MustMarshalRetentionOption(&RetentionOption{
			Mode:  RetentionModeDays,
			Value: 30,
		})
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if raw != `{"mode":"days","value":30}` {
			t.Fatalf("expected compact JSON, got %q", raw)
		}
	})

	t.Run("none mode outputs zero value", func(t *testing.T) {
		t.Parallel()
		raw, err := MustMarshalRetentionOption(&RetentionOption{
			Mode:  RetentionModeNone,
			Value: 0,
		})
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if raw != `{"mode":"none","value":0}` {
			t.Fatalf("expected compact JSON, got %q", raw)
		}
	})

	t.Run("round-trip parse in -> marshal -> parse produces same option", func(t *testing.T) {
		t.Parallel()
		original, err := ParseRetentionOption(`{"mode":"count","value":50}`)
		if err != nil {
			t.Fatalf("parse failed: %v", err)
		}
		raw, err := MustMarshalRetentionOption(original)
		if err != nil {
			t.Fatalf("marshal failed: %v", err)
		}
		roundtripped, err := ParseRetentionOption(raw)
		if err != nil {
			t.Fatalf("re-parse failed: %v", err)
		}
		if roundtripped.Mode != original.Mode || roundtripped.Value != original.Value {
			t.Fatalf("round-trip mismatch: original (%q, %d) vs result (%q, %d)",
				original.Mode, original.Value, roundtripped.Mode, roundtripped.Value)
		}
	})
}

// TestStopReasonValues verifies stop reason constants have the documented values.
func TestStopReasonValues(t *testing.T) {
	t.Parallel()

	if got := string(StopReasonManual); got != "manual" {
		t.Fatalf("expected StopReasonManual=%q, got %q", "manual", got)
	}
	if got := string(StopReasonPluginUnavailable); got != "plugin_unavailable" {
		t.Fatalf("expected StopReasonPluginUnavailable=%q, got %q", "plugin_unavailable", got)
	}
	if got := string(StopReasonMaxExecutionsReached); got != "max_executions_reached" {
		t.Fatalf("expected StopReasonMaxExecutionsReached=%q, got %q", "max_executions_reached", got)
	}
}

// TestCodeDefinitions verifies every jobmeta bizerr.Code is non-nil, carries
// a non-empty runtime code and fallback, and targets a valid GoFrame type code.
func TestCodeDefinitions(t *testing.T) {
	t.Parallel()

	codes := []struct {
		name string
		code *bizerr.Code
	}{
		{"CodeJobNotFound", CodeJobNotFound},
		{"CodeJobTaskTypeUnsupported", CodeJobTaskTypeUnsupported},
		{"CodeJobCronFieldCountUnsupported", CodeJobCronFieldCountUnsupported},
		{"CodeJobHandlerUnavailable", CodeJobHandlerUnavailable},
		{"CodeJobSnapshotMarshalFailed", CodeJobSnapshotMarshalFailed},
		{"CodeJobLogNotRunning", CodeJobLogNotRunning},
		{"CodeJobShellExecutorUninitialized", CodeJobShellExecutorUninitialized},
		{"CodeJobShellDisabled", CodeJobShellDisabled},
		{"CodeJobShellCommandRequired", CodeJobShellCommandRequired},
		{"CodeJobShellTimeoutInvalid", CodeJobShellTimeoutInvalid},
		{"CodeJobShellStartFailed", CodeJobShellStartFailed},
		{"CodeJobShellExecutionFailed", CodeJobShellExecutionFailed},
		{"CodeJobShellWorkdirRootDenied", CodeJobShellWorkdirRootDenied},
		{"CodeJobShellWorkdirValidateFailed", CodeJobShellWorkdirValidateFailed},
		{"CodeJobShellWorkdirNotDirectory", CodeJobShellWorkdirNotDirectory},
		{"CodeJobRetentionParseFailed", CodeJobRetentionParseFailed},
		{"CodeJobRetentionModeUnsupported", CodeJobRetentionModeUnsupported},
		{"CodeJobRetentionValueInvalid", CodeJobRetentionValueInvalid},
		{"CodeJobRetentionMarshalFailed", CodeJobRetentionMarshalFailed},
	}

	for _, tc := range codes {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if tc.code == nil {
				t.Fatal("expected code to be non-nil")
			}
			if tc.code.RuntimeCode() == "" {
				t.Fatalf("expected non-empty runtime code for %s", tc.name)
			}
			if tc.code.Fallback() == "" {
				t.Fatalf("expected non-empty fallback for %s", tc.name)
			}
			if tc.code.TypeCode() == gcode.CodeUnknown {
				t.Fatalf("expected valid GoFrame type code for %s, got CodeUnknown", tc.name)
			}
		})
	}
}
