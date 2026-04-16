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

func mustScanConfig(ctx context.Context, key string, target any) {
	if target == nil {
		panic(gerror.New("配置扫描目标不能为空"))
	}

	value := g.Cfg().MustGet(ctx, key)
	if value == nil {
		return
	}
	if err := value.Scan(target); err != nil {
		panic(gerror.Wrapf(err, "读取配置 %s 失败", key))
	}
}

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

func mustParsePositiveDuration(key string, raw string) time.Duration {
	duration, err := time.ParseDuration(strings.TrimSpace(raw))
	if err != nil {
		panic(gerror.Wrapf(err, "解析配置 %s 失败", key))
	}
	if duration <= 0 {
		panic(gerror.Newf("配置 %s 必须大于 0", key))
	}
	return duration
}

func mustValidateSecondAlignedDuration(key string, duration time.Duration) time.Duration {
	if duration < time.Second {
		panic(gerror.Newf("配置 %s 必须至少为 1s", key))
	}
	if duration%time.Second != 0 {
		panic(gerror.Newf("配置 %s 必须为整秒时长", key))
	}
	return duration
}
