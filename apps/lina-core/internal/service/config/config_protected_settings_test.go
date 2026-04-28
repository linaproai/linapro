// This file verifies protected configuration helper behavior for runtime and
// public-frontend settings.

package config

import (
	"context"
	"testing"
	"time"
)

// TestRuntimeParamSpecsReturnsCopy verifies callers cannot mutate the shared
// built-in runtime-parameter specification slice.
func TestRuntimeParamSpecsReturnsCopy(t *testing.T) {
	specs := RuntimeParamSpecs()
	if len(specs) == 0 {
		t.Fatal("expected runtime param specs to be present")
	}

	original := runtimeParamSpecs[0].DefaultValue
	specs[0].DefaultValue = "mutated"
	if runtimeParamSpecs[0].DefaultValue != original {
		t.Fatal("expected RuntimeParamSpecs to return a detached copy")
	}
}

// TestPublicFrontendSettingSpecsReturnsCopy verifies callers cannot mutate the
// shared public-frontend setting specification slice.
func TestPublicFrontendSettingSpecsReturnsCopy(t *testing.T) {
	specs := PublicFrontendSettingSpecs()
	if len(specs) == 0 {
		t.Fatal("expected public frontend setting specs to be present")
	}

	original := publicFrontendSettingSpecs[0].DefaultValue
	specs[0].DefaultValue = "mutated"
	if publicFrontendSettingSpecs[0].DefaultValue != original {
		t.Fatal("expected PublicFrontendSettingSpecs to return a detached copy")
	}
}

// TestPublicFrontendSettingSpecsExposeUpdatedLoginDefaults verifies the host
// exposes the latest login copy and layout defaults through spec lookup.
func TestPublicFrontendSettingSpecsExposeUpdatedLoginDefaults(t *testing.T) {
	descSpec, ok := LookupPublicFrontendSettingSpec(PublicFrontendSettingKeyAuthPageDesc)
	if !ok {
		t.Fatal("expected login page description spec to be present")
	}
	if descSpec.DefaultValue != "面向业务演进，提供开箱即用的管理入口与灵活可插拔的扩展机制" {
		t.Fatalf("unexpected login page description default: %q", descSpec.DefaultValue)
	}

	layoutSpec, ok := LookupPublicFrontendSettingSpec(PublicFrontendSettingKeyAuthLoginPanelLayout)
	if !ok {
		t.Fatal("expected login panel layout spec to be present")
	}
	if layoutSpec.DefaultValue != string(PublicFrontendAuthPanelLayoutRight) {
		t.Fatalf("unexpected login panel layout default: %q", layoutSpec.DefaultValue)
	}

	avatarSpec, ok := LookupPublicFrontendSettingSpec(PublicFrontendSettingKeyUserDefaultAvatar)
	if !ok {
		t.Fatal("expected default avatar spec to be present")
	}
	if avatarSpec.DefaultValue != "/avatar.webp" {
		t.Fatalf("unexpected default avatar value: %q", avatarSpec.DefaultValue)
	}
}

// TestIsProtectedConfigParamRecognizesRuntimeAndFrontendKeys verifies both
// protected-key families are visible through one helper.
func TestIsProtectedConfigParamRecognizesRuntimeAndFrontendKeys(t *testing.T) {
	if !IsProtectedConfigParam(RuntimeParamKeyJWTExpire) {
		t.Fatal("expected runtime param key to be protected")
	}
	if !IsProtectedConfigParam(PublicFrontendSettingKeyAppName) {
		t.Fatal("expected public frontend key to be protected")
	}
	if IsProtectedConfigParam("sys.unknown.key") {
		t.Fatal("expected unknown key not to be protected")
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

// TestProtectedConfigHelpersPreferOverridesAndFallbackDefaults verifies the
// helper readers trim overrides and fall back to built-in defaults.
func TestProtectedConfigHelpersPreferOverridesAndFallbackDefaults(t *testing.T) {
	withCachedRuntimeParamValue(t, PublicFrontendSettingKeyAppName, " LinaPro Custom ")
	svc := New().(*serviceImpl)

	if value := svc.getProtectedConfigValueOrDefault(context.Background(), PublicFrontendSettingKeyAppName); value != "LinaPro Custom" {
		t.Fatalf("expected trimmed protected override value, got %q", value)
	}
	if value := svc.getProtectedConfigValueOrDefault(context.Background(), RuntimeParamKeyJWTExpire); value != runtimeParamSpecByKey[RuntimeParamKeyJWTExpire].DefaultValue {
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

// TestResolveCurrentSystemTimezoneUsesEnvironment verifies a valid TZ
// environment variable is exposed directly to the frontend.
func TestResolveCurrentSystemTimezoneUsesEnvironment(t *testing.T) {
	t.Setenv("TZ", "Asia/Tokyo")
	oldLocal := time.Local
	time.Local = time.UTC
	t.Cleanup(func() {
		time.Local = oldLocal
	})

	if timezone := resolveCurrentSystemTimezone(); timezone != "Asia/Tokyo" {
		t.Fatalf("expected timezone from TZ environment, got %q", timezone)
	}
}

// TestResolveCurrentSystemTimezoneFallsBackToProcessLocation verifies the
// process location is used when TZ is invalid.
func TestResolveCurrentSystemTimezoneFallsBackToProcessLocation(t *testing.T) {
	t.Setenv("TZ", "Invalid/Timezone")
	oldLocal := time.Local
	time.Local = time.UTC
	t.Cleanup(func() {
		time.Local = oldLocal
	})

	if timezone := resolveCurrentSystemTimezone(); timezone != "UTC" {
		t.Fatalf("expected timezone fallback to process location UTC, got %q", timezone)
	}
}

// TestResolveCurrentSystemTimezoneUsesProjectDefault verifies the helper uses
// the project default when both environment and process location are local.
func TestResolveCurrentSystemTimezoneUsesProjectDefault(t *testing.T) {
	t.Setenv("TZ", "")
	oldLocal := time.Local
	time.Local = time.FixedZone("Local", 0)
	t.Cleanup(func() {
		time.Local = oldLocal
	})

	if timezone := resolveCurrentSystemTimezone(); timezone != "Asia/Shanghai" {
		t.Fatalf("expected project default timezone Asia/Shanghai, got %q", timezone)
	}
}
