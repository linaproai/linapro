// This file defines built-in runtime parameters backed by sys_config and their
// validation rules.

package config

import (
	"context"
	"net"
	"strconv"
	"strings"
	"time"

	"lina-core/pkg/bizerr"
)

// Built-in runtime parameter keys stored in sys_config.
const (
	// RuntimeParamKeyJWTExpire stores the runtime JWT token lifetime.
	RuntimeParamKeyJWTExpire = "sys.jwt.expire"
	// RuntimeParamKeySessionTimeout stores the runtime online-session inactivity timeout.
	RuntimeParamKeySessionTimeout = "sys.session.timeout"
	// RuntimeParamKeyUploadMaxSize stores the runtime upload size ceiling in MB.
	RuntimeParamKeyUploadMaxSize = "sys.upload.maxSize"
	// RuntimeParamKeyLoginBlackIPList stores the runtime login IP blacklist.
	RuntimeParamKeyLoginBlackIPList = "sys.login.blackIPList"
	// RuntimeParamKeyCronShellEnabled stores the global shell-job enable switch.
	RuntimeParamKeyCronShellEnabled = "cron.shell.enabled"
	// RuntimeParamKeyCronLogRetention stores the default cron-log retention policy.
	RuntimeParamKeyCronLogRetention = "cron.log.retention"
)

// RuntimeParamSpec describes one built-in runtime parameter managed through
// sys_config.
type RuntimeParamSpec struct {
	Key          string // Key is the sys_config key consumed by host runtime paths.
	DefaultValue string // DefaultValue is the host fallback value.
}

// runtimeParamSpecs lists all built-in runtime parameters backed by sys_config.
var runtimeParamSpecs = []RuntimeParamSpec{
	{
		Key:          RuntimeParamKeyJWTExpire,
		DefaultValue: "24h",
	},
	{
		Key:          RuntimeParamKeySessionTimeout,
		DefaultValue: "24h",
	},
	{
		Key:          RuntimeParamKeyUploadMaxSize,
		DefaultValue: "20",
	},
	{
		Key:          RuntimeParamKeyLoginBlackIPList,
		DefaultValue: "",
	},
	{
		Key:          RuntimeParamKeyCronShellEnabled,
		DefaultValue: "false",
	},
	{
		Key:          RuntimeParamKeyCronLogRetention,
		DefaultValue: `{"mode":"days","value":30}`,
	},
}

// runtimeParamSpecByKey indexes runtimeParamSpecs by key for validation and
// lookup operations on protected runtime settings.
var runtimeParamSpecByKey = func() map[string]RuntimeParamSpec {
	specByKey := make(map[string]RuntimeParamSpec, len(runtimeParamSpecs))
	for _, spec := range runtimeParamSpecs {
		specByKey[spec.Key] = spec
	}
	return specByKey
}()

// runtimeParamKeys preserves the deterministic built-in runtime-parameter key order.
var runtimeParamKeys = []string{
	RuntimeParamKeyJWTExpire,
	RuntimeParamKeySessionTimeout,
	RuntimeParamKeyUploadMaxSize,
	RuntimeParamKeyLoginBlackIPList,
	RuntimeParamKeyCronShellEnabled,
	RuntimeParamKeyCronLogRetention,
}

// RuntimeParamSpecs returns all built-in runtime parameter specs.
func RuntimeParamSpecs() []RuntimeParamSpec {
	specs := make([]RuntimeParamSpec, len(runtimeParamSpecs))
	copy(specs, runtimeParamSpecs)
	return specs
}

// LookupRuntimeParamSpec returns one built-in runtime parameter spec by key.
func LookupRuntimeParamSpec(key string) (RuntimeParamSpec, bool) {
	spec, ok := runtimeParamSpecByKey[strings.TrimSpace(key)]
	return spec, ok
}

// IsProtectedRuntimeParam reports whether the key belongs to one built-in
// runtime parameter that must not be renamed or deleted.
func IsProtectedRuntimeParam(key string) bool {
	_, ok := LookupRuntimeParamSpec(key)
	return ok
}

