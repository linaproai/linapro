// This file defines scheduled-job runtime configuration backed by protected
// sys_config entries.

package config

import (
	"context"
	"encoding/json"
	"io"
	"runtime"
	"strings"

	"lina-core/pkg/bizerr"
)

// Cron shell disabled-reason constants define the runtime i18n key and English
// source fallback returned when the current platform does not allow shell jobs.
const (
	cronShellUnsupportedReasonKey = "config.cron.shell.unsupportedReason"
	cronShellUnsupportedReason    = "Current platform does not support shell mode"
)

// CronLogRetentionMode defines the supported cron-log cleanup strategies.
type CronLogRetentionMode string

// Supported cron-log cleanup strategies.
const (
	// CronLogRetentionModeDays removes logs older than the configured day count.
	CronLogRetentionModeDays CronLogRetentionMode = "days"
	// CronLogRetentionModeCount keeps only the configured number of latest logs.
	CronLogRetentionModeCount CronLogRetentionMode = "count"
	// CronLogRetentionModeNone disables automatic cleanup for matching logs.
	CronLogRetentionModeNone CronLogRetentionMode = "none"
)

// CronConfig describes runtime settings used by scheduled-job management.
type CronConfig struct {
	Shell        CronShellConfig        `json:"shell"`        // Shell groups shell-execution gates.
	LogRetention CronLogRetentionConfig `json:"logRetention"` // LogRetention stores the default cleanup policy.
}

// CronShellConfig describes the shell-job execution gate visible to runtime
// services and the frontend.
type CronShellConfig struct {
	Enabled           bool   `json:"enabled"`                     // Enabled reports whether shell jobs are currently allowed.
	Supported         bool   `json:"supported"`                   // Supported reports whether the current platform supports shell jobs.
	DisabledReason    string `json:"disabledReason,omitempty"`    // DisabledReason explains why shell jobs are unavailable.
	DisabledReasonKey string `json:"disabledReasonKey,omitempty"` // DisabledReasonKey stores the runtime i18n key for DisabledReason.
}

// CronLogRetentionConfig stores one normalized cron-log cleanup policy.
type CronLogRetentionConfig struct {
	Mode  CronLogRetentionMode `json:"mode"`  // Mode selects days, count, or none.
	Value int64                `json:"value"` // Value stores the positive threshold for the selected mode.
}

// cronLogRetentionPayload mirrors the raw JSON stored in sys_config before
// normalization and validation.
type cronLogRetentionPayload struct {
	Mode  CronLogRetentionMode `json:"mode"`
	Value int64                `json:"value"`
}

// GetCron reads runtime cron-management parameters from protected sys_config entries.
func (s *serviceImpl) GetCron(ctx context.Context) (*CronConfig, error) {
	shellEnabled, err := s.getProtectedConfigBoolOrDefault(ctx, RuntimeParamKeyCronShellEnabled)
	if err != nil {
		return nil, err
	}
	logRetention, err := s.GetCronLogRetention(ctx)
	if err != nil {
		return nil, err
	}
	return &CronConfig{
		Shell:        buildCronShellConfig(shellEnabled, runtime.GOOS),
		LogRetention: *logRetention,
	}, nil
}

// IsCronShellEnabled reports whether shell-type cron jobs are currently allowed.
func (s *serviceImpl) IsCronShellEnabled(ctx context.Context) (bool, error) {
	shellEnabled, err := s.getProtectedConfigBoolOrDefault(ctx, RuntimeParamKeyCronShellEnabled)
	if err != nil {
		return false, err
	}
	return buildCronShellConfig(shellEnabled, runtime.GOOS).Enabled, nil
}

// GetCronLogRetention returns the runtime-effective default cron log retention policy.
func (s *serviceImpl) GetCronLogRetention(ctx context.Context) (*CronLogRetentionConfig, error) {
	value := s.getProtectedConfigValueOrDefault(ctx, RuntimeParamKeyCronLogRetention)
	cfg, err := parseCronLogRetentionValue(RuntimeParamKeyCronLogRetention, value)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

// buildCronShellConfig normalizes one shell gate against the current platform
// support boundary.
func buildCronShellConfig(enabled bool, goos string) CronShellConfig {
	cfg := CronShellConfig{
		Enabled:   enabled,
		Supported: true,
	}
	if strings.EqualFold(strings.TrimSpace(goos), "windows") {
		cfg.Enabled = false
		cfg.Supported = false
		cfg.DisabledReason = cronShellUnsupportedReason
		cfg.DisabledReasonKey = cronShellUnsupportedReasonKey
	}
	return cfg
}

// validateCronLogRetentionValue validates the stored JSON payload for the
// default cron-log retention policy.
func validateCronLogRetentionValue(key string, value string) error {
	_, err := parseCronLogRetentionValue(key, value)
	return err
}

// parseCronLogRetentionValue parses and normalizes one cron-log retention JSON
// payload from protected sys_config.
func parseCronLogRetentionValue(key string, value string) (*CronLogRetentionConfig, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, bizerr.NewCode(CodeConfigParamRequired, bizerr.P("key", key))
	}

	var payload cronLogRetentionPayload
	decoder := json.NewDecoder(strings.NewReader(trimmed))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		return nil, bizerr.WrapCode(err, CodeConfigParamJSONObjectInvalid, bizerr.P("key", key))
	}
	if err := ensureJSONDecoderEOF(decoder, key); err != nil {
		return nil, err
	}

	mode := CronLogRetentionMode(strings.TrimSpace(string(payload.Mode)))
	switch mode {
	case CronLogRetentionModeDays, CronLogRetentionModeCount:
		if payload.Value <= 0 {
			return nil, bizerr.NewCode(CodeConfigCronRetentionValuePositiveRequired, bizerr.P("key", key))
		}
		return &CronLogRetentionConfig{
			Mode:  mode,
			Value: payload.Value,
		}, nil

	case CronLogRetentionModeNone:
		if payload.Value < 0 {
			return nil, bizerr.NewCode(CodeConfigCronRetentionValueNonNegativeRequired, bizerr.P("key", key))
		}
		return &CronLogRetentionConfig{
			Mode:  CronLogRetentionModeNone,
			Value: 0,
		}, nil
	}

	return nil, bizerr.NewCode(CodeConfigCronRetentionModeUnsupported, bizerr.P("key", key))
}

// ensureJSONDecoderEOF verifies the JSON decoder has no trailing non-space
// content after the first decoded value.
func ensureJSONDecoderEOF(decoder *json.Decoder, key string) error {
	var trailing struct{}
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return bizerr.NewCode(CodeConfigParamJSONSingleObjectRequired, bizerr.P("key", key))
		}
		return bizerr.WrapCode(err, CodeConfigParamJSONTrailingContent, bizerr.P("key", key))
	}
	return nil
}
