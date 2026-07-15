// This file verifies public frontend settings managed by sys_config and the
// safe frontend payload exposed by the host.

package config

import (
	"context"
	"strings"
	"testing"
	"time"
)

// TestPublicFrontendSettingSpecCopiesAreDetached verifies callers cannot mutate the
// shared public-frontend setting specification slice.
func TestPublicFrontendSettingSpecCopiesAreDetached(t *testing.T) {
	specs := publicFrontendSettingSpecsCopy()
	if len(specs) == 0 {
		t.Fatal("expected public frontend setting specs to be present")
	}

	original := publicFrontendSettingSpecs[0].DefaultValue
	specs[0].DefaultValue = "mutated"
	if publicFrontendSettingSpecs[0].DefaultValue != original {
		t.Fatal("expected publicFrontendSettingSpecsCopy to return a detached copy")
	}
}

// TestPublicFrontendSettingSpecDefaultsExposeUpdatedLoginCopy verifies the host
// exposes the latest login copy and layout defaults through spec lookup.
func TestPublicFrontendSettingSpecDefaultsExposeUpdatedLoginCopy(t *testing.T) {
	descSpec, ok := LookupPublicFrontendSettingSpec(PublicFrontendSettingKeyAuthPageDesc)
	if !ok {
		t.Fatal("expected login page description spec to be present")
	}
	if descSpec.DefaultValue != "Built for evolving business needs, with an out-of-the-box admin entry point and a flexible pluggable extension model" {
		t.Fatalf("unexpected login page description default: %q", descSpec.DefaultValue)
	}

	titleSpec, ok := LookupPublicFrontendSettingSpec(PublicFrontendSettingKeyAuthPageTitle)
	if !ok {
		t.Fatal("expected login page title spec to be present")
	}
	if titleSpec.DefaultValue != "An AI-native full-stack framework engineered for sustainable delivery" {
		t.Fatalf("unexpected login page title default: %q", titleSpec.DefaultValue)
	}

	layoutSpec, ok := LookupPublicFrontendSettingSpec(PublicFrontendSettingKeyAuthLoginPanelLayout)
	if !ok {
		t.Fatal("expected login panel layout spec to be present")
	}
	if layoutSpec.DefaultValue != string(PublicFrontendAuthPanelLayoutCenter) {
		t.Fatalf("unexpected login panel layout default: %q", layoutSpec.DefaultValue)
	}

	sloganSpec, ok := LookupPublicFrontendSettingSpec(PublicFrontendSettingKeyAuthSloganImage)
	if !ok {
		t.Fatal("expected login slogan image spec to be present")
	}
	if sloganSpec.DefaultValue != "/slogan.svg" {
		t.Fatalf("unexpected login slogan image default: %q", sloganSpec.DefaultValue)
	}

	avatarSpec, ok := LookupPublicFrontendSettingSpec(PublicFrontendSettingKeyUserDefaultAvatar)
	if !ok {
		t.Fatal("expected default avatar spec to be present")
	}
	if avatarSpec.DefaultValue != "/avatar.webp" {
		t.Fatalf("unexpected default avatar value: %q", avatarSpec.DefaultValue)
	}

	forgetSpec, ok := LookupPublicFrontendSettingSpec(PublicFrontendSettingKeyAuthForgetPasswordEnabled)
	if !ok {
		t.Fatal("expected forget-password switch spec to be present")
	}
	if forgetSpec.DefaultValue != "true" {
		t.Fatalf("unexpected forget-password switch default: %q", forgetSpec.DefaultValue)
	}

	registerSpec, ok := LookupPublicFrontendSettingSpec(PublicFrontendSettingKeyAuthRegisterEnabled)
	if !ok {
		t.Fatal("expected register switch spec to be present")
	}
	if registerSpec.DefaultValue != "true" {
		t.Fatalf("unexpected register switch default: %q", registerSpec.DefaultValue)
	}
}

// TestIsManagedSysConfigKeyRecognizesRuntimeAndFrontendKeys verifies both
// managed sys_config key families are visible through one helper.
func TestIsManagedSysConfigKeyRecognizesRuntimeAndFrontendKeys(t *testing.T) {
	if !IsManagedSysConfigKey(RuntimeParamKeyJWTExpire) {
		t.Fatal("expected runtime param key to be managed")
	}
	if !IsManagedSysConfigKey(PublicFrontendSettingKeyAppName) {
		t.Fatal("expected public frontend key to be managed")
	}
	if IsManagedSysConfigKey("sys.unknown.key") {
		t.Fatal("expected unknown key not to be managed")
	}
}

