// This file prints normalized build settings for make recipes.

package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

// printBuildEnv emits normalized build config as shell-safe assignments for make recipes.
func printBuildEnv(cfg buildConfig) {
	multiPlatform := "0"
	if cfg.MultiPlatform() {
		multiPlatform = "1"
	}
	fmt.Printf("BUILD_PLATFORM=%s\n", shellQuote(cfg.Platform))
	fmt.Printf("BUILD_PLATFORMS=%s\n", shellQuote(joinPlatformSpace(cfg.Targets)))
	fmt.Printf("BUILD_PLATFORM_COUNT=%d\n", len(cfg.Targets))
	fmt.Printf("BUILD_MULTI_PLATFORM=%s\n", shellQuote(multiPlatform))
	fmt.Printf("BUILD_CGO_ENABLED=%s\n", shellQuote(cgoEnabledValue(cfg.CGOEnabled)))
	fmt.Printf("BUILD_OUTPUT_DIR=%s\n", shellQuote(filepath.ToSlash(cfg.OutputDir)))
	fmt.Printf("BUILD_BINARY_NAME=%s\n", shellQuote(cfg.BinaryName))
	fmt.Printf("BUILD_BINARY_PATH=%s\n", shellQuote(filepath.ToSlash(defaultBuildBinaryPath(cfg))))
}

// shellQuote returns a POSIX shell single-quoted literal.
func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

// cgoEnabledValue converts a boolean into the value expected by CGO_ENABLED.
func cgoEnabledValue(enabled bool) string {
	if enabled {
		return "1"
	}
	return "0"
}
