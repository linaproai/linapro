// This file defines public frontend settings managed by sys_config and the
// safe whitelist payload exposed to login pages and admin workspace startup.

package config

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
)

// Protected public frontend setting keys stored in sys_config.
const (
	// PublicFrontendSettingKeyAppName stores the public-facing application name.
	PublicFrontendSettingKeyAppName = "sys.app.name"
	// PublicFrontendSettingKeyAppLogo stores the default light-theme logo source.
	PublicFrontendSettingKeyAppLogo = "sys.app.logo"
	// PublicFrontendSettingKeyAppLogoDark stores the dark-theme logo source.
	PublicFrontendSettingKeyAppLogoDark = "sys.app.logoDark"
	// PublicFrontendSettingKeyAuthPageTitle stores the login-page headline.
	PublicFrontendSettingKeyAuthPageTitle = "sys.auth.pageTitle"
	// PublicFrontendSettingKeyAuthPageDesc stores the login-page description.
	PublicFrontendSettingKeyAuthPageDesc = "sys.auth.pageDesc"
	// PublicFrontendSettingKeyAuthLoginSubtitle stores the login-form subtitle.
	PublicFrontendSettingKeyAuthLoginSubtitle = "sys.auth.loginSubtitle"
	// PublicFrontendSettingKeyAuthLoginPanelLayout stores the login-form panel layout.
	PublicFrontendSettingKeyAuthLoginPanelLayout = "sys.auth.loginPanelLayout"
	// PublicFrontendSettingKeyUIThemeMode stores the frontend theme mode.
	PublicFrontendSettingKeyUIThemeMode = "sys.ui.theme.mode"
	// PublicFrontendSettingKeyUILayout stores the admin layout mode.
	PublicFrontendSettingKeyUILayout = "sys.ui.layout"
	// PublicFrontendSettingKeyUIWatermarkEnabled stores whether watermark is enabled.
	PublicFrontendSettingKeyUIWatermarkEnabled = "sys.ui.watermark.enabled"
	// PublicFrontendSettingKeyUIWatermarkContent stores the watermark content.
	PublicFrontendSettingKeyUIWatermarkContent = "sys.ui.watermark.content"
)

// PublicFrontendAuthPanelLayout defines the supported login-form panel layouts.
type PublicFrontendAuthPanelLayout string

const (
	// PublicFrontendAuthPanelLayoutLeft aligns the login panel to the left.
	PublicFrontendAuthPanelLayoutLeft PublicFrontendAuthPanelLayout = "panel-left"
	// PublicFrontendAuthPanelLayoutCenter centers the login panel.
	PublicFrontendAuthPanelLayoutCenter PublicFrontendAuthPanelLayout = "panel-center"
	// PublicFrontendAuthPanelLayoutRight aligns the login panel to the right.
	PublicFrontendAuthPanelLayoutRight PublicFrontendAuthPanelLayout = "panel-right"
)

