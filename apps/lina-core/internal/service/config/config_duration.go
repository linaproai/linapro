// This file provides shared parsing helpers for duration-based configuration
// values.

package config

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// mustScanConfig scans one config section into the target object and panics on
// invalid configuration.
func mustScanConfig(ctx context.Context, key string, target any) {
	if target == nil {
		panic(gerror.New("config scan target cannot be nil"))
	}

	value := g.Cfg().MustGet(ctx, key)
	if value == nil {
		return
	}
	if err := value.Scan(target); err != nil {
		panic(gerror.Wrapf(err, "read config %s failed", key))
	}
}

// mustLoadDurationConfig loads one duration config value or returns the given
// default when the key is absent.
func mustLoadDurationConfig(ctx context.Context, key string, defaultValue time.Duration) time.Duration {
	if key == "" {
		return defaultValue
	}

	value := g.Cfg().MustGet(ctx, key)
	if value == nil || value.IsEmpty() {
		return defaultValue
	}

	return mustParsePositiveDuration(key, value.String())
}

// mustParsePositiveDuration parses one positive Go duration string or panics.
func mustParsePositiveDuration(key string, raw string) time.Duration {
	duration, err := time.ParseDuration(strings.TrimSpace(raw))
	if err != nil {
		panic(gerror.Wrapf(err, "parse config %s failed", key))
	}
	if duration <= 0 {
		panic(gerror.Newf("config %s must be greater than 0", key))
	}
	return duration
}

// mustValidateSecondAlignedDuration validates that a duration is at least one
// second and aligns to whole-second boundaries.
func mustValidateSecondAlignedDuration(key string, duration time.Duration) time.Duration {
	if duration < time.Second {
		panic(gerror.Newf("config %s must be at least 1s", key))
	}
	if duration%time.Second != 0 {
		panic(gerror.Newf("config %s must align to whole seconds", key))
	}
	return duration
}