// TestValidateProtectedConfigValueRoutesToFamilyValidators verifies the
// unified validator dispatches to runtime and public-frontend rules.
func TestValidateProtectedConfigValueRoutesToFamilyValidators(t *testing.T) {
	if err := ValidateProtectedConfigValue(RuntimeParamKeyJWTExpire, "48h"); err != nil {
		t.Fatalf("expected runtime duration validation success, got %v", err)
	}
	if err := ValidateProtectedConfigValue(RuntimeParamKeyJWTExpire, "bad-duration"); err == nil {
		t.Fatal("expected runtime duration validation error")
	}
	if err := ValidateProtectedConfigValue(PublicFrontendSettingKeyUIThemeMode, "auto"); err != nil {
		t.Fatalf("expected public frontend enum validation success, got %v", err)
	}
	if err := ValidateProtectedConfigValue(PublicFrontendSettingKeyUIThemeMode, "night"); err == nil {
		t.Fatal("expected public frontend enum validation error")
	}
}

// TestPublicFrontendSettingValueValidation verifies protected public frontend
// settings enforce their supported value formats.
func TestPublicFrontendSettingValueValidation(t *testing.T) {
	testCases := []struct {
		key       string
		value     string
		shouldErr bool
	}{
		{key: PublicFrontendSettingKeyAppName, value: "LinaPro"},
		{key: PublicFrontendSettingKeyAppName, value: "", shouldErr: true},
		{key: PublicFrontendSettingKeyUserDefaultAvatar, value: "/avatar.webp"},
		{key: PublicFrontendSettingKeyUserDefaultAvatar, value: "", shouldErr: true},
		{key: PublicFrontendSettingKeyAuthLoginPanelLayout, value: "panel-center"},
		{key: PublicFrontendSettingKeyAuthLoginPanelLayout, value: "panel-bottom", shouldErr: true},
		{key: PublicFrontendSettingKeyAuthSloganImage, value: ""},
		{key: PublicFrontendSettingKeyAuthSloganImage, value: "/slogan.svg"},
		{key: PublicFrontendSettingKeyAuthSloganImage, value: strings.Repeat("a", 501), shouldErr: true},
		{key: PublicFrontendSettingKeyUIThemeMode, value: "dark"},
		{key: PublicFrontendSettingKeyUIThemeMode, value: "night", shouldErr: true},
		{key: PublicFrontendSettingKeyUILayout, value: "header-nav"},
		{key: PublicFrontendSettingKeyUILayout, value: "invalid-layout", shouldErr: true},
		{key: PublicFrontendSettingKeyUIWatermarkEnabled, value: "true"},
		{key: PublicFrontendSettingKeyUIWatermarkEnabled, value: "yes", shouldErr: true},
	}

	for _, testCase := range testCases {
		err := validatePublicFrontendSettingValue(testCase.key, testCase.value)
		if testCase.shouldErr && err == nil {
			t.Fatalf("expected validation error for %s=%q", testCase.key, testCase.value)
		}
		if !testCase.shouldErr && err != nil {
			t.Fatalf("expected validation success for %s=%q, got %v", testCase.key, testCase.value, err)
		}
	}
}

// TestPublicFrontendSettingValueAllowsFiveHundredCharacterLoginDescription
// verifies the login-page description accepts up to 500 characters and rejects
// longer protected text values.
func TestPublicFrontendSettingValueAllowsFiveHundredCharacterLoginDescription(
	t *testing.T,
) {
	validDesc := strings.Repeat("能力", 250)
	if err := validatePublicFrontendSettingValue(PublicFrontendSettingKeyAuthPageDesc, validDesc); err != nil {
		t.Fatalf("expected 500-character login description to pass validation, got %v", err)
	}

	tooLongDesc := validDesc + "扩"
	if err := validatePublicFrontendSettingValue(PublicFrontendSettingKeyAuthPageDesc, tooLongDesc); err == nil {
		t.Fatal("expected login description longer than 500 characters to fail validation")
	}
}