// publicFrontendSettingSpecs lists the built-in public frontend settings that
// can be overridden through protected sys_config entries.
var publicFrontendSettingSpecs = []RuntimeParamSpec{
	{
		Key:          PublicFrontendSettingKeyAppName,
		Name:         "品牌展示-应用名称",
		DefaultValue: "LinaPro",
		Remark:       "控制浏览器标题、登录页品牌名称和工作台 Logo 文案展示，建议填写简洁的产品名称。",
	},
	{
		Key:          PublicFrontendSettingKeyAppLogo,
		Name:         "品牌展示-应用 Logo",
		DefaultValue: "https://unpkg.com/@vbenjs/static-source@0.1.7/source/logo-v1.webp",
		Remark:       "控制登录页与工作台默认 Logo 图片地址，支持 http(s) 或站内绝对路径。",
	},
	{
		Key:          PublicFrontendSettingKeyAppLogoDark,
		Name:         "品牌展示-深色 Logo",
		DefaultValue: "https://unpkg.com/@vbenjs/static-source@0.1.7/source/logo-v1.webp",
		Remark:       "控制深色主题下的 Logo 图片地址，支持 http(s) 或站内绝对路径。",
	},
	{
		Key:          PublicFrontendSettingKeyAuthPageTitle,
		Name:         "登录展示-页面标题",
		DefaultValue: "AI驱动的全栈开发框架",
		Remark:       "控制登录页顶部主标题文案。",
	},
	{
		Key:          PublicFrontendSettingKeyAuthPageDesc,
		Name:         "登录展示-页面说明",
		DefaultValue: "核心宿主服务、默认管理工作台与插件扩展能力",
		Remark:       "控制登录页顶部说明文案。",
	},
	{
		Key:          PublicFrontendSettingKeyAuthLoginSubtitle,
		Name:         "登录展示-登录副标题",
		DefaultValue: "请输入您的帐户信息以进入 LinaPro 宿主工作区",
		Remark:       "控制登录表单上方的提示说明文案。",
	},
	{
		Key:          PublicFrontendSettingKeyAuthLoginPanelLayout,
		Name:         "登录展示-登录框位置",
		DefaultValue: string(PublicFrontendAuthPanelLayoutCenter),
		Remark:       "控制登录框默认布局，可选值：panel-left、panel-center、panel-right。",
	},
	{
		Key:          PublicFrontendSettingKeyUIThemeMode,
		Name:         "界面风格-主题模式",
		DefaultValue: "light",
		Remark:       "控制默认主题模式，可选值：light、dark、auto。",
	},
	{
		Key:          PublicFrontendSettingKeyUILayout,
		Name:         "界面风格-工作台布局",
		DefaultValue: "sidebar-nav",
		Remark:       "控制后台默认布局，可选值：sidebar-nav、sidebar-mixed-nav、header-nav、header-sidebar-nav、header-mixed-nav、mixed-nav、full-content。",
	},
	{
		Key:          PublicFrontendSettingKeyUIWatermarkEnabled,
		Name:         "界面风格-是否启用水印",
		DefaultValue: "false",
		Remark:       "控制工作台是否启用水印，可选值：true、false。",
	},
	{
		Key:          PublicFrontendSettingKeyUIWatermarkContent,
		Name:         "界面风格-水印文案",
		DefaultValue: "LinaPro",
		Remark:       "控制工作台水印文案内容。",
	},
}

// publicFrontendSettingSpecByKey indexes publicFrontendSettingSpecs by key for
// constant-time lookup in validation and projection paths.
var publicFrontendSettingSpecByKey = func() map[string]RuntimeParamSpec {
	specByKey := make(map[string]RuntimeParamSpec, len(publicFrontendSettingSpecs))
	for _, spec := range publicFrontendSettingSpecs {
		specByKey[spec.Key] = spec
	}
	return specByKey
}()

// publicFrontendSettingKeys keeps the deterministic key order for public
// frontend protected-config queries.
var publicFrontendSettingKeys = []string{
	PublicFrontendSettingKeyAppName,
	PublicFrontendSettingKeyAppLogo,
	PublicFrontendSettingKeyAppLogoDark,
	PublicFrontendSettingKeyAuthPageTitle,
	PublicFrontendSettingKeyAuthPageDesc,
	PublicFrontendSettingKeyAuthLoginSubtitle,
	PublicFrontendSettingKeyAuthLoginPanelLayout,
	PublicFrontendSettingKeyUIThemeMode,
	PublicFrontendSettingKeyUILayout,
	PublicFrontendSettingKeyUIWatermarkEnabled,
	PublicFrontendSettingKeyUIWatermarkContent,
}

// protectedConfigKeys contains all built-in config keys whose lifecycle is
// governed by the host and must not be renamed or deleted.
var protectedConfigKeys = appendProtectedConfigKeys()

