// This file verifies image build platform normalization and release safety
// checks without invoking Docker.

package main

import "testing"

// TestNormalizeBuildConfigSinglePlatform verifies os/arch overrides keep the
// standard single-binary output contract.
func TestNormalizeBuildConfigSinglePlatform(t *testing.T) {
	cfg := defaultBuildConfig()
	cfg.Arch = "arm64"

	if err := normalizeBuildConfig(&cfg, map[string]bool{"arch": true}); err != nil {
		t.Fatalf("normalizeBuildConfig returned error: %v", err)
	}

	if cfg.OS != "linux" {
		t.Fatalf("OS = %q, want linux", cfg.OS)
	}
	if cfg.Arch != "arm64" {
		t.Fatalf("Arch = %q, want arm64", cfg.Arch)
	}
	if cfg.Platform != "linux/arm64" {
		t.Fatalf("Platform = %q, want linux/arm64", cfg.Platform)
	}
	if cfg.MultiPlatform() {
		t.Fatalf("MultiPlatform = true, want false")
	}
	path := buildOutputBinaryRelPath(cfg, cfg.Targets[0])
	if path != "temp/output/lina" {
		t.Fatalf("single-platform binary path = %q, want temp/output/lina", path)
	}
}

// TestNormalizeBuildConfigMultiPlatform verifies platform lists become stable
// per-platform binary output directories.
func TestNormalizeBuildConfigMultiPlatform(t *testing.T) {
	cfg := defaultBuildConfig()
	cfg.Platform = "linux/amd64,linux/arm64"

	if err := normalizeBuildConfig(&cfg, map[string]bool{"platform": true}); err != nil {
		t.Fatalf("normalizeBuildConfig returned error: %v", err)
	}

	if !cfg.MultiPlatform() {
		t.Fatalf("MultiPlatform = false, want true")
	}
	if cfg.Platform != "linux/amd64,linux/arm64" {
		t.Fatalf("Platform = %q, want linux/amd64,linux/arm64", cfg.Platform)
	}
	if cfg.OS != "linux" || cfg.Arch != "amd64" {
		t.Fatalf("primary target = %s/%s, want linux/amd64", cfg.OS, cfg.Arch)
	}
	if got := buildOutputBinaryRelPath(cfg, cfg.Targets[0]); got != "temp/output/linux_amd64/lina" {
		t.Fatalf("amd64 binary path = %q, want temp/output/linux_amd64/lina", got)
	}
	if got := buildOutputBinaryRelPath(cfg, cfg.Targets[1]); got != "temp/output/linux_arm64/lina" {
		t.Fatalf("arm64 binary path = %q, want temp/output/linux_arm64/lina", got)
	}
}

// TestNormalizeBuildConfigRejectsMixedOSArchWithMultiPlatform verifies that
// explicit os/arch overrides cannot conflict with a platform matrix.
func TestNormalizeBuildConfigRejectsMixedOSArchWithMultiPlatform(t *testing.T) {
	cfg := defaultBuildConfig()
	cfg.Platform = "linux/amd64,linux/arm64"
	cfg.Arch = "amd64"

	err := normalizeBuildConfig(&cfg, map[string]bool{"platform": true, "arch": true})
	if err == nil {
		t.Fatalf("normalizeBuildConfig returned nil, want error")
	}
}

// TestValidateImageBuildRequestRequiresPushForMultiPlatform verifies buildx
// multi-platform publishing is rejected unless push is enabled.
func TestValidateImageBuildRequestRequiresPushForMultiPlatform(t *testing.T) {
	cfg := defaultBuildConfig()
	cfg.Platform = "linux/amd64,linux/arm64"
	if err := normalizeBuildConfig(&cfg, map[string]bool{"platform": true}); err != nil {
		t.Fatalf("normalizeBuildConfig returned error: %v", err)
	}

	image := defaultImageConfig()
	image.Push = false
	if err := validateImageBuildRequest(image, cfg); err == nil {
		t.Fatalf("validateImageBuildRequest returned nil, want error")
	}

	image.Push = true
	image.Registry = "ghcr.io/linaproai"
	if err := validateImageBuildRequest(image, cfg); err != nil {
		t.Fatalf("validateImageBuildRequest returned error: %v", err)
	}
}