// TestProtectedConfigHelpersPreferOverridesAndFallbackDefaults verifies the
// helper readers trim overrides and fall back to built-in defaults.
func TestProtectedConfigHelpersPreferOverridesAndFallbackDefaults(t *testing.T) {
	withCachedRuntimeParamValue(t, PublicFrontendSettingKeyAppName, " LinaPro Custom ")
	svc := New().(*serviceImpl)

	value, err := svc.getProtectedConfigValueOrDefault(context.Background(), PublicFrontendSettingKeyAppName)
	if err != nil {
		t.Fatalf("get protected override value: %v", err)
	}
	if value != "LinaPro Custom" {
		t.Fatalf("expected trimmed protected override value, got %q", value)
	}
	value, err = svc.getProtectedConfigValueOrDefault(context.Background(), RuntimeParamKeyJWTExpire)
	if err != nil {
		t.Fatalf("get protected default value: %v", err)
	}
	if value != runtimeParamSpecByKey[RuntimeParamKeyJWTExpire].DefaultValue {
		t.Fatalf("expected runtime default value fallback, got %q", value)
	}

	withCachedRuntimeParamValue(t, PublicFrontendSettingKeyUIWatermarkEnabled, "true")
	enabled, err := svc.getProtectedConfigBoolOrDefault(context.Background(), PublicFrontendSettingKeyUIWatermarkEnabled)
	if err != nil {
		t.Fatalf("get protected boolean override: %v", err)
	}
	if !enabled {
		t.Fatal("expected protected boolean override to parse as true")
	}
}

// TestPublicFrontendInvalidBooleanReturnsError verifies malformed boolean
// runtime values are propagated to public frontend config readers.
func TestPublicFrontendInvalidBooleanReturnsError(t *testing.T) {
	withCachedRuntimeParamValue(t, PublicFrontendSettingKeyUIWatermarkEnabled, "yes")

	if _, err := New().GetPublicFrontend(context.Background()); err == nil {
		t.Fatal("expected invalid watermark boolean to return an error")
	}
}

// TestGetPublicFrontendUsesProtectedConfigValues verifies protected public
// frontend settings flow into the public frontend payload.
func TestGetPublicFrontendUsesProtectedConfigValues(t *testing.T) {
	withRuntimeParamValue(t, PublicFrontendSettingKeyAppName, "LinaPro Console")
	withRuntimeParamValue(
		t,
		PublicFrontendSettingKeyAuthPageTitle,
		"统一品牌登录入口",
	)
	withRuntimeParamValue(
		t,
		PublicFrontendSettingKeyAuthPageDesc,
		"面向业务演进的宿主入口，支持灵活扩展与统一治理",
	)
	withRuntimeParamValue(
		t,
		PublicFrontendSettingKeyAuthLoginSubtitle,
		"请使用管理员账号登录宿主工作区",
	)
	withRuntimeParamValue(t, PublicFrontendSettingKeyUserDefaultAvatar, "/avatar.webp")
	withRuntimeParamValue(t, PublicFrontendSettingKeyAuthLoginPanelLayout, "panel-right")
	withRuntimeParamValue(t, PublicFrontendSettingKeyAuthSloganImage, "/custom-slogan.webp")
	withRuntimeParamValue(t, PublicFrontendSettingKeyAuthForgetPasswordEnabled, "false")
	withRuntimeParamValue(t, PublicFrontendSettingKeyAuthRegisterEnabled, "false")
	withRuntimeParamValue(t, PublicFrontendSettingKeyUIThemeMode, "dark")
	withRuntimeParamValue(t, PublicFrontendSettingKeyUILayout, "header-nav")
	withRuntimeParamValue(t, PublicFrontendSettingKeyUIWatermarkEnabled, "true")
	withRuntimeParamValue(t, PublicFrontendSettingKeyUIWatermarkContent, "LinaPro Watermark")
	withRuntimeParamValue(t, RuntimeParamKeyCronLogRetention, `{"mode":"count","value":120}`)

	cfg, err := New().GetPublicFrontend(context.Background())
	if err != nil {
		t.Fatalf("get public frontend config: %v", err)
	}
	if cfg.App.Name != "LinaPro Console" {
		t.Fatalf("expected app name override, got %q", cfg.App.Name)
	}
	if cfg.Auth.PageTitle != "统一品牌登录入口" {
		t.Fatalf("expected auth page title override, got %q", cfg.Auth.PageTitle)
	}
	if cfg.Auth.PageDesc != "面向业务演进的宿主入口，支持灵活扩展与统一治理" {
		t.Fatalf("expected auth page description override, got %q", cfg.Auth.PageDesc)
	}
	if cfg.Auth.LoginSubtitle != "请使用管理员账号登录宿主工作区" {
		t.Fatalf("expected auth login subtitle override, got %q", cfg.Auth.LoginSubtitle)
	}
	if cfg.User.DefaultAvatar != "/avatar.webp" {
		t.Fatalf("expected user default avatar override, got %q", cfg.User.DefaultAvatar)
	}
	if cfg.Auth.PanelLayout != PublicFrontendAuthPanelLayoutRight {
		t.Fatalf("expected auth panel layout override, got %q", cfg.Auth.PanelLayout)
	}
	if cfg.Auth.SloganImage != "/custom-slogan.webp" {
		t.Fatalf("expected auth slogan image override, got %q", cfg.Auth.SloganImage)
	}
	if cfg.Auth.ForgetPasswordEnabled {
		t.Fatal("expected forget-password switch override to false")
	}
	if cfg.Auth.RegisterEnabled {
		t.Fatal("expected register switch override to false")
	}
	if cfg.UI.ThemeMode != "dark" {
		t.Fatalf("expected dark theme mode, got %q", cfg.UI.ThemeMode)
	}
	if cfg.UI.Layout != "header-nav" {
		t.Fatalf("expected header-nav layout, got %q", cfg.UI.Layout)
	}
	if !cfg.UI.WatermarkEnabled {
		t.Fatal("expected watermark enabled override")
	}
	if cfg.UI.WatermarkContent != "LinaPro Watermark" {
		t.Fatalf("expected watermark content override, got %q", cfg.UI.WatermarkContent)
	}
	if cfg.Cron.LogRetention.Mode != CronLogRetentionModeCount || cfg.Cron.LogRetention.Value != 120 {
		t.Fatalf(
			"expected public frontend cron log retention count/120, got mode=%q value=%d",
			cfg.Cron.LogRetention.Mode,
			cfg.Cron.LogRetention.Value,
		)
	}
	if cfg.Cron.Timezone.Current == "" {
		t.Fatal("expected public frontend cron timezone current value to be present")
	}
	if cfg.Workspace.BasePath != defaultWorkspaceBasePath {
		t.Fatalf("expected public frontend workspace base path %q, got %q", defaultWorkspaceBasePath, cfg.Workspace.BasePath)
	}
}

