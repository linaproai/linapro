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
	// PublicFrontendSettingKeyAuthSloganImage stores the optional login-page
	// side slogan illustration image source.
	PublicFrontendSettingKeyAuthSloganImage = "sys.auth.sloganImage"
	// PublicFrontendSettingKeyAuthForgetPasswordEnabled stores whether the
	// login-page forget-password entry is exposed.
	PublicFrontendSettingKeyAuthForgetPasswordEnabled = "sys.auth.forgetPasswordEnabled"
	// PublicFrontendSettingKeyAuthRegisterEnabled stores whether the
	// login-page create-account entry is exposed.
	PublicFrontendSettingKeyAuthRegisterEnabled = "sys.auth.registerEnabled"
	// PublicFrontendSettingKeyAuthPrivacyPolicy stores the privacy-policy body
	// shown on the public registration page.
	PublicFrontendSettingKeyAuthPrivacyPolicy = "sys.auth.privacyPolicy"
	// PublicFrontendSettingKeyAuthTermsOfService stores the terms-of-service body
	// shown on the public registration page.
	PublicFrontendSettingKeyAuthTermsOfService = "sys.auth.termsOfService"
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
		validator:    validateRequiredTextConfigValue(120),
	},
	{
		Key:          PublicFrontendSettingKeyAppLogo,
		DefaultValue: "/logo.webp",
		validator:    validateRequiredTextConfigValue(500),
	},
	{
		Key:          PublicFrontendSettingKeyAppLogoDark,
		DefaultValue: "/logo.webp",
		validator:    validateRequiredTextConfigValue(500),
	},
	{
		Key:          PublicFrontendSettingKeyUserDefaultAvatar,
		DefaultValue: "/avatar.webp",
		validator:    validateRequiredTextConfigValue(500),
	},
	{
		Key:          PublicFrontendSettingKeyAuthPageTitle,
		DefaultValue: "An AI-native full-stack framework engineered for sustainable delivery",
		validator:    validateRequiredTextConfigValue(120),
	},
	{
		Key:          PublicFrontendSettingKeyAuthPageDesc,
		DefaultValue: "Built for evolving business needs, with an out-of-the-box admin entry point and a flexible pluggable extension model",
		validator:    validateRequiredTextConfigValue(500),
	},
	{
		Key:          PublicFrontendSettingKeyAuthLoginSubtitle,
		DefaultValue: "Enter your account credentials to start managing your projects",
		validator:    validateRequiredTextConfigValue(120),
	},
	{
		Key:          PublicFrontendSettingKeyAuthLoginPanelLayout,
		DefaultValue: string(PublicFrontendAuthPanelLayoutCenter),
		validator: validateAllowedStringConfigValue(
			string(PublicFrontendAuthPanelLayoutLeft),
			string(PublicFrontendAuthPanelLayoutCenter),
			string(PublicFrontendAuthPanelLayoutRight),
		),
	},
	{
		Key:          PublicFrontendSettingKeyAuthSloganImage,
		DefaultValue: "/slogan.svg",
		validator:    validateOptionalTextConfigValue(500),
	},
	{
		Key:          PublicFrontendSettingKeyAuthForgetPasswordEnabled,
		DefaultValue: "true",
		validator:    validateStrictBoolConfigValue,
	},
	{
		Key:          PublicFrontendSettingKeyAuthRegisterEnabled,
		DefaultValue: "true",
		validator:    validateStrictBoolConfigValue,
	},
	{
		Key:          PublicFrontendSettingKeyAuthPrivacyPolicy,
		DefaultValue: "Privacy Policy\n\nThis service collects account information required for authentication and workspace access, including username and email. Data is used only to provide the host workspace and related security features. Contact your administrator for data retention and export requests.",
		validator:    validateRequiredTextConfigValue(20000),
	},
	{
		Key:          PublicFrontendSettingKeyAuthTermsOfService,
		DefaultValue: "Terms of Service\n\nBy creating an account you agree to use this workspace in accordance with your organization's policies. Accounts may be suspended for abuse or security risk. The operator may update these terms; continued use after notice constitutes acceptance.",
		validator:    validateRequiredTextConfigValue(20000),
	},
	{
		Key:          PublicFrontendSettingKeyUIThemeMode,
		DefaultValue: "light",
		validator:    validateAllowedStringConfigValue("light", "dark", "auto"),
	},
	{
		Key:          PublicFrontendSettingKeyUILayout,
		DefaultValue: "sidebar-nav",
		validator: validateAllowedStringConfigValue(
			"sidebar-nav",
			"sidebar-mixed-nav",
			"header-nav",
			"header-sidebar-nav",
			"header-mixed-nav",
			"mixed-nav",
			"full-content",
		),
	},
	{
		Key:          PublicFrontendSettingKeyUIWatermarkEnabled,
		DefaultValue: "false",
		validator:    validateStrictBoolConfigValue,
	},
	{
		Key:          PublicFrontendSettingKeyUIWatermarkContent,
		DefaultValue: "LinaPro",
		validator:    validateRequiredTextConfigValue(120),
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

// PublicFrontendConfig describes the safe frontend settings exposed by the host.
type PublicFrontendConfig struct {
	App       PublicFrontendAppConfig       `json:"app"`       // App groups brand-related settings.
	Auth      PublicFrontendAuthConfig      `json:"auth"`      // Auth groups login-page copy settings.
	User      PublicFrontendUserConfig      `json:"user"`      // User groups user-facing fallback settings.
	UI        PublicFrontendUIConfig        `json:"ui"`        // UI groups theme, layout, and watermark settings.
	Cron      PublicFrontendCronConfig      `json:"cron"`      // Cron groups public-safe scheduled-job capability flags.
	Workspace PublicFrontendWorkspaceConfig `json:"workspace"` // Workspace exposes startup-scoped admin workspace settings.
}

// PublicFrontendAppConfig stores brand-related public settings.
type PublicFrontendAppConfig struct {
	Name     string `json:"name"`     // Name is the public application name.
	Logo     string `json:"logo"`     // Logo is the default logo source.
	LogoDark string `json:"logoDark"` // LogoDark is the dark-theme logo source.
}

// PublicFrontendAuthConfig stores login-page copy and entry-switch settings.
type PublicFrontendAuthConfig struct {
	PageTitle             string                        `json:"pageTitle"`             // PageTitle is the login-page headline.
	PageDesc              string                        `json:"pageDesc"`              // PageDesc is the login-page description.
	LoginSubtitle         string                        `json:"loginSubtitle"`         // LoginSubtitle is the form subtitle.
	PanelLayout           PublicFrontendAuthPanelLayout `json:"panelLayout"`           // PanelLayout selects the login-panel placement.
	SloganImage           string                        `json:"sloganImage"`           // SloganImage is the optional login side-slogan illustration URL.
	ForgetPasswordEnabled bool                          `json:"forgetPasswordEnabled"` // ForgetPasswordEnabled reports whether the forget-password entry is exposed.
	RegisterEnabled       bool                          `json:"registerEnabled"`       // RegisterEnabled reports whether the create-account entry is exposed.
	PrivacyPolicy         string                        `json:"privacyPolicy"`         // PrivacyPolicy is the privacy-policy body for registration consent.
	TermsOfService        string                        `json:"termsOfService"`        // TermsOfService is the terms body for registration consent.
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

// PublicFrontendWorkspaceConfig stores public-safe admin workspace routing settings.
type PublicFrontendWorkspaceConfig struct {
	BasePath string `json:"basePath"` // BasePath is the admin workspace entry path.
}

// LookupPublicFrontendSettingSpec returns one built-in public frontend setting spec by key.
func LookupPublicFrontendSettingSpec(key string) (RuntimeParamSpec, bool) {
	spec, ok := publicFrontendSettingSpecByKey[strings.TrimSpace(key)]
	return spec, ok
}

// IsManagedSysConfigKey reports whether the key belongs to one built-in
// sys_config value managed by the host runtime.
func IsManagedSysConfigKey(key string) bool {
	if isManagedRuntimeParamKey(key) {
		return true
	}
	_, ok := LookupPublicFrontendSettingSpec(key)
	return ok
}

// ValidateProtectedConfigValue validates one built-in protected config value.
func ValidateProtectedConfigValue(key string, value string) error {
	trimmedKey := strings.TrimSpace(key)
	if isManagedRuntimeParamKey(trimmedKey) {
		return validateRuntimeParamValue(trimmedKey, value)
	}
	return validatePublicFrontendSettingValue(trimmedKey, value)
}

// normalizeProtectedConfigValue trims whitespace and lowercases one protected
// config value before enum-style comparisons.
func normalizeProtectedConfigValue(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

// validatePublicFrontendSettingValue validates one built-in public frontend setting value.
func validatePublicFrontendSettingValue(key string, value string) error {
	spec, ok := LookupPublicFrontendSettingSpec(key)
	if !ok {
		return nil
	}
	return validateConfigSpecValue(spec, strings.TrimSpace(key), value)
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
	// Empty sloganImage is intentional ("hide illustration") and must not fall
	// back to the built-in default when the sys_config row exists with "".
	authSloganImage, err := s.getProtectedConfigValueAllowEmpty(ctx, PublicFrontendSettingKeyAuthSloganImage)
	if err != nil {
		return nil, err
	}
	authForgetPasswordEnabled, err := s.getProtectedConfigBoolOrDefault(ctx, PublicFrontendSettingKeyAuthForgetPasswordEnabled)
	if err != nil {
		return nil, err
	}
	authRegisterEnabled, err := s.getProtectedConfigBoolOrDefault(ctx, PublicFrontendSettingKeyAuthRegisterEnabled)
	if err != nil {
		return nil, err
	}
	authPrivacyPolicy, err := s.getProtectedConfigValueOrDefault(ctx, PublicFrontendSettingKeyAuthPrivacyPolicy)
	if err != nil {
		return nil, err
	}
	authTermsOfService, err := s.getProtectedConfigValueOrDefault(ctx, PublicFrontendSettingKeyAuthTermsOfService)
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
			PageTitle:             authPageTitle,
			PageDesc:              authPageDesc,
			LoginSubtitle:         authLoginSubtitle,
			PanelLayout:           PublicFrontendAuthPanelLayout(authPanelLayout),
			SloganImage:           authSloganImage,
			ForgetPasswordEnabled: authForgetPasswordEnabled,
			RegisterEnabled:       authRegisterEnabled,
			PrivacyPolicy:         authPrivacyPolicy,
			TermsOfService:        authTermsOfService,
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
		Workspace: PublicFrontendWorkspaceConfig{
			BasePath: s.GetWorkspaceBasePath(ctx),
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

// getProtectedConfigValueOrDefault returns the runtime override when present,
// then the active static config value, then built-in host default metadata.
// Empty runtime or static values are treated as missing so required settings
// fall back to their built-in defaults.
func (s *serviceImpl) getProtectedConfigValueOrDefault(ctx context.Context, key string) (string, error) {
	normalizedKey := strings.TrimSpace(key)
	if value, ok, err := s.lookupRuntimeParamValue(ctx, normalizedKey); err != nil {
		return "", err
	} else if ok {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed, nil
		}
	}

	if value, ok, err := lookupStaticHostConfigValue(ctx, normalizedKey); err != nil {
		return "", err
	} else if ok {
		trimmed := strings.TrimSpace(value.String())
		if trimmed != "" {
			return trimmed, nil
		}
	}

	defaultValue, ok := lookupHostConfigDefaultValue(normalizedKey)
	if ok {
		return strings.TrimSpace(hostConfigDefaultValueString(defaultValue)), nil
	}
	return "", nil
}

// getProtectedConfigValueAllowEmpty returns the runtime override when the key
// exists (including intentional empty strings), then the static config value
// when present, then the built-in default. Use this for optional display assets
// where empty means "disabled" rather than "use default".
func (s *serviceImpl) getProtectedConfigValueAllowEmpty(ctx context.Context, key string) (string, error) {
	normalizedKey := strings.TrimSpace(key)
	if value, ok, err := s.lookupRuntimeParamValue(ctx, normalizedKey); err != nil {
		return "", err
	} else if ok {
		return strings.TrimSpace(value), nil
	}

	if value, ok, err := lookupStaticHostConfigValue(ctx, normalizedKey); err != nil {
		return "", err
	} else if ok {
		return strings.TrimSpace(value.String()), nil
	}

	defaultValue, ok := lookupHostConfigDefaultValue(normalizedKey)
	if ok {
		return strings.TrimSpace(hostConfigDefaultValueString(defaultValue)), nil
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

// validateRequiredTextConfigValue returns metadata validation for required text
// protected settings.
func validateRequiredTextConfigValue(maxLen int) protectedConfigValidator {
	return func(key string, value string) error {
		return validateRequiredTextValue(key, value, maxLen)
	}
}

// validateOptionalTextConfigValue returns metadata validation for optional text
// protected settings. Empty values are accepted; non-empty values enforce maxLen.
func validateOptionalTextConfigValue(maxLen int) protectedConfigValidator {
	return func(key string, value string) error {
		return validateOptionalTextValue(key, value, maxLen)
	}
}

// validateAllowedStringConfigValue returns metadata validation for enum-style
// protected settings.
func validateAllowedStringConfigValue(allowed ...string) protectedConfigValidator {
	return func(key string, value string) error {
		return validateAllowedStringValue(key, value, allowed)
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

// validateOptionalTextValue validates one protected text value that may be
// empty, while still enforcing a maximum character-length constraint when set.
func validateOptionalTextValue(key string, value string, maxLen int) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
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