// PublicFrontendConfig describes the safe frontend settings exposed by the host.
type PublicFrontendConfig struct {
	App  PublicFrontendAppConfig  `json:"app"`  // App groups brand-related settings.
	Auth PublicFrontendAuthConfig `json:"auth"` // Auth groups login-page copy settings.
	UI   PublicFrontendUIConfig   `json:"ui"`   // UI groups theme, layout, and watermark settings.
	Cron PublicFrontendCronConfig `json:"cron"` // Cron groups public-safe scheduled-job capability flags.
}

// PublicFrontendAppConfig stores brand-related public settings.
type PublicFrontendAppConfig struct {
	Name     string `json:"name"`     // Name is the public application name.
	Logo     string `json:"logo"`     // Logo is the default logo source.
	LogoDark string `json:"logoDark"` // LogoDark is the dark-theme logo source.
}

// PublicFrontendAuthConfig stores login-page copy settings.
type PublicFrontendAuthConfig struct {
	PageTitle     string                        `json:"pageTitle"`     // PageTitle is the login-page headline.
	PageDesc      string                        `json:"pageDesc"`      // PageDesc is the login-page description.
	LoginSubtitle string                        `json:"loginSubtitle"` // LoginSubtitle is the form subtitle.
	PanelLayout   PublicFrontendAuthPanelLayout `json:"panelLayout"`   // PanelLayout selects the login-panel placement.
}

// PublicFrontendUIConfig stores safe theme and layout preferences.
type PublicFrontendUIConfig struct {
	ThemeMode        string `json:"themeMode"`        // ThemeMode is one of light, dark, or auto.
	Layout           string `json:"layout"`           // Layout is the default admin layout mode.
	WatermarkEnabled bool   `json:"watermarkEnabled"` // WatermarkEnabled reports whether watermark is enabled.
	WatermarkContent string `json:"watermarkContent"` // WatermarkContent is the watermark text.
}

// PublicFrontendCronConfig stores public-safe scheduled-job runtime settings.
type PublicFrontendCronConfig struct {
	LogRetention PublicFrontendCronLogRetentionConfig `json:"logRetention"` // LogRetention exposes the system-wide job-log cleanup policy to the UI.
	Shell        PublicFrontendCronShellConfig        `json:"shell"`        // Shell exposes whether shell jobs are available to the UI.
	Timezone     PublicFrontendCronTimezoneConfig     `json:"timezone"`     // Timezone exposes the current host timezone to the UI.
}

// PublicFrontendCronLogRetentionConfig stores the frontend-visible default
// job-log retention policy.
type PublicFrontendCronLogRetentionConfig struct {
	Mode  CronLogRetentionMode `json:"mode"`  // Mode selects days, count, or none.
	Value int64                `json:"value"` // Value stores the current system threshold.
}

// PublicFrontendCronShellConfig stores the frontend-visible shell-job gate.
type PublicFrontendCronShellConfig struct {
	Enabled        bool   `json:"enabled"`                  // Enabled reports whether shell jobs are currently allowed.
	Supported      bool   `json:"supported"`                // Supported reports whether the current platform supports shell jobs.
	DisabledReason string `json:"disabledReason,omitempty"` // DisabledReason explains why shell jobs are unavailable.
}

// PublicFrontendCronTimezoneConfig stores the frontend-visible default timezone.
type PublicFrontendCronTimezoneConfig struct {
	Current string `json:"current"` // Current is the current host timezone identifier.
}

// PublicFrontendSettingSpecs returns all built-in public frontend setting specs.
func PublicFrontendSettingSpecs() []RuntimeParamSpec {
	specs := make([]RuntimeParamSpec, len(publicFrontendSettingSpecs))
	copy(specs, publicFrontendSettingSpecs)
	return specs
}

// LookupPublicFrontendSettingSpec returns one built-in public frontend setting spec by key.
func LookupPublicFrontendSettingSpec(key string) (RuntimeParamSpec, bool) {
	spec, ok := publicFrontendSettingSpecByKey[strings.TrimSpace(key)]
	return spec, ok
}