// ValidateRuntimeParamValue validates one built-in runtime parameter value.
func ValidateRuntimeParamValue(key string, value string) error {
	switch strings.TrimSpace(key) {
	case RuntimeParamKeyJWTExpire:
		_, err := validatePositiveDurationValue(key, value)
		return err

	case RuntimeParamKeySessionTimeout:
		_, err := validatePositiveDurationValue(key, value)
		return err

	case RuntimeParamKeyUploadMaxSize:
		_, err := validatePositiveInt64Value(key, value)
		return err

	case RuntimeParamKeyLoginBlackIPList:
		return validateIPBlacklistValue(key, value)

	case RuntimeParamKeyCronShellEnabled:
		_, err := parseStrictBoolValue(key, value)
		return err

	case RuntimeParamKeyCronLogRetention:
		return validateCronLogRetentionValue(key, value)
	}
	return nil
}

// lookupRuntimeParamValue reads one protected runtime parameter value from the
// current immutable snapshot.
func (s *serviceImpl) lookupRuntimeParamValue(ctx context.Context, key string) (value string, ok bool) {
	snapshot := s.getRuntimeParamSnapshot(ctx)
	if snapshot == nil {
		return "", false
	}
	return snapshot.lookupValue(key)
}

// resolveRuntimeDurationOverride returns one runtime duration override when the
// protected parameter exists, or the current static value when it is absent.
func (s *serviceImpl) resolveRuntimeDurationOverride(
	ctx context.Context,
	key string,
	current time.Duration,
) (time.Duration, error) {
	snapshot := s.getRuntimeParamSnapshot(ctx)
	if snapshot == nil {
		return current, nil
	}
	duration, ok, err := snapshot.lookupDuration(key)
	if err != nil {
		return 0, err
	}
	if !ok {
		return current, nil
	}
	return duration, nil
}

// resolveRuntimeInt64Override returns one runtime integer override when the
// protected parameter exists, or the current static value when it is absent.
func (s *serviceImpl) resolveRuntimeInt64Override(
	ctx context.Context,
	key string,
	current int64,
) (int64, error) {
	snapshot := s.getRuntimeParamSnapshot(ctx)
	if snapshot == nil {
		return current, nil
	}
	parsed, ok, err := snapshot.lookupInt64(key)
	if err != nil {
		return 0, err
	}
	if !ok {
		return current, nil
	}
	return parsed, nil
}

// splitSemicolonValues splits one semicolon-delimited config value into
// trimmed non-empty items.
func splitSemicolonValues(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ";")
	values := make([]string, 0, len(parts))
	for _, item := range parts {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		values = append(values, trimmed)
	}
	return values
}

// validatePositiveDurationValue validates one duration-form runtime parameter.
func validatePositiveDurationValue(key string, value string) (time.Duration, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, bizerr.NewCode(CodeConfigParamRequired, bizerr.P("key", key))
	}
	duration, err := time.ParseDuration(trimmed)
	if err != nil {
		return 0, bizerr.WrapCode(err, CodeConfigParamDurationInvalid, bizerr.P("key", key))
	}
	if duration <= 0 {
		return 0, bizerr.NewCode(CodeConfigParamPositiveRequired, bizerr.P("key", key))
	}
	return duration, nil
}

// validatePositiveInt64Value validates one positive integer runtime parameter.
func validatePositiveInt64Value(key string, value string) (int64, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, bizerr.NewCode(CodeConfigParamRequired, bizerr.P("key", key))
	}
	parsed, err := strconv.ParseInt(trimmed, 10, 64)
	if err != nil {
		return 0, bizerr.WrapCode(err, CodeConfigParamIntegerInvalid, bizerr.P("key", key))
	}
	if parsed <= 0 {
		return 0, bizerr.NewCode(CodeConfigParamPositiveRequired, bizerr.P("key", key))
	}
	return parsed, nil
}

// validateIPBlacklistValue validates one semicolon-delimited IP blacklist made
// of individual IPs or CIDR ranges.
func validateIPBlacklistValue(key string, value string) error {
	for _, item := range splitSemicolonValues(value) {
		if net.ParseIP(item) != nil {
			continue
		}
		if _, _, err := net.ParseCIDR(item); err == nil {
			continue
		}
		return bizerr.NewCode(
			CodeConfigParamIPCIDRInvalid,
			bizerr.P("key", key),
			bizerr.P("value", item),
		)
	}
	return nil
}
