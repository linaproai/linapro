// This file exposes the runtime-effective logger TraceID switch backed by
// static config.yaml plus a protected sys_config override.

package config

import "context"

// LoggerTraceIDRuntimeMode defines the supported runtime override modes for
// logger TraceID output.
type LoggerTraceIDRuntimeMode string

// Supported runtime override modes for logger TraceID output.
const (
	// LoggerTraceIDRuntimeModeInherit keeps the runtime-effective value aligned
	// with logger.extensions.traceIDEnabled from config.yaml.
	LoggerTraceIDRuntimeModeInherit LoggerTraceIDRuntimeMode = "inherit"
)

// IsLoggerTraceIDEnabled reports whether the current runtime-effective logger
// output should include TraceID.
func (s *serviceImpl) IsLoggerTraceIDEnabled(ctx context.Context) bool {
	staticEnabled := false
	if cfg := s.getStaticLoggerConfig(ctx); cfg != nil {
		staticEnabled = cfg.Extensions.TraceIDEnabled
	}

	value, ok := s.lookupRuntimeParamValue(ctx, RuntimeParamKeyLoggerTraceIDEnabled)
	if !ok {
		return staticEnabled
	}

	enabled, inherit, err := parseLoggerTraceIDRuntimeValue(
		RuntimeParamKeyLoggerTraceIDEnabled,
		value,
	)
	if err != nil {
		panic(err)
	}
	if inherit {
		return staticEnabled
	}
	return enabled
}

// validateLoggerTraceIDRuntimeValue validates one logger TraceID runtime
// override value.
func validateLoggerTraceIDRuntimeValue(key string, value string) error {
	_, _, err := parseLoggerTraceIDRuntimeValue(key, value)
	return err
}

// parseLoggerTraceIDRuntimeValue parses one logger TraceID runtime override.
// The returned inherit flag indicates the caller should fall back to static
// config.yaml instead of using the boolean value.
func parseLoggerTraceIDRuntimeValue(
	key string,
	value string,
) (enabled bool, inherit bool, err error) {
	switch LoggerTraceIDRuntimeMode(normalizeProtectedConfigValue(value)) {
	case LoggerTraceIDRuntimeModeInherit:
		return false, true, nil
	}

	parsed, err := parseStrictBoolValue(key, value)
	if err != nil {
		return false, false, err
	}
	return parsed, false, nil
}