// IsProtectedConfigParam reports whether the key belongs to one built-in host
// parameter whose key name and record lifecycle are protected.
func IsProtectedConfigParam(key string) bool {
	if IsProtectedRuntimeParam(key) {
		return true
	}
	_, ok := LookupPublicFrontendSettingSpec(key)
	return ok
}

// ValidateProtectedConfigValue validates one built-in protected config value.
func ValidateProtectedConfigValue(key string, value string) error {
	trimmedKey := strings.TrimSpace(key)
	if IsProtectedRuntimeParam(trimmedKey) {
		return ValidateRuntimeParamValue(trimmedKey, value)
	}
	return ValidatePublicFrontendSettingValue(trimmedKey, value)
}

// normalizeProtectedConfigValue trims whitespace and lowercases one protected
// config value before enum-style comparisons.
func normalizeProtectedConfigValue(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

// ValidatePublicFrontendSettingValue validates one built-in public frontend setting value.
func ValidatePublicFrontendSettingValue(key string, value string) error {
	switch strings.TrimSpace(key) {
	case PublicFrontendSettingKeyAppName,
		PublicFrontendSettingKeyAuthPageTitle,
		PublicFrontendSettingKeyAuthPageDesc,
		PublicFrontendSettingKeyAuthLoginSubtitle,
		PublicFrontendSettingKeyUIWatermarkContent:
		return validateRequiredTextValue(key, value, 120)

	case PublicFrontendSettingKeyAuthLoginPanelLayout:
		return validateAllowedStringValue(key, value, []string{
			string(PublicFrontendAuthPanelLayoutLeft),
			string(PublicFrontendAuthPanelLayoutCenter),
			string(PublicFrontendAuthPanelLayoutRight),
		})

	case PublicFrontendSettingKeyAppLogo, PublicFrontendSettingKeyAppLogoDark:
		return validateRequiredTextValue(key, value, 500)

	case PublicFrontendSettingKeyUIThemeMode:
		return validateAllowedStringValue(key, value, []string{"light", "dark", "auto"})

	case PublicFrontendSettingKeyUILayout:
		return validateAllowedStringValue(key, value, []string{
			"sidebar-nav",
			"sidebar-mixed-nav",
			"header-nav",
			"header-sidebar-nav",
			"header-mixed-nav",
			"mixed-nav",
			"full-content",
		})

	case PublicFrontendSettingKeyUIWatermarkEnabled:
		_, err := parseStrictBoolValue(key, value)
		return err
	}
	return nil
}

// GetPublicFrontend returns the public-safe frontend branding and display
// configuration consumed by login pages and admin workspace startup.
func (s *serviceImpl) GetPublicFrontend(ctx context.Context) *PublicFrontendConfig {
	cronCfg := s.GetCron(ctx)
	return &PublicFrontendConfig{
		App: PublicFrontendAppConfig{
			Name:     s.getProtectedConfigValueOrDefault(ctx, PublicFrontendSettingKeyAppName),
			Logo:     s.getProtectedConfigValueOrDefault(ctx, PublicFrontendSettingKeyAppLogo),
			LogoDark: s.getProtectedConfigValueOrDefault(ctx, PublicFrontendSettingKeyAppLogoDark),
		},
		Auth: PublicFrontendAuthConfig{
			PageTitle:     s.getProtectedConfigValueOrDefault(ctx, PublicFrontendSettingKeyAuthPageTitle),
			PageDesc:      s.getProtectedConfigValueOrDefault(ctx, PublicFrontendSettingKeyAuthPageDesc),
			LoginSubtitle: s.getProtectedConfigValueOrDefault(ctx, PublicFrontendSettingKeyAuthLoginSubtitle),
			PanelLayout: PublicFrontendAuthPanelLayout(
				s.getProtectedConfigValueOrDefault(ctx, PublicFrontendSettingKeyAuthLoginPanelLayout),
			),
		},
		UI: PublicFrontendUIConfig{
			ThemeMode:        s.getProtectedConfigValueOrDefault(ctx, PublicFrontendSettingKeyUIThemeMode),
			Layout:           s.getProtectedConfigValueOrDefault(ctx, PublicFrontendSettingKeyUILayout),
			WatermarkEnabled: s.getProtectedConfigBoolOrDefault(ctx, PublicFrontendSettingKeyUIWatermarkEnabled),
			WatermarkContent: s.getProtectedConfigValueOrDefault(ctx, PublicFrontendSettingKeyUIWatermarkContent),
		},
		Cron: PublicFrontendCronConfig{
			LogRetention: PublicFrontendCronLogRetentionConfig{
				Mode:  cronCfg.LogRetention.Mode,
				Value: cronCfg.LogRetention.Value,
			},
			Shell: PublicFrontendCronShellConfig{
				Enabled:        cronCfg.Shell.Enabled,
				Supported:      cronCfg.Shell.Supported,
				DisabledReason: cronCfg.Shell.DisabledReason,
			},
			Timezone: PublicFrontendCronTimezoneConfig{
				Current: resolveCurrentSystemTimezone(),
			},
		},
	}
}

// resolveCurrentSystemTimezone returns the host timezone identifier exposed to the frontend.
func resolveCurrentSystemTimezone() string {
	if timezone := strings.TrimSpace(os.Getenv("TZ")); timezone != "" && timezone != "Local" {
		if _, err := time.LoadLocation(timezone); err == nil {
			return timezone
		}
	}
	if timezone := strings.TrimSpace(time.Now().Location().String()); timezone != "" && timezone != "Local" {
		if _, err := time.LoadLocation(timezone); err == nil {
			return timezone
		}
	}
	return "Asia/Shanghai"
}

// appendProtectedConfigKeys returns the full protected-config key list by
// combining runtime backend settings and public frontend settings.
func appendProtectedConfigKeys() []string {
	keys := make([]string, 0, len(runtimeParamKeys)+len(publicFrontendSettingKeys))
	keys = append(keys, runtimeParamKeys...)
	keys = append(keys, publicFrontendSettingKeys...)
	return keys
}

// getProtectedConfigValueOrDefault returns the runtime override when present or
// falls back to the built-in default from the protected setting spec.
func (s *serviceImpl) getProtectedConfigValueOrDefault(ctx context.Context, key string) string {
	if value, ok := s.lookupRuntimeParamValue(ctx, key); ok {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	spec, ok := LookupPublicFrontendSettingSpec(key)
	if ok {
		return spec.DefaultValue
	}
	specRuntime, ok := LookupRuntimeParamSpec(key)
	if ok {
		return specRuntime.DefaultValue
	}
	return ""
}

// getProtectedConfigBoolOrDefault returns one protected boolean setting using
// the default-aware string lookup path first.
func (s *serviceImpl) getProtectedConfigBoolOrDefault(ctx context.Context, key string) bool {
	value := s.getProtectedConfigValueOrDefault(ctx, key)
	parsed, err := parseStrictBoolValue(key, value)
	if err != nil {
		panic(err)
	}
	return parsed
}

// parseStrictBoolValue parses one protected boolean setting accepting only
// explicit true or false literals.
func parseStrictBoolValue(key string, value string) (bool, error) {
	switch normalizeProtectedConfigValue(value) {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, gerror.Newf("参数 %s 必须为 true 或 false", key)
	}
}

// validateAllowedStringValue validates one protected string against a fixed
// whitelist of allowed values.
func validateAllowedStringValue(key string, value string, allowed []string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return gerror.Newf("参数 %s 不能为空", key)
	}
	for _, item := range allowed {
		if trimmed == item {
			return nil
		}
	}
	return gerror.Newf("参数 %s 不在支持范围内", key)
}

// validateRequiredTextValue validates one non-empty protected text value with
// a maximum length constraint.
func validateRequiredTextValue(key string, value string, maxLen int) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return gerror.Newf("参数 %s 不能为空", key)
	}
	if len(trimmed) > maxLen {
		return gerror.Newf("参数 %s 长度不能超过 %d 个字符", key, maxLen)
	}
	return nil
}