// TestGetPublicFrontendPreservesEmptySloganImage verifies an intentional empty
// slogan image is not replaced by the built-in default illustration path.
func TestGetPublicFrontendPreservesEmptySloganImage(t *testing.T) {
	withCachedRuntimeParamValue(t, PublicFrontendSettingKeyAuthSloganImage, "")

	cfg, err := New().GetPublicFrontend(context.Background())
	if err != nil {
		t.Fatalf("get public frontend config: %v", err)
	}
	if cfg.Auth.SloganImage != "" {
		t.Fatalf("expected empty slogan image to remain empty, got %q", cfg.Auth.SloganImage)
	}
}

// TestGetPublicFrontendDefaultsSloganImageWhenMissing verifies missing slogan
// image configuration falls back to the built-in Vben illustration path.
func TestGetPublicFrontendDefaultsSloganImageWhenMissing(t *testing.T) {
	withCachedRuntimeParamSnapshot(t, &runtimeParamSnapshot{
		values:         map[string]string{},
		durationValues: make(map[string]time.Duration),
		int64Values:    make(map[string]int64),
		parseErrors:    make(map[string]error),
	})

	cfg, err := New().GetPublicFrontend(context.Background())
	if err != nil {
		t.Fatalf("get public frontend config: %v", err)
	}
	if cfg.Auth.SloganImage != "/slogan.svg" {
		t.Fatalf("expected default slogan image /slogan.svg, got %q", cfg.Auth.SloganImage)
	}
}

// TestResolveCurrentSystemTimezoneUsesEnvironment verifies a valid TZ
// environment variable is exposed directly to the frontend.
func TestResolveCurrentSystemTimezoneUsesEnvironment(t *testing.T) {
	if timezone := resolveSystemTimezone("Asia/Tokyo", "UTC"); timezone != "Asia/Tokyo" {
		t.Fatalf("expected timezone from TZ environment, got %q", timezone)
	}
}

// TestResolveCurrentSystemTimezoneFallsBackToProcessLocation verifies the
// process location is used when TZ is invalid.
func TestResolveCurrentSystemTimezoneFallsBackToProcessLocation(t *testing.T) {
	if timezone := resolveSystemTimezone("Invalid/Timezone", "UTC"); timezone != "UTC" {
		t.Fatalf("expected timezone fallback to process location UTC, got %q", timezone)
	}
}

// TestResolveCurrentSystemTimezoneUsesProjectDefault verifies the helper uses
// the project default when both environment and process location are local.
func TestResolveCurrentSystemTimezoneUsesProjectDefault(t *testing.T) {
	if timezone := resolveSystemTimezone("", "Local"); timezone != "Asia/Shanghai" {
		t.Fatalf("expected project default timezone Asia/Shanghai, got %q", timezone)
	}
}
