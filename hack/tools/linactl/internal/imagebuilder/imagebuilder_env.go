// This file prints normalized build settings for make recipes.

package imagebuilder

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

// printBuildEnv emits normalized build config as shell-safe assignments for make recipes.
func printBuildEnv(stdout io.Writer, cfg buildConfig) {
	multiPlatform := "0"
	if cfg.MultiPlatform() {
		multiPlatform = "1"
	}
	fmt.Fprintf(stdout, "BUILD_PLATFORM=%s\n", shellQuote(cfg.Platform))
	fmt.Fprintf(stdout, "BUILD_PLATFORMS=%s\n", shellQuote(joinPlatformSpace(cfg.Targets)))
	fmt.Fprintf(stdout, "BUILD_PLATFORM_COUNT=%d\n", len(cfg.Targets))
	fmt.Fprintf(stdout, "BUILD_MULTI_PLATFORM=%s\n", shellQuote(multiPlatform))
	fmt.Fprintf(stdout, "BUILD_CGO_ENABLED=%s\n", shellQuote(cgoEnabledValue(cfg.CGOEnabled)))
	fmt.Fprintf(stdout, "BUILD_OUTPUT_DIR=%s\n", shellQuote(filepath.ToSlash(cfg.OutputDir)))
	fmt.Fprintf(stdout, "BUILD_BINARY_NAME=%s\n", shellQuote(cfg.BinaryName))
	fmt.Fprintf(stdout, "BUILD_BINARY_PATH=%s\n", shellQuote(filepath.ToSlash(defaultBuildBinaryPath(cfg))))
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
