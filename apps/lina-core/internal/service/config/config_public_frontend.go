// This file defines public frontend settings managed by sys_config and the
// safe whitelist payload exposed to login pages and admin workspace startup.

package config

import (
	"context"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"lina-core/pkg/bizerr"
)

// Protected public frontend setting keys stored in sys_config.
const (
	// PublicFrontendSettingKeyAppName stores the public-facing application name.
	PublicFrontendSettingKeyAppName = "sys.app.name"
	// PublicFrontendSettingKeyAppLogo stores the default light-theme logo source.
	PublicFrontendSettingKeyAppLogo = "sys.app.logo"
	// PublicFrontendSettingKeyAppLogoDark stores the dark-theme logo source.
	PublicFrontendSettingKeyAppLogoDark = "sys.app.logoDark"
	// PublicFrontendSettingKeyUserDefaultAvatar stores the fallback user avatar source.
	PublicFrontendSettingKeyUserDefaultAvatar = "sys.user.defaultAvatar"
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
		DefaultValue: "LinaPro.AI",
	},
	{
		Key:          PublicFrontendSettingKeyAppLogo,
		DefaultValue: "/logo.webp",
	},
	{
		Key:          PublicFrontendSettingKeyAppLogoDark,
		DefaultValue: "/logo.webp",
	},
	{
		Key:          PublicFrontendSettingKeyUserDefaultAvatar,
		DefaultValue: "/avatar.webp",
	},
	{
		Key:          PublicFrontendSettingKeyAuthPageTitle,
		DefaultValue: "An AI-native full-stack framework engineered for sustainable delivery",
	},
	{
		Key:          PublicFrontendSettingKeyAuthPageDesc,
		DefaultValue: "Built for evolving business needs, with an out-of-the-box admin entry point and a flexible pluggable extension model",
	},
	{
		Key:          PublicFrontendSettingKeyAuthLoginSubtitle,
		DefaultValue: "Enter your account credentials to start managing your projects",
	},
	{
		Key:          PublicFrontendSettingKeyAuthLoginPanelLayout,
		DefaultValue: string(PublicFrontendAuthPanelLayoutRight),
	},
	{
		Key:          PublicFrontendSettingKeyUIThemeMode,
		DefaultValue: "light",
	},
	{
		Key:          PublicFrontendSettingKeyUILayout,
		DefaultValue: "sidebar-nav",
	},
	{
		Key:          PublicFrontendSettingKeyUIWatermarkEnabled,
		DefaultValue: "false",
	},
	{
		Key:          PublicFrontendSettingKeyUIWatermarkContent,
		DefaultValue: "LinaPro",
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
	PublicFrontendSettingKeyUserDefaultAvatar,
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
	User PublicFrontendUserConfig `json:"user"` // User groups user-facing fallback settings.
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

// PublicFrontendUserConfig stores user-facing fallback settings.
type PublicFrontendUserConfig struct {
	DefaultAvatar string `json:"defaultAvatar"` // DefaultAvatar is used when a user has no profile avatar.
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
	Enabled           bool   `json:"enabled"`                     // Enabled reports whether shell jobs are currently allowed.
	Supported         bool   `json:"supported"`                   // Supported reports whether the current platform supports shell jobs.
	DisabledReason    string `json:"disabledReason,omitempty"`    // DisabledReason explains why shell jobs are unavailable.
	DisabledReasonKey string `json:"disabledReasonKey,omitempty"` // DisabledReasonKey stores the runtime i18n key for DisabledReason.
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
		PublicFrontendSettingKeyAuthLoginSubtitle,
		PublicFrontendSettingKeyUIWatermarkContent:
		return validateRequiredTextValue(key, value, 120)

	case PublicFrontendSettingKeyAuthLoginPanelLayout:
		return validateAllowedStringValue(key, value, []string{
			string(PublicFrontendAuthPanelLayoutLeft),
			string(PublicFrontendAuthPanelLayoutCenter),
			string(PublicFrontendAuthPanelLayoutRight),
		})

	case PublicFrontendSettingKeyAuthPageDesc,
		PublicFrontendSettingKeyAppLogo,
		PublicFrontendSettingKeyAppLogoDark,
		PublicFrontendSettingKeyUserDefaultAvatar:
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
func (s *serviceImpl) GetPublicFrontend(ctx context.Context) (*PublicFrontendConfig, error) {
	cronCfg, err := s.GetCron(ctx)
	if err != nil {
		return nil, err
	}
	watermarkEnabled, err := s.getProtectedConfigBoolOrDefault(ctx, PublicFrontendSettingKeyUIWatermarkEnabled)
	if err != nil {
		return nil, err
	}
	appName, err := s.getProtectedConfigValueOrDefault(ctx, PublicFrontendSettingKeyAppName)
	if err != nil {
		return nil, err
	}
	appLogo, err := s.getProtectedConfigValueOrDefault(ctx, PublicFrontendSettingKeyAppLogo)
	if err != nil {
		return nil, err
	}
	appLogoDark, err := s.getProtectedConfigValueOrDefault(ctx, PublicFrontendSettingKeyAppLogoDark)
	if err != nil {
		return nil, err
	}
	authPageTitle, err := s.getProtectedConfigValueOrDefault(ctx, PublicFrontendSettingKeyAuthPageTitle)
	if err != nil {
		return nil, err
	}
	authPageDesc, err := s.getProtectedConfigValueOrDefault(ctx, PublicFrontendSettingKeyAuthPageDesc)
	if err != nil {
		return nil, err
	}
	authLoginSubtitle, err := s.getProtectedConfigValueOrDefault(ctx, PublicFrontendSettingKeyAuthLoginSubtitle)
	if err != nil {
		return nil, err
	}
	authPanelLayout, err := s.getProtectedConfigValueOrDefault(ctx, PublicFrontendSettingKeyAuthLoginPanelLayout)
	if err != nil {
		return nil, err
	}
	userDefaultAvatar, err := s.getProtectedConfigValueOrDefault(ctx, PublicFrontendSettingKeyUserDefaultAvatar)
	if err != nil {
		return nil, err
	}
	uiThemeMode, err := s.getProtectedConfigValueOrDefault(ctx, PublicFrontendSettingKeyUIThemeMode)
	if err != nil {
		return nil, err
	}
	uiLayout, err := s.getProtectedConfigValueOrDefault(ctx, PublicFrontendSettingKeyUILayout)
	if err != nil {
		return nil, err
	}
	uiWatermarkContent, err := s.getProtectedConfigValueOrDefault(ctx, PublicFrontendSettingKeyUIWatermarkContent)
	if err != nil {
		return nil, err
	}

	return &PublicFrontendConfig{
		App: PublicFrontendAppConfig{
			Name:     appName,
			Logo:     appLogo,
			LogoDark: appLogoDark,
		},
		Auth: PublicFrontendAuthConfig{
			PageTitle:     authPageTitle,
			PageDesc:      authPageDesc,
			LoginSubtitle: authLoginSubtitle,
			PanelLayout:   PublicFrontendAuthPanelLayout(authPanelLayout),
		},
		User: PublicFrontendUserConfig{
			DefaultAvatar: userDefaultAvatar,
		},
		UI: PublicFrontendUIConfig{
			ThemeMode:        uiThemeMode,
			Layout:           uiLayout,
			WatermarkEnabled: watermarkEnabled,
			WatermarkContent: uiWatermarkContent,
		},
		Cron: PublicFrontendCronConfig{
			LogRetention: PublicFrontendCronLogRetentionConfig{
				Mode:  cronCfg.LogRetention.Mode,
				Value: cronCfg.LogRetention.Value,
			},
			Shell: PublicFrontendCronShellConfig{
				Enabled:           cronCfg.Shell.Enabled,
				Supported:         cronCfg.Shell.Supported,
				DisabledReason:    cronCfg.Shell.DisabledReason,
				DisabledReasonKey: cronCfg.Shell.DisabledReasonKey,
			},
			Timezone: PublicFrontendCronTimezoneConfig{
				Current: resolveCurrentSystemTimezone(),
			},
		},
	}, nil
}

// resolveCurrentSystemTimezone returns the host timezone identifier exposed to the frontend.
func resolveCurrentSystemTimezone() string {
	return resolveSystemTimezone(os.Getenv("TZ"), time.Now().Location().String())
}

// resolveSystemTimezone selects the first valid system timezone candidate.
func resolveSystemTimezone(envTimezone string, processTimezone string) string {
	if timezone := strings.TrimSpace(envTimezone); timezone != "" && timezone != "Local" {
		if _, err := time.LoadLocation(timezone); err == nil {
			return timezone
		}
	}
	if timezone := strings.TrimSpace(processTimezone); timezone != "" && timezone != "Local" {
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
func (s *serviceImpl) getProtectedConfigValueOrDefault(ctx context.Context, key string) (string, error) {
	if value, ok, err := s.lookupRuntimeParamValue(ctx, key); err != nil {
		return "", err
	} else if ok {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed, nil
		}
	}
	spec, ok := LookupPublicFrontendSettingSpec(key)
	if ok {
		return spec.DefaultValue, nil
	}
	specRuntime, ok := LookupRuntimeParamSpec(key)
	if ok {
		return specRuntime.DefaultValue, nil
	}
	return "", nil
}

// getProtectedConfigBoolOrDefault returns one protected boolean setting using
// the default-aware string lookup path first.
func (s *serviceImpl) getProtectedConfigBoolOrDefault(ctx context.Context, key string) (bool, error) {
	value, err := s.getProtectedConfigValueOrDefault(ctx, key)
	if err != nil {
		return false, err
	}
	parsed, err := parseStrictBoolValue(key, value)
	if err != nil {
		return false, err
	}
	return parsed, nil
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
		return false, bizerr.NewCode(CodeConfigParamBoolInvalid, bizerr.P("key", key))
	}
}

// validateAllowedStringValue validates one protected string against a fixed
// whitelist of allowed values.
func validateAllowedStringValue(key string, value string, allowed []string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return bizerr.NewCode(CodeConfigParamRequired, bizerr.P("key", key))
	}
	for _, item := range allowed {
		if trimmed == item {
			return nil
		}
	}
	return bizerr.NewCode(CodeConfigParamAllowedValueInvalid, bizerr.P("key", key))
}

// validateRequiredTextValue validates one non-empty protected text value with
// a maximum character-length constraint.
func validateRequiredTextValue(key string, value string, maxLen int) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return bizerr.NewCode(CodeConfigParamRequired, bizerr.P("key", key))
	}
	if utf8.RuneCountInString(trimmed) > maxLen {
		return bizerr.NewCode(
			CodeConfigParamTextTooLong,
			bizerr.P("key", key),
			bizerr.P("maxLen", maxLen),
		)
	}
	return nil
}
